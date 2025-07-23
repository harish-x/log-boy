const { app } = require("@azure/functions");
const pg = require("../config/postgresql");
const esClient = require("../config/elsaticsearch")
const { DataLakeServiceClient } = require("@azure/storage-file-datalake");
const fs = require("fs");
const path = require("path");


app.timer("backup-cron", {
  schedule: "0 30 6 1 * *",
  handler: async (myTimer, context) => {
    await pg.connect();
    let res = await pg.query("SELECT name, active, retention_period, created_at FROM projects");

    if (res.length > 0) {
      const connectionString = process.env.BlobStorageAccount;
      const serviceClient = DataLakeServiceClient.fromConnectionString(connectionString);

      for (let i = 0; i < res.length; i++) {
        const element = res[i];
        var { name, active, retention_period, created_at } = element;

        const createddate = new Date(created_at);
        const currentDate = new Date(); // current date

        retention_period = parseInt(retention_period.split(" ")[0]);

        const yearDiff = currentDate.getFullYear() - createddate.getFullYear();
        const monthDiff = currentDate.getMonth() - createddate.getMonth();

        const totalMonthsPassed = yearDiff * 12 + monthDiff;

        context.warn(totalMonthsPassed >= retention_period && active);
        context.warn(totalMonthsPassed, " and " + retention_period, " and " + active);

        if (totalMonthsPassed >= retention_period && active) {
          try {
            const isValidProjectName = /^[a-zA-Z0-9_]+$/.test(name);
            if (!isValidProjectName) {
              context.error(`Invalid project name: ${name}. Skipping.`);
              continue;
            }

            // paths
            const backupUploadDirectory = "./backupUpload";
            const localLogsFileName = `${name}_logs_temp_archive.log`;
            const localMetricsFileName = `${name}_metrics_temp_archive.log`;
            const localLogsFilePath = path.join(backupUploadDirectory, localLogsFileName);
            const localMetricsFilePath = path.join(backupUploadDirectory, localMetricsFileName);

            // date range for the logs and metrics
            const archiveStartDate = new Date(currentDate);
            archiveStartDate.setMonth(archiveStartDate.getMonth() - 1); // Go back one month
            archiveStartDate.setDate(1); // Set to the 1st day of that month
            archiveStartDate.setHours(0, 0, 0, 0); // Set to start of the day

            const archiveEndDate = new Date(currentDate);
            archiveEndDate.setDate(1); // Set to the 1st day of the current month
            archiveEndDate.setHours(0, 0, 0, 0); // Set to start of the day

            // Convert to epoch milliseconds for Elasticsearch queries
            const archiveStartEpoch = archiveStartDate.getTime();
            const archiveEndEpoch = archiveEndDate.getTime();
            fs.mkdirSync(backupUploadDirectory, { recursive: true });

            const logsIndexName = `logs-${name}`;

            try {
              const logsIndexExists = await esClient.indices.exists({
                index: logsIndexName,
              });

              let logs = [];
              if (logsIndexExists.body) {
                // Fetch logs
                const logsResponse = await esClient.search({
                  index: logsIndexName,
                  scroll: "1m",
                  size: 1000,
                  body: {
                    query: {
                      bool: {
                        filter: [
                          {
                            range: {
                              timestamp: {
                                gte: archiveStartEpoch,
                                lt: archiveEndEpoch,
                              },
                            },
                          },
                        ],
                      },
                    },
                    sort: [{ timestamp: { order: "asc" } }],
                  },
                });

                logs = logsResponse.body.hits.hits.map((hit) => hit._source);
                let scrollId = logsResponse.body._scroll_id;

                while (logsResponse.body.hits.hits.length > 0) {
                  const scrollResponse = await esClient.scroll({
                    scroll_id: scrollId,
                    scroll: "1m",
                  });

                  if (scrollResponse.body.hits.hits.length === 0) {
                    break;
                  }

                  logs = logs.concat(scrollResponse.body.hits.hits.map((hit) => hit._source));
                  scrollId = scrollResponse.body._scroll_id;
                }

                // Clear scroll
                if (scrollId) {
                  await esClient.clearScroll({ scroll_id: scrollId });
                }
              }

              if (logs.length > 0) {
                context.log(`Found ${logs.length} logs to archive for project ${name}.`);

                // Write Logs to a Local File
                const logsNdjsonContent = logs.map((log) => JSON.stringify(log)).join("\n");

                fs.writeFileSync(localLogsFilePath, logsNdjsonContent);
                context.log(`Logs written to local file: ${localLogsFilePath}`);
              } else {
                context.warn(
                  `No logs found in ${logsIndexName} for the period from ${archiveStartDate.toISOString()} to ${archiveEndDate.toISOString()}.`
                );
              }
            } catch (logsError) {
              context.error(`Error fetching logs for project ${name}:`, logsError);
            }
            let allMetrics = [];

            const metricsIndices = [];
            const currentDateIterator = new Date(archiveStartDate);

            while (currentDateIterator < archiveEndDate) {
              const dateStr = currentDateIterator.toISOString().split("T")[0];
              metricsIndices.push(`metrics${name}-${dateStr}`);
              currentDateIterator.setDate(currentDateIterator.getDate() + 1);
            }

            try {
              // fetch data
              for (const metricsIndex of metricsIndices) {
                try {
                  const metricsIndexExists = await esClient.indices.exists({
                    index: metricsIndex,
                  });

                  if (metricsIndexExists.body) {
                    const metricsResponse = await esClient.search({
                      index: metricsIndex,
                      scroll: "1m",
                      size: 1000,
                      body: {
                        query: {
                          bool: {
                            filter: [
                              {
                                range: {
                                  timestamp: {
                                    gte: archiveStartEpoch,
                                    lt: archiveEndEpoch,
                                  },
                                },
                              },
                            ],
                          },
                        },
                        sort: [{ timestamp: { order: "asc" } }],
                      },
                    });

                    let indexMetrics = metricsResponse.body.hits.hits.map((hit) => hit._source);
                    let scrollId = metricsResponse.body._scroll_id;

                    while (metricsResponse.body.hits.hits.length > 0) {
                      const scrollResponse = await esClient.scroll({
                        scroll_id: scrollId,
                        scroll: "1m",
                      });

                      if (scrollResponse.body.hits.hits.length === 0) {
                        break;
                      }

                      indexMetrics = indexMetrics.concat(scrollResponse.body.hits.hits.map((hit) => hit._source));
                      scrollId = scrollResponse.body._scroll_id;
                    }

                    // Clear scroll
                    if (scrollId) {
                      await esClient.clearScroll({ scroll_id: scrollId });
                    }

                    allMetrics = allMetrics.concat(indexMetrics);
                  }
                } catch (indexError) {
                  context.warn(`Could not fetch from metrics index ${metricsIndex}:`, indexError.message);
                }
              }

              if (allMetrics.length > 0) {
                context.log(`Found ${allMetrics.length} metrics to archive for project ${name}.`);

                const metricsNdjsonContent = allMetrics.map((metric) => JSON.stringify(metric)).join("\n");

                fs.writeFileSync(localMetricsFilePath, metricsNdjsonContent);
                context.log(`Metrics written to local file: ${localMetricsFilePath}`);
              } else {
                context.warn(
                  `No metrics found for project ${name} for the period from ${archiveStartDate.toISOString()} to ${archiveEndDate.toISOString()}.`
                );
              }
            } catch (metricsError) {
              context.error(`Error fetching metrics for project ${name}:`, metricsError);
            }

            // Generate the Azure Blob Filename
            const startMonthName = archiveStartDate.toLocaleString("en-US", { month: "short" });
            const startYear = archiveStartDate.getFullYear();

            const endMonthName = new Date(archiveEndDate.getFullYear(), archiveEndDate.getMonth() - 1, 1).toLocaleString("en-US", { month: "short" }); // Get month *before* archiveEndDate
            const endYear = new Date(archiveEndDate.getFullYear(), archiveEndDate.getMonth() - 1, 1).getFullYear();

            let remoteFileNameBase;
            if (retention_period > 1) {
              remoteFileNameBase = `${startMonthName} ${startYear} to ${endMonthName} ${endYear}`;
            } else {
              remoteFileNameBase = `${startMonthName} ${startYear}`;
            }

            // Upload Files to Azure Data Lake
            const fileSystemName = "backuplogs";
            const remoteDirectoryName = name;

            const fileSystemClient = serviceClient.getFileSystemClient(fileSystemName);
            await fileSystemClient.createIfNotExists();

            const directoryClient = fileSystemClient.getDirectoryClient(remoteDirectoryName);
            await directoryClient.createIfNotExists();

            // Upload Logs if file exists
            if (fs.existsSync(localLogsFilePath)) {
              const logsRemoteFileName = `${remoteFileNameBase}_logs.log`;
              const logsFileClient = directoryClient.getFileClient(logsRemoteFileName);
              const logsFileContents = fs.readFileSync(localLogsFilePath);

              await logsFileClient.upload(logsFileContents, {
                metadata: { project: name, date: remoteFileNameBase, type: "logs" },
              });

              context.log(`Successfully uploaded logs to Data Lake at: ${remoteDirectoryName}/${logsRemoteFileName}`);
            }

            // Upload Metrics if file exists
            if (fs.existsSync(localMetricsFilePath)) {
              const metricsRemoteFileName = `${remoteFileNameBase}_metrics.log`;
              const metricsFileClient = directoryClient.getFileClient(metricsRemoteFileName);
              const metricsFileContents = fs.readFileSync(localMetricsFilePath);

              await metricsFileClient.upload(metricsFileContents, {
                metadata: { project: name, date: remoteFileNameBase, type: "metrics" },
              });

              context.log(`Successfully uploaded metrics to Data Lake at: ${remoteDirectoryName}/${metricsRemoteFileName}`);
            }

            // Cleanup Local Files
            if (fs.existsSync(localLogsFilePath)) {
              fs.unlinkSync(localLogsFilePath);
              context.log(`Cleaned up local logs file: ${localLogsFilePath}`);
            }

            if (fs.existsSync(localMetricsFilePath)) {
              fs.unlinkSync(localMetricsFilePath);
              context.log(`Cleaned up local metrics file: ${localMetricsFilePath}`);
            }

            // Delete Archived Data from Elasticsearch
            try {
              // Delete logs
              if (logs.length > 0) {
                await esClient.deleteByQuery({
                  index: logsIndexName,
                  body: {
                    query: {
                      bool: {
                        filter: [
                          {
                            range: {
                              timestamp: {
                                gte: archiveStartEpoch,
                                lt: archiveEndEpoch,
                              },
                            },
                          },
                        ],
                      },
                    },
                  },
                });
                context.log(`Deleted archived logs from Elasticsearch index ${logsIndexName} for project ${name}.`);
              }

              // Delete metrics
              if (allMetrics.length > 0) {
                for (const metricsIndex of metricsIndices) {
                  try {
                    const indexExists = await esClient.indices.exists({ index: metricsIndex });
                    if (indexExists.body) {
                      await esClient.deleteByQuery({
                        index: metricsIndex,
                        body: {
                          query: {
                            bool: {
                              filter: [
                                {
                                  range: {
                                    timestamp: {
                                      gte: archiveStartEpoch,
                                      lt: archiveEndEpoch,
                                    },
                                  },
                                },
                              ],
                            },
                          },
                        },
                      });
                      context.log(`Deleted archived metrics from Elasticsearch index ${metricsIndex} for project ${name}.`);
                    }
                  } catch (deleteError) {
                    context.warn(`Could not delete from metrics index ${metricsIndex}:`, deleteError.message);
                  }
                }
              }
            } catch (deleteError) {
              context.error(`Error deleting archived data from Elasticsearch for project ${name}:`, deleteError);
            }
          } catch (error) {
            context.error(`An error occurred while archiving data for project ${name}:`, error);
          }
        }
      }
    }
  },
});

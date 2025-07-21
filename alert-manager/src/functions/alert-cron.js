const { app } = require("@azure/functions");
const pg = require("../config/pg");
const esclient = require("../config/elasticsearch");
const { client, multiExecAsync } = require("../config/redis");
const {
  formatAlertMethods,
  operators,
  timewindowFormater,
  calculatePriority,
  createAlertCacheKey,
  groupAlertsByType,
  selectHighestPriorityAlerts,
} = require("../utils/helper");

const ALERT_CACHE_PREFIX = "alert_cache:";
const ALERT_PUBSUB_CHANNEL = process.env.REDIS_ALERT_CHANNEL || "alerts";
const ALERT_COOLDOWN_PERIOD = parseInt(process.env.ALERT_COOLDOWN_SECONDS || "300");

app.timer("alert-cron", {
  schedule: "0 */1 * * * *", // Every minute
  handler: async (myTimer, context) => {
    context.log("Alert cron job started");

    try {
      await pg.connect();
      const projects = await pg.query("SELECT id, name, active, active_monitoring FROM projects");

      if (projects.rows.length === 0) {
        context.log("No projects found");
        return;
      }

      for (const project of projects.rows) {
        if (!project.active || !project.active_monitoring) {
          continue;
        }

        const { name: projectName } = project;
        context.log(`Processing project: ${projectName}`);

        // Get alerts
        const alerts = await pg.query("SELECT * FROM alerts WHERE project_name = $1 ORDER BY threshold DESC", [projectName]);

        if (alerts.rows.length === 0) {
          continue;
        }

        const triggeredAlerts = await processAlerts(alerts.rows, context);

        // Group alerts by metric/log field
        const groupedAlerts = groupAlertsByType(triggeredAlerts);

        // Send only the highest priority alert for each group
        const alertsToSend = selectHighestPriorityAlerts(groupedAlerts);

        // Filter out recently sent alerts using Redis
        const newAlerts = await filterRecentAlertsWithRedis(alertsToSend, context);

        await publishAlertsToRedis(newAlerts, context);
      }
    } catch (error) {
      context.error("Alert cron job failed:", error);
    }
  },
});

async function processAlerts(alerts, context) {
  const triggeredAlerts = [];

  for (const alert of alerts) {
    try {
      let isTriggered = false;
      let alertData = null;

      if (alert.rule_type === "metric_avg") {
        const result = await processMetricAlert(alert);
        if (result.triggered) {
          isTriggered = true;
          alertData = result.data;
        }
      } else if (alert.rule_type === "log_count") {
        const result = await processLogAlert(alert);
        if (result.triggered) {
          isTriggered = true;
          alertData = result.data;
        }
      } else if (alert.rule_type === "event_count") {
        const result = await processEventAlert(alert);
        if (result.triggered) {
          isTriggered = true;
          alertData = result.data;
        }
      }

      if (isTriggered && alertData) {
        const alertmethods = await pg.query("SELECT method, value FROM alert_methods WHERE alert_id = $1", [alert.id]);

        if (alertmethods.rows.length === 0) {
          context.warn(`No alert methods found for alert ID: ${alert.id}`);
          continue;
        }

        const methods = formatAlertMethods(alertmethods.rows);

        triggeredAlerts.push({
          ...alertData,
          id: alert.id,
          project_name: alert.project_name,
          operator: alert.operator,
          threshold: alert.threshold,
          time_window: alert.time_window,
          rule_type: alert.rule_type,
          methods,
          timestamp: new Date().toISOString(),
          priority: calculatePriority(alert),
        });
      }
    } catch (error) {
      context.error(`Error processing alert ${alert.id}:`, error);
    }
  }

  return triggeredAlerts;
}

async function processMetricAlert(alert) {
  const timeWindow = timewindowFormater(alert.time_window);

  if (alert.metric_name === "cpu_usage") {
    const query = {
      query: {
        bool: {
          filter: [
            { term: { serviceName: alert.project_name } },
            {
              range: {
                "cpuUsage.timestamp": {
                  gte: timeWindow,
                  lte: "now",
                },
              },
            },
          ],
        },
      },
      aggs: {
        avg_cpu: {
          avg: {
            field: "cpuUsage.average",
          },
        },
      },
    };

    const res = await esclient.search({
      index: `m-${alert.project_name}-*`,
      body: query,
    });

    const avgCpu = res?.aggregations?.avg_cpu?.value;

    if (avgCpu == null) {
      return { triggered: false };
    }

    if (operators[alert.operator]?.(avgCpu, alert.threshold)) {
      return {
        triggered: true,
        data: {
          metric_name: alert.metric_name,
          current_value: parseFloat(avgCpu.toFixed(2)),
        },
      };
    }
  } else if (alert.metric_name === "memory_usage") {
    const query = {
      query: {
        bool: {
          filter: [
            { term: { serviceName: alert.project_name } },
            {
              range: {
                "memoryUsage.timestamp": {
                  gte: timeWindow,
                  lte: "now",
                },
              },
            },
          ],
        },
      },
      aggs: {
        avg_memory: {
          avg: {
            field: "memoryUsage.memoryUsagePercentage",
          },
        },
      },
    };

    const res = await esclient.search({
      index: `m-${alert.project_name}-*`,
      body: query,
    });

    const avgMemory = res?.aggregations?.avg_memory?.value;

    if (avgMemory == null) {
      return { triggered: false };
    }

    if (operators[alert.operator]?.(avgMemory, alert.threshold)) {
      return {
        triggered: true,
        data: {
          metric_name: alert.metric_name,
          current_value: parseFloat(avgMemory.toFixed(2)),
        },
      };
    }
  }

  return { triggered: false };
}

async function processLogAlert(alert) {
  const timeWindow = timewindowFormater(alert.time_window);

  if (alert.log_field === "level") {
    const query = {
      size: 0,
      query: {
        bool: {
          filter: [
            { term: { serviceName: alert.project_name } },
            {
              range: {
                timestamp: {
                  gte: timeWindow,
                  lte: "now",
                },
              },
            },
          ],
        },
      },
      aggs: {
        logs_filter_count: {
          filter: {
            term: {
              level: alert.log_field_value,
            },
          },
        },
      },
    };

    const res = await esclient.search({
      index: `logs-${alert.project_name}`,
      body: query,
    });

    const errorCount = res?.aggregations?.logs_filter_count?.doc_count || 0;
    const totalCount = res?.hits?.total?.value || 0;

    if (totalCount === 0) {
      return { triggered: false };
    }

    const percentage = (errorCount / totalCount) * 100;

    if (operators[alert.operator]?.(percentage, alert.threshold)) {
      return {
        triggered: true,
        data: {
          log_field: alert.log_field,
          log_field_value: alert.log_field_value,
          current_value: parseFloat(percentage.toFixed(2)),
          error_count: errorCount,
          total_count: totalCount,
        },
      };
    }
  } else if (alert.log_field === "status_code") {
    const statusType = alert.log_field_value;
    let filterScript = "";

    if (statusType === "4xx") {
      filterScript = `
        if (doc['responseStatus.keyword'].size() == 0) return false;
        try {
          def code = Integer.parseInt(doc['responseStatus.keyword'].value);
          return code >= 400 && code < 500;
        } catch (Exception e) {
          return false;
        }
      `;
    } else if (statusType === "5xx") {
      filterScript = `
        if (doc['responseStatus.keyword'].size() == 0) return false;
        try {
          def code = Integer.parseInt(doc['responseStatus.keyword'].value);
          return code >= 500;
        } catch (Exception e) {
          return false;
        }
      `;
    } else {
      return { triggered: false };
    }

    const query = {
      size: 0,
      query: {
        bool: {
          filter: [
            { term: { serviceName: alert.project_name } },
            {
              range: {
                timestamp: {
                  gte: timeWindow,
                  lte: "now",
                },
              },
            },
          ],
        },
      },
      aggs: {
        logs_filter_count: {
          filter: {
            script: {
              script: {
                lang: "painless",
                source: filterScript,
              },
            },
          },
        },
      },
    };

    const response = await esclient.search({
      index: `logs-${alert.project_name}`,
      body: query,
    });

    const totalCount = response?.hits?.total?.value || 0;
    const errorCount = response?.aggregations?.logs_filter_count?.doc_count || 0;

    if (totalCount === 0) {
      return { triggered: false };
    }

    const percentage = (errorCount / totalCount) * 100;

    if (operators[alert.operator]?.(percentage, alert.threshold)) {
      return {
        triggered: true,
        data: {
          log_field: alert.log_field,
          log_field_value: alert.log_field_value,
          current_value: parseFloat(percentage.toFixed(2)),
          error_count: errorCount,
          total_count: totalCount,
        },
      };
    }
  } else if (alert.log_field === "ip_address") {
    const query = {
      size: 0,
      query: {
        bool: {
          filter: [
            { term: { serviceName: alert.project_name } },
            {
              range: {
                timestamp: {
                  gte: timeWindow,
                  lte: "now",
                },
              },
            },
          ],
        },
      },
      aggs: {
        ip_distribution: {
          terms: {
            field: "ipAddress.keyword",
            size: 10,
          },
        },
      },
    };

    const res = await esclient.search({
      index: `logs-${alert.project_name}`,
      body: query,
    });

    const buckets = res?.aggregations?.ip_distribution?.buckets || [];

    if (buckets.length === 0) {
      return { triggered: false };
    }

    const total = buckets.reduce((sum, b) => sum + b.doc_count, 0);

    if (total === 0) {
      return { triggered: false };
    }

    const ipPercentages = buckets.map((bucket) => ({
      ip: bucket.key,
      count: bucket.doc_count,
      percentage: parseFloat(((bucket.doc_count / total) * 100).toFixed(2)),
    }));

    const triggeredIp = ipPercentages.find((p) => operators[alert.operator]?.(p.percentage, alert.threshold));

    if (triggeredIp) {
      return {
        triggered: true,
        data: {
          log_field: alert.log_field,
          log_field_value: alert.log_field_value,
          current_value: triggeredIp.percentage,
          triggered_ip: triggeredIp.ip,
          ip_count: triggeredIp.count,
        },
      };
    }
  }

  return { triggered: false };
}

async function processEventAlert(alert) {
  const timeWindow = timewindowFormater(alert.time_window);

  const query = {
    size: 0,
    query: {
      bool: {
        filter: [
          { term: { serviceName: alert.project_name } },
          {
            range: {
              timestamp: {
                gte: timeWindow,
                lte: "now",
              },
            },
          },
          {
            bool: {
              should: [
                { match_phrase: { message: "server started" } },
                { match_phrase: { message: "database connected" } },
                { match_phrase: { message: "database disconnected" } },
                { match_phrase: { message: "db connected" } },
                { match_phrase: { message: "server shutdown" } },
                { match_phrase: { message: "app shutdown" } },
              ],
              minimum_should_match: 1,
            },
          },
        ],
      },
    },
    aggs: {
      message_counts: {
        terms: {
          field: "message.keyword",
          size: 10,
          order: { _count: "desc" },
        },
      },
    },
  };

  const res = await esclient.search({
    index: `logs-${alert.project_name}`,
    body: query,
  });

  const buckets = res?.aggregations?.message_counts?.buckets || [];

  if (buckets.length === 0) {
    return { triggered: false };
  }

  const total = buckets.reduce((sum, b) => sum + b.doc_count, 0);

  if (total === 0) {
    return { triggered: false };
  }

  const messagePercentages = buckets.map((bucket) => ({
    message: bucket.key,
    count: bucket.doc_count,
    percentage: parseFloat(((bucket.doc_count / total) * 100).toFixed(2)),
  }));

  const triggeredMessage = messagePercentages.find((p) => operators[alert.operator]?.(p.percentage, alert.threshold));

  if (triggeredMessage) {
    return {
      triggered: true,
      data: {
        log_field: alert.log_field,
        log_field_value: alert.log_field_value,
        current_value: triggeredMessage.percentage,
        triggered_message: triggeredMessage.message,
        message_count: triggeredMessage.count,
      },
    };
  }

  return { triggered: false };
}

// Filter out alerts using Redis
async function filterRecentAlertsWithRedis(alerts, context) {
  if (alerts.length === 0) return [];

  const filteredAlerts = [];
  const commands = [];
  const alertKeys = [];

  for (const alert of alerts) {
    const alertKey = createAlertCacheKey(alert);
    const redisKey = `${ALERT_CACHE_PREFIX}${alertKey}`;
    alertKeys.push({ alert, redisKey });
    commands.push(["GET", redisKey]);
  }

  try {
    const results = await multiExecAsync(commands);

    for (let i = 0; i < alertKeys.length; i++) {
      const { alert, redisKey } = alertKeys[i];
      const lastSentTime = results[i];
      const now = Math.floor(Date.now() / 1000);

      if (!lastSentTime || now - parseInt(lastSentTime) > ALERT_COOLDOWN_PERIOD) {
        filteredAlerts.push(alert);
        await client.setEx(redisKey, ALERT_COOLDOWN_PERIOD, now.toString());
      } else {
        context.log(`Alert skipped due to cooldown: ${redisKey}`);
      }
    }
  } catch (error) {
    context.error("Redis error:", error);
    return alerts;
  }

  return filteredAlerts;
}

// Publish alerts to Redis Pub/Sub
async function publishAlertsToRedis(alerts, context) {
  if (alerts.length === 0) {
    return;
  }

  context.log(`Publishing ${alerts.length} alerts to Redis channel: ${ALERT_PUBSUB_CHANNEL}`);

  try {
    const commands = alerts.map((alert) => [
      "PUBLISH",
      ALERT_PUBSUB_CHANNEL,
      JSON.stringify({
        ...alert,
        published_at: new Date().toISOString(),
        source: "alert-cron",
        version: "1.0",
      }),
    ]);

    const results = await multiExecAsync(commands);

    for (let i = 0; i < results.length; i++) {
      const subscribers = results[i];
      const alert = alerts[i];

      context.log(`Alert published for project ${alert.project_name}: ` + `${alert.rule_type} (${subscribers} subscribers)`);
    }
  } catch (error) {
    context.error("Failed to publish alerts to Redis:", error);

    await publishAlertsIndividually(alerts, context);
  }
}

async function publishAlertsIndividually(alerts, context) {
  for (const alert of alerts) {
    try {
      const alertPayload = {
        ...alert,
        published_at: new Date().toISOString(),
        source: "alert-cron",
        version: "1.0",
      };

      const subscribers = await publishAsync(ALERT_PUBSUB_CHANNEL, JSON.stringify(alertPayload));

      context.log(`Alert published (fallback) for project ${alert.project_name}: ` + `${alert.rule_type} (${subscribers} subscribers)`);
    } catch (error) {
      context.error(`Failed to publish individual alert ${alert.id}:`, error);

      await handleFailedAlert(alert, error, context);
    }
  }
}

async function handleFailedAlert(alert, error, context) {
  context.error(`CRITICAL: Alert ${alert.id} failed to publish after retries`, {
    alert,
    error,
  });
}
import React from "react";
import { ScrollArea } from "../ui/scroll-area";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Skeleton } from "../ui/skeleton";

// color based on log level
const logLevelClass = {
  warn: "text-warn",
  info: "text-info",
  error: "text-error",
  http: "text-http",
  verbose: "text-verbose",
  debug: "text-debug",
  silly: "text-silly",
};

const LogLists = ({ logError, isLoading, isFetchingLogs, logsData, clearFilters }) => {
  // truncate project description
  function truncate(message, maxLength = 100) {
    if (typeof message !== "string") return "";
    return message.length > maxLength ? message.slice(0, maxLength) + "..." : message;
  }

  // format date with time
  function formatDatewithTime(data) {
    const date = new Date(data);
    return date.toLocaleString("en-US", {
      day: "numeric",
      month: "long",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }
  // loading skeleton
  if (isLoading || isFetchingLogs) {
    return (
      <div className="px-4 py-2 border border-primary/[0.20] w-full h-[calc(100vh-15rem)]">
        <div className="flex flex-col gap-2">
          {Array.from({ length: 10 }).map((_, index) => (
            <Skeleton key={index} className="h-8 w-full " />
          ))}
        </div>
      </div>
    );
  }
  return (
    <div>
      <ScrollArea className="px-4 py-2 border border-primary/[0.20] w-full h-[calc(100vh-15rem)] ">
        <div>
          <Accordion type="single" collapsible>
            {logError && logError?.data?.message === "No logs found" ? (
              <div className="text-center text-muted-foreground mt-10">
                <p>No logs found for the selected criteria.</p>
                <Button variant="default" size="sm" className="mt-4" onClick={clearFilters}>
                  Clear Filters
                </Button>
              </div>
            ) : (
              !isLoading &&
              !isFetchingLogs &&
              !logError &&
              logsData?.data?.logs.map((log, index) => {
                return (
                  <AccordionItem value={`item-${index}`} key={index}>
                    <AccordionTrigger className={"hover:no-underline"}>
                      <div className={`flex items-center justify-between gap-4 w-full bg-warn}`}>
                        {" "}
                        <div>
                          <span className={`font-semibold uppercase ${logLevelClass[log.level]}`}>{log.level}</span>
                        </div>
                        <div className="flex-1">
                          <span className="text-sm text-muted-foreground truncate whitespace-nowrap clamp-2-lines">{truncate(log.message)}</span>
                        </div>
                        <div>
                          <span className="text-xs text-muted-foreground">{formatDatewithTime(log.timestamp)}</span>
                        </div>
                      </div>
                    </AccordionTrigger>
                    <AccordionContent>
                      <div>
                        <div className="space-y-3 text-sm p-4 bg-muted rounded-md border border-border">
                          <div className="flex justify-between items-center">
                            <div>
                              <p className="font-medium">{log?.serviceName}</p>
                              <p className="text-xs text-muted-foreground">
                                v{log.buildDetails?.appVersion} • Node {log.buildDetails?.nodeVersion}
                              </p>
                            </div>
                            <span className={`font-semibold capitalize ${logLevelClass[log.level]}`}>{log?.level}</span>
                          </div>

                          <div>
                            <p className="text-muted-foreground font-medium">Message:</p>
                            <p className="break-words">{log?.message || "No message provided"}</p>
                          </div>

                          {log?.stack && (
                            <div>
                              <p className="text-muted-foreground font-medium">Stack Trace:</p>
                              <pre className="bg-muted text-xs mt-1 p-2 rounded whitespace-pre-wrap break-words border border-border">
                                {log.stack}
                              </pre>
                            </div>
                          )}

                          <div className="grid grid-cols-2 gap-2">
                            <p>
                              <span className="text-muted-foreground">Request ID:</span> {log?.requestId || "-"}
                            </p>
                            <p>
                              <span className="text-muted-foreground">Method:</span> {log?.requestMethod || "-"}
                            </p>
                            <p>
                              <span className="text-muted-foreground">URL:</span> {log?.requestUrl || "-"}
                            </p>
                            <p>
                              <span className="text-muted-foreground">IP:</span> {log?.ipAddress || "-"}
                            </p>
                          </div>

                          <div>
                            <p className="text-muted-foreground">User Agent:</p>
                            <p className="break-words">{log?.userAgent || "-"}</p>
                          </div>

                          <p className="text-xs text-muted-foreground">Timestamp: {formatDatewithTime(log.timestamp)}</p>
                        </div>
                      </div>
                    </AccordionContent>
                  </AccordionItem>
                );
              })
            )}
          </Accordion>
        </div>
      </ScrollArea>
    </div>
  );
};

export default LogLists;

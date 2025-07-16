import React, { memo, useMemo, useCallback, useState, useEffect, useRef } from "react";
import { ScrollArea } from "../ui/scroll-area";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Skeleton } from "../ui/skeleton";

const logLevelClass = {
  warn: "text-warn",
  info: "text-info",
  error: "text-error",
  http: "text-http",
  verbose: "text-verbose",
  debug: "text-debug",
  silly: "text-silly",
};

const truncate = (message, maxLength = 100) => {
  if (typeof message !== "string") return "";
  return message.length > maxLength ? message.slice(0, maxLength) + "..." : message;
};

const formatDatewithTime = (data) => {
  const date = new Date(data);
  return date.toLocaleString("en-US", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
};

const LogItem = memo(({ log, index }) => {
  const truncatedMessage = useMemo(() => truncate(log.message), [log.message]);
  const formattedTime = useMemo(() => formatDatewithTime(log.timestamp), [log.timestamp]);
  const levelClass = useMemo(() => logLevelClass[log.level] || "text-muted-foreground", [log.level]);

  return (
    <AccordionItem value={`item-${log.id || index}`}>
      <AccordionTrigger className="hover:no-underline">
        <div className="flex items-center justify-between gap-4 w-full">
          <div className="flex-shrink-0">
            <span className={`font-semibold uppercase ${levelClass}`}>{log.level}</span>
          </div>
          <div className="flex-1 min-w-0">
            <span className="text-sm text-muted-foreground truncate whitespace-nowrap block">{truncatedMessage}</span>
          </div>
          <div className="flex-shrink-0">
            <span className="text-xs text-muted-foreground">{formattedTime}</span>
          </div>
        </div>
      </AccordionTrigger>
      <AccordionContent>
        <div className="space-y-3 text-sm p-4 bg-muted rounded-md border border-border">
          <div className="flex justify-between items-center">
            <div>
              <p className="font-medium">{log?.serviceName || "Unknown Service"}</p>
              <p className="text-xs text-muted-foreground">
                v{log.buildDetails?.appVersion || "N/A"} • Node {log.buildDetails?.nodeVersion || "N/A"}
              </p>
            </div>
            <span className={`font-semibold capitalize ${levelClass}`}>{log?.level}</span>
          </div>

          <div>
            <p className="text-muted-foreground font-medium">Message:</p>
            <p className="break-words">{log?.message || "No message provided"}</p>
          </div>

          {log?.stack && (
            <div>
              <p className="text-muted-foreground font-medium">Stack Trace:</p>
              <pre className="bg-muted text-xs mt-1 p-2 rounded whitespace-pre-wrap break-words border border-border overflow-x-auto">
                {log.stack}
              </pre>
            </div>
          )}

          <div className="grid grid-cols-2 gap-2 text-xs">
            <p>
              <span className="text-muted-foreground">Request ID:</span> <span className="font-mono">{log?.requestId || "-"}</span>
            </p>
            <p>
              <span className="text-muted-foreground">Method:</span> <span className="font-mono">{log?.requestMethod || "-"}</span>
            </p>
            <p className="col-span-2">
              <span className="text-muted-foreground">URL:</span> <span className="font-mono break-all">{log?.requestUrl || "-"}</span>
            </p>
            <p className="">
              <span className="text-muted-foreground">Response Time:</span> <span className="font-mono break-all">{ log?.responseTime || "-"}</span>
            </p>
            <p className="">
              <span className="text-muted-foreground">Response Status:</span>{" "}
              <span
                className={`font-mono break-all ${
                  parseInt(log?.responseStatus) < 400 ? "text-green-600" : parseInt(log?.responseStatus) < 500 ? "text-warn" : "text-destructive"
                }`}
              >
                {log?.responseStatus || "-"}
              </span>
            </p>
            <p>
              <span className="text-muted-foreground">IP:</span> <span className={`font-mono`}>{log?.ipAddress || "-"}</span>
            </p>
          </div>

          <div>
            <p className="text-muted-foreground text-xs">User Agent:</p>
            <p className="break-words text-xs font-mono">{log?.userAgent || "-"}</p>
          </div>

          <p className="text-xs text-muted-foreground border-t pt-2">Timestamp: {formattedTime}</p>
        </div>
      </AccordionContent>
    </AccordionItem>
  );
});

LogItem.displayName = "LogItem";

const VirtualizedLogList = memo(({ logs, itemHeight = 60, containerHeight = 400 }) => {
  const [scrollTop, setScrollTop] = useState(0);
  const scrollElementRef = useRef(null);

  const visibleRange = useMemo(() => {
    const start = Math.floor(scrollTop / itemHeight);
    const visibleCount = Math.ceil(containerHeight / itemHeight);
    const end = Math.min(start + visibleCount + 5, logs.length);
    return { start: Math.max(0, start - 5), end };
  }, [scrollTop, itemHeight, containerHeight, logs.length]);

  const handleScroll = useCallback((e) => {
    setScrollTop(e.target.scrollTop);
  }, []);

  const visibleLogs = useMemo(() => {
    return logs.slice(visibleRange.start, visibleRange.end);
  }, [logs, visibleRange]);

  const totalHeight = logs.length * itemHeight;
  const offsetY = visibleRange.start * itemHeight;

  return (
    <div ref={scrollElementRef} onScroll={handleScroll}>
      <div>
        <div style={{ transform: `translateY(${offsetY}px)` }}>
          <Accordion type="single" collapsible>
            {visibleLogs.map((log, index) => (
              <LogItem key={log.id || `${visibleRange.start + index}`} log={log} index={visibleRange.start + index} />
            ))}
          </Accordion>
        </div>
      </div>
    </div>
  );
});

VirtualizedLogList.displayName = "VirtualizedLogList";

const LogLists = memo(({ logError, isLoading, isFetchingLogs, logsData, clearFilters, isLiveMode = false }) => {
  const [isVirtualized, setIsVirtualized] = useState(false);

  const logs = useMemo(() => {
    return logsData?.data?.logs || [];
  }, [logsData]);

  useEffect(() => {
    setIsVirtualized(logs.length > 50 || isLiveMode);
  }, [logs.length, isLiveMode]);

  const loadingSkeleton = useMemo(
    () => (
      <div className="px-4 py-2 border border-primary/[0.20] w-full h-[calc(100vh-15rem)]">
        <div className="flex flex-col gap-2">
          {Array.from({ length: 10 }).map((_, index) => (
            <Skeleton key={index} className="h-8 w-full" />
          ))}
        </div>
      </div>
    ),
    []
  );

  const emptyState = useMemo(
    () => (
      <div className="text-center text-muted-foreground mt-10">
        <p>No logs found for the selected criteria.</p>
        {clearFilters && (
          <Button variant="default" size="sm" className="mt-4" onClick={clearFilters}>
            Clear Filters
          </Button>
        )}
      </div>
    ),
    [clearFilters]
  );

  if (isLoading || isFetchingLogs) {
    return loadingSkeleton;
  }

  if (logError && logError?.data?.message === "No logs found") {
    return <div className="px-4 py-2 border border-primary/[0.20] w-full h-[calc(100vh-15rem)]">{emptyState}</div>;
  }

  if (logs.length === 0) {
    return (
      <div className="px-4 py-2 border border-primary/[0.20] w-full h-[calc(100vh-15rem)]">
        <div className="text-center text-muted-foreground mt-10">
          <p>No logs available.</p>
        </div>
      </div>
    );
  }

  return (
    <div>
      <ScrollArea className="px-4 py-2 border border-primary/[0.20] w-full h-[calc(100vh-15rem)]" id="log-list-scroll-area">
        {isVirtualized ? (
          <VirtualizedLogList logs={logs} />
        ) : (
          <Accordion type="single" collapsible>
            {logs.map((log, index) => (
              <LogItem key={log.id || index} log={log} index={index} />
            ))}
          </Accordion>
        )}
      </ScrollArea>
    </div>
  );
});

LogLists.displayName = "LogLists";

export default LogLists;

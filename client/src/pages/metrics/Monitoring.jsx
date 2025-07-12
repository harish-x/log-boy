import React, { useEffect, useState, useRef, useCallback, useMemo } from "react";
import { useParams } from "react-router-dom";
import { useMsal } from "@azure/msal-react";
import { apiTokenRequest } from "@/authConfig";
import { toast } from "sonner";
import LiveMetrics from "@/components/metrics/LiveMetrics";

const Monitoring = () => {
  const { projectName } = useParams();
  const [metricsData, setMetricsData] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const [lastUpdateTime, setLastUpdateTime] = useState(null);
  const { instance } = useMsal();
  const eventSourceRef = useRef(null);
  const cleanupRef = useRef(null);
  const isConnectingRef = useRef(false);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttempts = useRef(0);


  const memoizedMetrics = useMemo(() => {
    return metricsData;
  }, [metricsData]);

  const updateMetricsData = useCallback((newMetrics) => {
    setMetricsData((prev) => {
      if (prev.length > 0 && JSON.stringify(prev[0]) === JSON.stringify(newMetrics)) {
        return prev;
      }

      const newLogs = [newMetrics, ...prev];
      // Keep only latest 100 metrics
      const trimmedLogs = newLogs.slice(0, 100);

      // Update last update time
      setLastUpdateTime(Date.now());

      return trimmedLogs;
    });
  }, []);


  const updateConnectionStatus = useCallback((status) => {
    setIsConnected(status);
  }, []);


  const clearMetrics = useCallback(() => {
    setMetricsData([]);
    setLastUpdateTime(null);
  }, []);


  const setupSSEConnection = useCallback(async () => {
    // Cleanup existing connections
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    if (cleanupRef.current) {
      cleanupRef.current();
      cleanupRef.current = null;
    }

    try {
      isConnectingRef.current = true;
      const accounts = instance.getAllAccounts();
      if (accounts.length === 0) {
        throw new Error("No accounts found");
      }

      const tokenResponse = await instance.acquireTokenSilent({
        ...apiTokenRequest,
        account: accounts[0],
      });
      const accessToken = tokenResponse.accessToken;
      const eventSource = new EventSource(`http://localhost:8080/api/v1/metrics/${projectName}/stream?bearer=${accessToken}`);
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        toast.success(`Connected to ${projectName} log stream`);
        isConnectingRef.current = false;
        reconnectAttempts.current = 0;
        updateConnectionStatus(true);
      };

      eventSource.onmessage = (event) => {
        // Skip the initial "connected" message
        if (event.data === "connected") {
          return;
        }

        // Skip heartbeat messages
        if (event.data === '{"type":"heartbeat"}') {
          return;
        }

        try {
          const metrics = JSON.parse(event.data);
          updateMetricsData(metrics);
        } catch (parseError) {
          console.error("Error parsing metrics data:", parseError);
        }
      };

      eventSource.onerror = (err) => {
        isConnectingRef.current = false;
        updateConnectionStatus(false);

        if (reconnectAttempts.current > 0) {
          console.warn("SSE connection lost, attempting to reconnect...");
        }
        eventSource.close();

        if (eventSourceRef.current === eventSource) {
          reconnectAttempts.current++;
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current - 1), 30000); // Max 30 seconds

          reconnectTimeoutRef.current = setTimeout(() => {
            if (eventSourceRef.current === eventSource) {
              setupSSEConnection();
            }
          }, delay);
        }
      };

      cleanupRef.current = () => {
        if (eventSource.readyState !== EventSource.CLOSED) {
          eventSource.close();
        }
      };
    } catch (error) {
      isConnectingRef.current = false;
      updateConnectionStatus(false);
      toast.error("Failed to establish connection to log stream");

      // Retry with exponential backoff
      reconnectAttempts.current++;
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current - 1), 30000);

      reconnectTimeoutRef.current = setTimeout(() => {
        setupSSEConnection();
      }, delay);
    }
  }, [instance, projectName, updateMetricsData, updateConnectionStatus]);

  const handleVisibilityChange = useCallback(() => {
    if (document.visibilityState === "visible" && !eventSourceRef.current) {
      setupSSEConnection();
    } else if (document.visibilityState === "hidden" && eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
      updateConnectionStatus(false);
    }
  }, [setupSSEConnection, updateConnectionStatus]);

  useEffect(() => {
    setupSSEConnection();
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      if (cleanupRef.current) {
        cleanupRef.current();
        cleanupRef.current = null;
      }
      isConnectingRef.current = false;
      reconnectAttempts.current = 0;
      updateConnectionStatus(false);
    };
  }, [setupSSEConnection, updateConnectionStatus]);

  useEffect(() => {
    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [handleVisibilityChange]);


  const performanceStats = useMemo(() => {
    if (metricsData.length === 0) return null;

    const avgCpu = metricsData.reduce((sum, metric) => sum + (metric.cpuUsage?.average || 0), 0) / metricsData.length;
    const avgMemory = metricsData.reduce((sum, metric) => sum + (metric.memoryUsage?.memoryUsagePercentage || 0), 0) / metricsData.length;
    const usedmemory = metricsData.reduce((sum, metric) => sum + (metric.memoryUsage?.usedMemory || 0), 0) / metricsData.length;

    return {
      averageCpu: avgCpu.toFixed(2),
      averageMemory: avgMemory.toFixed(2),
      usedMemory: usedmemory.toFixed(2),
      freeMemory: (metricsData.reduce((sum, metric) => sum + (metric.memoryUsage?.freeMemory || 0), 0) / metricsData.length).toFixed(2),
      totalDataPoints: metricsData.length,
    };
  }, [metricsData]);

const liveMetricsProps = useMemo(
    () => ({
      metrics: memoizedMetrics,
      performanceStats,
      onClearMetrics: clearMetrics,
      projectName,
    }),
    [memoizedMetrics, performanceStats, clearMetrics, projectName]
  );


  return (
    <>
      <div className="projects-container bg-background px-2 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] min-h-screen">
        <LiveMetrics {...liveMetricsProps} />
      </div>
    </>
  );
};

export default Monitoring;

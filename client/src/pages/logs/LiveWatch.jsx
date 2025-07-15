import LogLists from "@/components/logs/LogLists";
import { Activity, AlertTriangle } from "lucide-react";
import React, { useEffect, useState, useRef, useCallback, useMemo } from "react";
import { useParams } from "react-router-dom";
import { toast } from "sonner";
import { useLazyGetLogsQuery } from "@/services/LogServices";
import { useGetProjectByNameQuery } from "@/services/ProjectService";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useMsal } from "@azure/msal-react";
import { apiTokenRequest } from "@/authConfig";

const LiveWatch = () => {
  const { projectName } = useParams();
  const [logsData, setLogsData] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const [getLogs] = useLazyGetLogsQuery();
  const { isLoading: isLoadingProject, isError: isErrorProject } = useGetProjectByNameQuery(projectName);
  const { instance } = useMsal();

  // Performance optimizations
  const logBufferRef = useRef([]);
  const batchTimeoutRef = useRef(null);
  const eventSourceRef = useRef(null);
  const cleanupRef = useRef(null);
  const isConnectingRef = useRef(false);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttempts = useRef(0);
  const lastToastRef = useRef(0);


  const MAX_LOGS = 100;
  const BATCH_SIZE = 10; 
  const BATCH_DELAY = 100;
  const TOAST_DEBOUNCE = 5000;

  // Optimized batch processing with debouncing
  const processBatchedLogs = useCallback(() => {
    if (logBufferRef.current.length === 0) return;

    const newLogs = logBufferRef.current.splice(0, BATCH_SIZE);
    const parsedLogs = [];

    for (const logStr of newLogs) {
      try {
        const parsed = JSON.parse(logStr);
        if (!parsed.timestamp) {
          parsed.timestamp = new Date().toISOString();
        }
        parsed.id = parsed.id || `${Date.now()}-${Math.random()}`;
        parsedLogs.push(parsed);
      } catch (error) {
        parsedLogs.push({
          message: logStr,
          level: "info",
          timestamp: new Date().toISOString(),
          id: `${Date.now()}-${Math.random()}`,
        });
      }
    }

    setLogsData((prevLogs) => {
      const combined = [...parsedLogs, ...prevLogs];
      return combined.slice(0, MAX_LOGS);
    });


    if (logBufferRef.current.length > 0) {
      batchTimeoutRef.current = setTimeout(processBatchedLogs, BATCH_DELAY);
    } else {
      batchTimeoutRef.current = null;
    }
  }, []);

  // Debounced toast function
  const showToast = useCallback((message, type = "success") => {
    const now = Date.now();
    if (now - lastToastRef.current > TOAST_DEBOUNCE) {
      toast[type](message);
      lastToastRef.current = now;
    }
  }, []);

  // Load initial logs
  useEffect(() => {
    if (isErrorProject || isLoadingProject) return;

    getLogs({
      project: projectName,
      limit: 10,
      page: 1,
    })
      .unwrap()
      .then((response) => {
        const initialLogs = response.data.logs.map((log, index) => ({
          ...log,
          id: log.id || `initial-${index}`,
        }));
        setLogsData(initialLogs);
      })
      .catch((error) => {
        showToast("Failed to load initial logs", "error");
      });
  }, [projectName, isLoadingProject, isErrorProject, getLogs, showToast]);

  // Optimized SSE connection setup
  const setupSSEConnection = useCallback(async () => {
    if (isErrorProject || isLoadingProject || isConnectingRef.current) {
      return;
    }

    // Clear existing timeouts and connections
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (batchTimeoutRef.current) {
      clearTimeout(batchTimeoutRef.current);
      batchTimeoutRef.current = null;
    }

    // Cleanup existing connection
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
      setIsConnected(false);

      const accounts = instance.getAllAccounts();
      if (accounts.length === 0) {
        throw new Error("No MSAL account found");
      }

      const tokenResponse = await instance.acquireTokenSilent({
        ...apiTokenRequest,
        account: accounts[0],
      });

      const accessToken = tokenResponse.accessToken;
      const eventSource = new EventSource(`http://localhost:8080/api/v1/logs/${projectName}/stream?bearer=${accessToken}`);

      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        setIsConnected(true);
        showToast(`Connected to ${projectName} log stream`);
        isConnectingRef.current = false;
        reconnectAttempts.current = 0;
      };

      eventSource.onmessage = (event) => {
        if (event.data === "connected" || event.data === '{"type":"heartbeat"}') {
          return;
        }
       
        logBufferRef.current.push(event.data);

        if (!batchTimeoutRef.current) {
          batchTimeoutRef.current = setTimeout(processBatchedLogs, BATCH_DELAY);
        }
      };

      eventSource.onerror = (err) => {
        setIsConnected(false);
        isConnectingRef.current = false;

        if (reconnectAttempts.current > 0) {
          console.warn("SSE connection lost, attempting to reconnect...");
        }

        eventSource.close();

        if (eventSourceRef.current === eventSource) {
          reconnectAttempts.current++;
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current - 1), 30000);

          if (reconnectAttempts.current <= 5) {
            // Limit reconnection attempts
            reconnectTimeoutRef.current = setTimeout(() => {
              if (eventSourceRef.current === eventSource) {
                setupSSEConnection();
              }
            }, delay);
          } else {
            showToast("Connection failed after multiple attempts", "error");
          }
        }
      };


      cleanupRef.current = () => {
        if (eventSource.readyState !== EventSource.CLOSED) {
          eventSource.close();
        }
      };
    } catch (error) {
      isConnectingRef.current = false;
      setIsConnected(false);
      showToast("Failed to establish connection to log stream", "error");

      if (reconnectAttempts.current < 5) {
        reconnectAttempts.current++;
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current - 1), 30000);

        reconnectTimeoutRef.current = setTimeout(() => {
          setupSSEConnection();
        }, delay);
      }
    }
  }, [projectName, isLoadingProject, isErrorProject, instance, showToast, processBatchedLogs]);


  useEffect(() => {
    setupSSEConnection();

    return () => {

      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (batchTimeoutRef.current) {
        clearTimeout(batchTimeoutRef.current);
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
      logBufferRef.current = [];
    };
  }, [setupSSEConnection]);


  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        if (!eventSourceRef.current || eventSourceRef.current.readyState === EventSource.CLOSED) {
          setupSSEConnection();
        }
      } else if (document.visibilityState === "hidden") {
        if (eventSourceRef.current && eventSourceRef.current.readyState === EventSource.OPEN) {
          eventSourceRef.current.close();
          eventSourceRef.current = null;
          setIsConnected(false);
        }
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [setupSSEConnection]);


  const logsFormatted = useMemo(
    () => ({
      data: {
        logs: logsData,
      },
    }),
    [logsData]
  );

  if (!isLoadingProject && isErrorProject) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-5rem)]">
        <Card className="max-w-md mx-auto text-center">
          <CardHeader>
            <CardTitle className="text-2xl flex items-center justify-center">
              <AlertTriangle className="w-8 h-8 mr-2 text-destructive" /> Project Not Found
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-muted-foreground">The project "{projectName}" could not be found or loaded.</p>
            <p className="text-muted-foreground mt-2">Please check the project name or try again later.</p>
          </CardContent>
          <CardFooter>
            <Button onClick={() => window.history.back()} variant="outline" className="w-full">
              Go Back
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div className="projects-container bg-background px-2 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] h-[calc(100vh-5rem)]">
      <div className="text-2xl font-bold mt-6 ml-4 flex items-center gap-2">
        <Activity className={isConnected ? "text-green-500" : "text-red-500"} />
        <span>Live Watch {projectName}</span>
        <span className={`text-sm font-normal ${isConnected ? "text-green-500" : "text-red-500"}`}>{isConnected ? "Connected" : "Disconnected"}</span>
      </div>
      <div className="mt-6">
        <LogLists
          logsData={logsFormatted}
          logError={null}
          isLoading={false}
          isFetchingLogs={false}
          clearFilters={null}
          isLiveMode={true}
        />
      </div>
    </div>
  );
};

export default LiveWatch;

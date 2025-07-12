import LogLists from "@/components/logs/LogLists";
import { Activity, AlertTriangle } from "lucide-react";
import React, { useEffect, useState, useRef, useCallback } from "react";
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
  const [getLogs] = useLazyGetLogsQuery();
  const { isLoading: isLoadingProject, isError: isErrorProject } = useGetProjectByNameQuery(projectName);
  const { instance } = useMsal();

  // Use refs to track EventSource and cleanup
  const eventSourceRef = useRef(null);
  const cleanupRef = useRef(null);
  const isConnectingRef = useRef(false);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttempts = useRef(0);

  useEffect(() => {
    if (isErrorProject || isLoadingProject) return;

    getLogs({
      project: projectName,
      limit: 5,
      page: 1,
    })
      .unwrap()
      .then((response) => {
        setLogsData((prevLogs) => {
          if (prevLogs.length <= 0) {
            return response.data.logs.map((log) => JSON.stringify(log));
          }
          return prevLogs;
        });
      })
      .catch((error) => {
        toast.error("Failed to load initial logs");
      });
  }, [projectName, isLoadingProject, isErrorProject, getLogs]);

  // Setup SSE connection with exponential backoff
  const setupSSEConnection = useCallback(async () => {
    if (isErrorProject || isLoadingProject || isConnectingRef.current) {
      return;
    }

    // Clear any existing reconnect timeout
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
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
        toast.success(`Connected to ${projectName} log stream`);
        isConnectingRef.current = false;
        reconnectAttempts.current = 0; // Reset attempts on successful connection
      };

      eventSource.onmessage = (event) => {
        // Skip the initial "connected" message
        if (event.data === "connected") {
          return;
        }
        // Skip heartbeat messages
        if (event.data === "{\"type\":\"heartbeat\"}") {
          return;
        }
        setLogsData((prevLogs) => {
          const newLogs = [event.data, ...prevLogs];
          // Keep only latest 100 logs
          return newLogs.slice(0, 100);
        });
      };

      eventSource.onerror = (err) => {
        isConnectingRef.current = false;

        if (reconnectAttempts.current > 0) {
          console.warn("SSE connection lost, attempting to reconnect...");
        }

        eventSource.close();

        // Exponential backoff reconnection
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

      // Setup cleanup function
      cleanupRef.current = () => {
        if (eventSource.readyState !== EventSource.CLOSED) {
          eventSource.close();
        }
      };
    } catch (error) {
      isConnectingRef.current = false;
      toast.error("Failed to establish connection to log stream");

      // Retry with exponential backoff
      reconnectAttempts.current++;
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.current - 1), 30000);

      reconnectTimeoutRef.current = setTimeout(() => {
        setupSSEConnection();
      }, delay);
    }
  }, [projectName, isLoadingProject, isErrorProject, instance]);

  // Setup SSE connection when component mounts or project changes
  useEffect(() => {
    setupSSEConnection();

    // Cleanup on unmount or project change
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
    };
  }, [setupSSEConnection]);


  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible" && !eventSourceRef.current) {
        setupSSEConnection();
      } else if (document.visibilityState === "hidden" && eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [setupSSEConnection]);

  const logsFormatted = {
    data: {
      logs: logsData.map((log) => {
        try {
          return JSON.parse(log);
        } catch (error) {
          return {
            message: log,
            level: "info",
            timestamp: new Date().toISOString(),
          };
        }
      }),
    },
  };

  if (isLoadingProject || isErrorProject) {
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
    <div className="projects-container bg-background px-2 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] h-[calc(100vh-5rem)] ">
      <div className="text-2xl font-bold mt-6 ml-4 flex items-center gap-2">
        <Activity /> <span className=" ">Live Watch {projectName}</span>
      </div>
      <div className="mt-6">
        <LogLists logsData={logsFormatted} logError={null} isLoading={false} isFetchingLogs={false} clearFilters={null} />
      </div>
    </div>
  );
};

export default LiveWatch;

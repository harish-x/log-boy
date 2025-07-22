import React, { useState, useEffect, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Settings, Bell, AlertTriangle, Activity, Eye, EyeOff, Trash2, Clock } from "lucide-react";
import { Link } from "react-router-dom";
import { useParams } from "react-router-dom";
import { apiTokenRequest } from "@/authConfig";
import { useMsal } from "@azure/msal-react";
import { useLazyGetAlertsQuery } from "@/services/AlertServices";

const Alerts = () => {
  const { projectName } = useParams();
  const { instance } = useMsal();
  const [alerts, setAlerts] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const [showRead, setShowRead] = useState(false);
  const [readAlerts, setReadAlerts] = useState(new Set());
  const eventSourceRef = useRef(null);

  const [getAlerts, { isLoading, isError, error, data: oldAlerts }] = useLazyGetAlertsQuery();

  useEffect(() => {
    if (projectName) {
      getAlerts(projectName);
    }
  }, [projectName, getAlerts]);

  console.log(oldAlerts);
  useEffect(() => {
    if (oldAlerts?.data) {
      setAlerts((prevAlerts) => {
        const apiAlerts = oldAlerts.data.map((alert) => ({
          ...alert,
          source_type: "api",
        }));

        const existingIds = new Set(prevAlerts.map((alert) => alert.id));
        const newApiAlerts = apiAlerts.filter((alert) => !existingIds.has(alert.id));

        return [...prevAlerts, ...newApiAlerts];
      });
    }
  }, [oldAlerts]);

  // Setup SSE connection
  const setupSSE = async () => {
    try {
      const accounts = instance.getAllAccounts();
      if (accounts.length === 0) {
        throw new Error("No MSAL account found");
      }

      const tokenResponse = await instance.acquireTokenSilent({
        ...apiTokenRequest,
        account: accounts[0],
      });

      const accessToken = tokenResponse.accessToken;

      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }

      const eventSource = new EventSource(`http://localhost:8080/api/v1/alerts/${projectName}/stream?bearer=${accessToken}`);

      eventSource.onopen = () => {
        setIsConnected(true);
        console.log("SSE connection opened");
      };

      eventSource.onmessage = (event) => {
        try {
          if (event.data === "connected" || event.data === '{"type":"heartbeat"}') {
            return;
          }
          console.log(event.data);
          const alertData = JSON.parse(event.data);
          const newAlert = {
            ...alertData,
            source_type: "sse",
          };

          setAlerts((prevAlerts) => {
            const exists = prevAlerts.some((alert) => alert.id === newAlert.id);
            if (exists) return prevAlerts;

            return [newAlert, ...prevAlerts];
          });
        } catch (err) {
          console.error("Error parsing SSE alert data:", err);
        }
      };

      eventSource.onerror = (error) => {
        console.error("SSE connection error:", error);
        setIsConnected(false);
      };

      eventSourceRef.current = eventSource;
    } catch (error) {
      console.error("Error setting up SSE:", error);
      setIsConnected(false);
    }
  };

  useEffect(() => {
    if (projectName && instance) {
      setupSSE();
    }

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        setIsConnected(false);
      }
    };
  }, [projectName, instance]);

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleString();
  };

  const getPriorityColor = (priority) => {
    const p = parseInt(priority);
    if (p >= 80) return "text-red-500 bg-red-50 border-red-200";
    if (p >= 60) return "text-orange-500 bg-orange-50 border-orange-200";
    if (p >= 40) return "text-yellow-500 bg-yellow-50 border-yellow-200";
    return "text-blue-500 bg-blue-50 border-blue-200";
  };

  const getAlertIcon = (type) => {
    switch (type) {
      case "metric_avg":
        return <Activity className="h-5 w-5" />;
      case "log_count":
        return <AlertTriangle className="h-5 w-5" />;
      case "event_count":
        return <Bell className="h-5 w-5" />;
      default:
        return <AlertTriangle className="h-5 w-5" />;
    }
  };

  const formatAlertMessage = (alert) => {
    switch (alert.type) {
      case "metric_avg":
        return `${alert.metric_name} is at ${alert.current_value}% (${alert.operator} ${alert.threshold}%) over the past ${alert.time_window}`;
      case "log_count":
        if (alert.log_field === "level") {
          return `${alert.log_field_value} level logs accounted for ${alert.current_value.toFixed(1)}% over the past ${alert.time_window}`;
        }
        if (alert.log_field === "status_code") {
          return `${alert.log_field_value} status codes occurred in more than ${alert.threshold}% of requests over the past ${alert.time_window}`;
        }
        if (alert.log_field === "ip_address") {
          return `${alert.error_count} requests from IP ${alert.log_field_value} out of ${alert.total_count} total (${alert.current_value.toFixed(
            1
          )}%) over the past ${alert.time_window}`;
        }
        return `${alert.error_count} occurrences of ${alert.log_field}: ${alert.log_field_value} out of ${alert.total_count} total over the past ${alert.time_window}`;
      case "event_count":
        return `Event count threshold exceeded: ${alert.current_value} events detected over the past ${alert.time_window}`;
      default:
        return "Alert triggered";
    }
  };

  const markAsRead = (alertId) => {
    setReadAlerts((prev) => new Set([...prev, alertId]));
  };

  const markAsUnread = (alertId) => {
    setReadAlerts((prev) => {
      const newSet = new Set(prev);
      newSet.delete(alertId);
      return newSet;
    });
  };

  const deleteAlert = (alertId) => {
    setAlerts((prev) => prev.filter((alert) => alert.id !== alertId));
    setReadAlerts((prev) => {
      const newSet = new Set(prev);
      newSet.delete(alertId);
      return newSet;
    });
  };

  console.log("======================", alerts);

  const renderAlert = (alert) => {
    const isRead = readAlerts.has(alert.id);
    const isLiveAlert = alert.source_type === "sse";

    return (
      <div
        key={alert.id}
        className={`border-card rounded-lg p-4 mb-3 transition-all duration-200 ${
          isRead ? "bg-gray-50 opacity-75" : "bg-card shadow-sm"
        }  hover:shadow-md`}
      >
        <div className="flex items-start justify-between">
          <div className="flex items-start space-x-3 flex-1">
            <div className="mt-1">{getAlertIcon(alert.type)}</div>
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-2">
                <span className="font-semibold text-accent-foreground">{alert.project_name}</span>
                <span className={`px-2 py-1 rounded text-xs font-medium ${getPriorityColor(alert.priority)}`}>Priority: {alert.priority}</span>
                {isLiveAlert && (
                  <span className="px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800 flex items-center gap-1">
                    <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                    Live
                  </span>
                )}
              </div>
              <p className=" font-medium mb-2">{formatAlertMessage(alert)}</p>
              <div className="flex items-center gap-4 text-sm text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Clock className="w-4 h-4" />
                  {formatTimestamp(alert.timestamp)}
                </span>
                <span>Rule: {alert.rule_type}</span>
                <span>Source: {alert.source}</span>
              </div>
            </div>
          </div>
          <div className="flex items-center gap-2 ml-4">
            {isRead ? (
              <Button variant="ghost" size="sm" onClick={() => markAsUnread(alert.id)} className="text-gray-500 hover:text-gray-700">
                <EyeOff className="w-4 h-4" />
              </Button>
            ) : (
              <Button variant="ghost" size="sm" onClick={() => markAsRead(alert.id)} className="text-gray-500 hover:text-gray-700">
                <Eye className="w-4 h-4" />
              </Button>
            )}
            <Button variant="ghost" size="sm" onClick={() => deleteAlert(alert.id)} className="text-red-500 hover:text-red-700">
              <Trash2 className="w-4 h-4" />
            </Button>
          </div>
        </div>
      </div>
    );
  };

  // Sort alerts by timestamp
  const sortedAlerts = alerts.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
  const filteredAlerts = showRead ? sortedAlerts : sortedAlerts.filter((alert) => !readAlerts.has(alert.id));

  const unreadCount = alerts.filter((alert) => !readAlerts.has(alert.id)).length;

  return (
    <div className="bg-background px-2 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] min-h-screen">
      <div className="flex items-center justify-between mt-3 p-5">
        <div className="relative flex items-center space-x-4">
          <div className="relative">
            <Bell className="h-8 w-8 text-white" />
            {unreadCount > 0 && (
              <div className="absolute -top-2 -right-2 bg-red-500 text-white rounded-full w-6 h-6 flex items-center justify-center text-xs font-bold">
                {unreadCount}
              </div>
            )}
          </div>
          <div>
            <h1 className="text-2xl font-bold">Alert Management</h1>
            <p className="text-sm text-gray-400">
              Project: {projectName} | Live Connection:{" "}
              <span className={isConnected ? "text-green-500" : "text-red-500"}>{isConnected ? "Connected" : "Disconnected"}</span>
            </p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <Button variant="outline" onClick={() => setShowRead(!showRead)} className="flex items-center gap-2">
            {showRead ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
            {showRead ? "Hide Read" : "Show Read"}
          </Button>

          <Link to={"manage"}>
            <Button className="flex items-center gap-2">
              <Settings className="w-4 h-4" />
              Create / Manage Alert Rules
            </Button>
          </Link>
        </div>
      </div>

      <div className="px-5 pb-5">
        {isLoading && (
          <div className="flex items-center justify-center py-8">
            <div className="text-gray-500">Loading alerts...</div>
          </div>
        )}

        {isError && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
            <div className="text-red-800 font-medium">Error loading alerts</div>
            <div className="text-red-600 text-sm">{error?.message || "Unknown error occurred"}</div>
          </div>
        )}

        {/* Unified Alerts List */}
        {filteredAlerts.length > 0 ? (
          <div>
            <h2 className="text-lg font-semibold mb-3">
              Alerts ({filteredAlerts.length}){!showRead && unreadCount > 0 && <span className="text-red-500 ml-2">({unreadCount} unread)</span>}
            </h2>
            {filteredAlerts.map((alert) => renderAlert(alert))}
          </div>
        ) : (
          <div className="text-center py-12">
            <Bell className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-600 mb-2">{isLoading ? "Loading alerts..." : "No Active Alerts"}</h3>
            <p className="text-gray-500">{showRead ? "No alerts to display." : "All alerts have been marked as read."}</p>
            {!isConnected && !isLoading && <p className="text-orange-500 mt-2">Live alerts are currently disconnected. Check your connection.</p>}
          </div>
        )}
      </div>
    </div>
  );
};

export default Alerts;

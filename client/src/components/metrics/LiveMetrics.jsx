import React, { useState, useEffect, useCallback, useMemo } from "react";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, BarChart, Bar, AreaChart, Area } from "recharts";
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from "@/components/ui/select";

const LiveMetrics = ({ metrics = [], performanceStats = {} }) => {
  const [historicalData, setHistoricalData] = useState([]);
  const [maxDataPoints, setMaxDataPoints] = useState(50);
  const [cpuChartView, setCpuChartView] = useState("coreUsage");

  const updateHistoricalData = useCallback(
    (newMetrics) => {
      if (!Array.isArray(newMetrics) || newMetrics.length === 0) return;

      const processedData = newMetrics.map((metric) => {
        const baseData = {
          timestamp: metric.memoryUsage?.timestamp || Date.now(),
          time: new Date(metric.memoryUsage?.timestamp || Date.now()).toLocaleTimeString(),
          cpuUsage: metric.cpuUsage?.average || 0,
          memoryUsage: metric.memoryUsage?.memoryUsagePercentage || 0,
          serviceName: metric.serviceName || "Unknown",
        };

        if (metric.cpuUsage?.cores) {
          metric.cpuUsage.cores.forEach((core) => {
            baseData[`core${core.core}`] = core.usage;
          });
        }

        return baseData;
      });

      setHistoricalData((prev) => {
        const combined = [...prev, ...processedData];
        const sorted = combined.sort((a, b) => a.timestamp - b.timestamp);
        return sorted.slice(-maxDataPoints);
      });
    },
    [maxDataPoints]
  );

  useEffect(() => {
    updateHistoricalData(metrics);
  }, [metrics, updateHistoricalData]);

  const latestMetrics = useMemo(() => {
    if (!metrics || metrics.length === 0) return null;
    return metrics[metrics.length - 1];
  }, [metrics]);

  // Format memory value
  const formatMemory = (bytes) => {
    if (!bytes) return "0 B";
    const units = ["B", "KB", "MB", "GB", "TB"];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(1)} ${units[unitIndex]}`;
  };

  // Custom tooltip for charts
  const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-gray-800 text-white p-3 rounded-lg shadow-lg border border-gray-600">
          <p className="text-sm font-medium">{`Time: ${label}`}</p>
          {payload.map((entry, index) => (
            <p key={index} className="text-sm" style={{ color: entry.color }}>
              {`${entry.name}: ${entry.value.toFixed(2)}%`}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  // Core usage chart data
  const coreUsageData = useMemo(() => {
    if (!latestMetrics?.cpuUsage?.cores) return [];
    return latestMetrics.cpuUsage.cores.map((core) => ({
      core: `Core ${core.core}`,
      usage: core.usage,
    }));
  }, [latestMetrics]);

  const coreColors = ["#ef4444", "#f97316", "#eab308", "#22c55e", "#06b6d4", "#3b82f6", "#8b5cf6", "#ec4899"];

  // Custom tooltip for area chart
  const AreaTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-gray-800 text-white p-3 rounded-lg shadow-lg border border-gray-600">
          <p className="text-sm font-medium">{`Time: ${label}`}</p>
          {payload.map((entry, index) => (
            <p key={index} className="text-sm" style={{ color: entry.color }}>
              {`${entry.name}: ${entry.value.toFixed(2)}%`}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  if (!latestMetrics) {
    return (
      <div className="w-full mx-auto p-6 my-2">
        <div className="flex items-center justify-center h-64 bg-gray-100 dark:bg-accent rounded-lg">
          <div className="text-center ">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
            <p className="mt-4 text-gray-600">Waiting for metrics data...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full mx-auto p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Live System Metrics</h1>
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <span>Service: {latestMetrics.serviceName}</span>
          <span>•</span>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-green-500 rounded-full animate-pulse"></div>
            <span>Live</span>
          </div>
        </div>
      </div>

      {/* Current Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-card p-6 rounded-xl shadow-sm border ">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">CPU Usage</p>
              <p
                className={`text-2xl font-bold ${
                  performanceStats.averageCpu < 50
                    ? "text-blue-600 dark:text-cpuusage"
                    : performanceStats.averageCpu < 80
                    ? "text-warn"
                    : "text-destructive"
                }`}
              >
                {performanceStats.averageCpu}%
              </p>
            </div>
            <div
              className={`p-3 ${
                performanceStats.averageCpu < 50
                  ? "bg-blue-600/20 dark:bg-cpuusage/20"
                  : performanceStats.averageCpu < 80
                  ? "bg-warn/20"
                  : "bg-destructive/20"
              } rounded-full`}
            >
              <div
                className={`w-6 h-6 ${
                  performanceStats.averageCpu < 50 ? " bg-blue-600 dark:bg-cpuusage" : performanceStats.averageCpu < 80 ? "bg-warn" : "bg-destructive"
                } rounded`}
              ></div>
            </div>
          </div>
        </div>

        <div className="bg-card p-6 rounded-xl shadow-sm border ">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Memory Usage</p>
              <p
                className={`text-2xl font-bold   ${
                  performanceStats.averageMemory < 50
                    ? "text-green-600 dark:text-green-400"
                    : performanceStats.averageMemory < 80
                    ? "text-warn"
                    : "text-destructive"
                }`}
              >
                {performanceStats.averageMemory}%
              </p>
            </div>
            <div
              className={`p-3 ${
                performanceStats.averageMemory < 50
                  ? "bg-green-100 dark:bg-green-400/20"
                  : performanceStats.averageMemory < 80
                  ? "bg-warn/20"
                  : "bg-destructive/20"
              } rounded-full`}
            >
              <div
                className={`w-6 h-6 ${
                  performanceStats.averageMemory < 50
                    ? "bg-green-600 dark:bg-green-400"
                    : performanceStats.averageMemory < 80
                    ? "bg-warn"
                    : "bg-destructive"
                } rounded`}
              ></div>
            </div>
          </div>
        </div>

        <div className="bg-card p-6 rounded-xl shadow-sm border ">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Total Memory</p>
              <p className="text-2xl font-bold text-purple-600">{formatMemory(latestMetrics.memoryUsage.totalMemory)}</p>
            </div>
            <div className="p-3 bg-purple-100 dark:bg-purple-400/20 rounded-full">
              <div className="w-6 h-6 bg-purple-600 dark:bg-purple-400 rounded"></div>
            </div>
          </div>
        </div>

        <div className="bg-card p-6 rounded-xl shadow-sm border ">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Free Memory</p>
              <p className="text-2xl font-bold text-orange-600">{formatMemory(performanceStats.freeMemory)}</p>
            </div>
            <div className="p-3 bg-orange-100 dark:bg-orange-400/20 rounded-full">
              <div className="w-6 h-6 bg-orange-600 dark:bg-orange-400 rounded"></div>
            </div>
          </div>
        </div>
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 2xl:grid-cols-2 gap-8">
        {/* Historical Usage Chart */}
        <div className="bg-card p-6 rounded-xl shadow-sm border ">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold text-gray-800 dark:text-white">Live Metrics</h2>
            <div className="flex items-center gap-2">
              <Select value={maxDataPoints} onValueChange={(value) => setMaxDataPoints(Number(value))}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Select max data points" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectLabel>Max Data Points</SelectLabel>
                    <SelectItem value={25}>25 points</SelectItem>
                    <SelectItem value={50}>50 points</SelectItem>
                    <SelectItem value={75}>75 points</SelectItem>
                    <SelectItem value={100}>100 points</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={historicalData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="time" stroke="#666" fontSize={12} tick={{ fill: "#666" }} />
              <YAxis stroke="#666" fontSize={12} tick={{ fill: "#666" }} domain={[0, 100]} />
              <Tooltip content={<CustomTooltip />} />
              <Legend />
              <Line
                type="monotone"
                dataKey="cpuUsage"
                stroke="#2563eb"
                strokeWidth={2}
                name="CPU Usage (%)"
                dot={{ fill: "#2563eb", strokeWidth: 2, r: 3 }}
                activeDot={{ r: 5, stroke: "#2563eb", strokeWidth: 2 }}
              />
              <Line
                type="monotone"
                dataKey="memoryUsage"
                stroke="#16a34a"
                strokeWidth={2}
                name="Memory Usage (%)"
                dot={{ fill: "#16a34a", strokeWidth: 2, r: 3 }}
                activeDot={{ r: 5, stroke: "#16a34a", strokeWidth: 2 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* CPU Cores Usage / Overall CPU Usage */}
        <div className="bg-card p-6 rounded-xl shadow-sm border ">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold text-gray-800 dark:text-white">Live Cpu Core Usage</h2>

            <Select value={cpuChartView} onValueChange={setCpuChartView}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Select CPU View" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectLabel>CPU View</SelectLabel>
                  <SelectItem value="coreUsage">Bar Graph</SelectItem>
                  <SelectItem value="overallUsage">Area Graph</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>

          {cpuChartView === "coreUsage" ? (
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={coreUsageData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="core" stroke="#666" fontSize={12} tick={{ fill: "#666" }} />
                <YAxis stroke="#666" fontSize={12} tick={{ fill: "#666" }} domain={[0, 100]} />
                <Tooltip
                  formatter={(value) => [`${value.toFixed(2)}%`, "Usage"]}
                  labelFormatter={(label) => `${label}`}
                  contentStyle={{
                    backgroundColor: "#1f2937",
                    border: "1px solid #374151",
                    borderRadius: "8px",
                    color: "#fff",
                  }}
                />
                <Bar dataKey="usage" fill="#8b5cf6" radius={[4, 4, 0, 0]} name="Core Usage (%)" />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={historicalData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                <XAxis dataKey="time" stroke="#666" fontSize={12} tick={{ fill: "#666" }} />
                <YAxis stroke="#666" fontSize={12} tick={{ fill: "#666" }} domain={[0, 100]} />
                <Tooltip content={<AreaTooltip />} />
                <Legend />
                {latestMetrics?.cpuUsage?.cores?.map((core, index) => (
                  <Area
                    key={`core${core.core}`}
                    type="monotone"
                    dataKey={`core${core.core}`}
                    stackId="1"
                    stroke={coreColors[index % coreColors.length]}
                    fill={coreColors[index % coreColors.length]}
                    fillOpacity={0.6}
                    name={`Core ${core.core}`}
                  />
                ))}
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>
      </div>

      {/* Performance Indicator */}
      <div className="mt-8 bg-card p-6 rounded-xl shadow-sm border">
        <h2 className="text-xl font-semibold text-gray-800 dark:text-white mb-4">System Status</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="flex items-center gap-3">
            <div
              className={`w-5 h-5 rounded-full ${
                latestMetrics.cpuUsage.average < 50 ? "bg-green-500" : latestMetrics.cpuUsage.average < 80 ? "bg-warn" : "bg-destructive"
              }`}
            ></div>
            <span className="text-sm font-medium">
              CPU Status: {latestMetrics.cpuUsage.average < 50 ? "Good" : latestMetrics.cpuUsage.average < 80 ? "Moderate" : "High"}
            </span>
          </div>
          <div className="flex items-center gap-3">
            <div
              className={`w-5 h-5 rounded-full ${
                latestMetrics.memoryUsage.memoryUsagePercentage < 60
                  ? "bg-green-500"
                  : latestMetrics.memoryUsage.memoryUsagePercentage < 80
                  ? "bg-warn"
                  : "bg-destructive"
              }`}
            ></div>
            <span className="text-sm font-medium">
              Memory Status:{" "}
              {latestMetrics.memoryUsage.memoryUsagePercentage < 60
                ? "Good"
                : latestMetrics.memoryUsage.memoryUsagePercentage < 80
                ? "Moderate"
                : "High"}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LiveMetrics;

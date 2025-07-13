import React, { useState, useEffect } from "react";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from "recharts";
import { CalendarDays, CalendarIcon, Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useParams } from "react-router-dom";
import { useLazyGetCpuUsageQuery } from "@/services/metricsServices";

import { Calendar } from "@/components/ui/calendar";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

const CpuUsageGraph = () => {
  const [getCpuUsage, { data: apiData, isLoading, isFetching, isError }] = useLazyGetCpuUsageQuery();
  const { projectName } = useParams();

  const transformData = (apiResponse) => {
    return apiResponse.map((bucket) => ({ timeLabel: bucket.timeLabel, average: bucket.average }));
  };

  const [data, setData] = useState(apiData?.data);
  const [timeRange, setTimeRange] = useState("hour");
  const [fromDate, setFromDate] = useState(new Date("2025-07-12"));
  const [toDate, setToDate] = useState(new Date("2025-07-13"));
  const [fromTime, setFromTime] = useState("0");
  const [toTime, setToTime] = useState("0");
  const [openc1, setOpenc1] = useState(false);
  const [openc2, setOpenc2] = useState(false);


  const formatDateForAPI = (date) => {
    return date.toISOString().split("T")[0];
  };


  const formatDateForInput = (date) => {
    return date.toISOString().split("T")[0];
  };


  const fetchData = () => {
    getCpuUsage({
      project: projectName,
      from: formatDateForAPI(fromDate) + "-" + fromTime,
      to: formatDateForAPI(toDate) + "-" + toTime,
      groupBy: timeRange,
    })
      .unwrap()
      .then((response) => {
        setData(transformData(response.data));
      });
  };

  useEffect(() => {
    fetchData();
  }, [timeRange, fromDate, toDate, fromTime, toTime]);

  const formatXAxisLabel = (tickItem) => {
    if (timeRange === "hour") {
      return new Date(tickItem).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        month: "short",
        day: "numeric",
      });
    } else {
      return new Date(tickItem).toLocaleDateString([], {
        month: "short",
        day: "numeric",
      });
    }
  };

  const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
      const data = payload[0].payload;
      return (
        <div className="bg-white p-3 border rounded-lg shadow-lg">
          <p className="font-medium">{`Time: ${data.timeLabel}`}</p>
          <p className="text-blue-600">{`CPU Usage: ${payload[0].value.toFixed(2)}%`}</p>
          <p className="text-gray-600 text-sm">{`Timestamp: ${data.timestamp}`}</p>
        </div>
      );
    }
    return null;
  };

  const getAverageCPU = () => {
    if (!data || data.length === 0) return 0;
    const validData = data.filter((item) => item.average !== null);
    const sum = validData.reduce((acc, item) => acc + item.average, 0);
    return (sum / validData.length).toFixed(2);
  };

  const getMaxCPU = () => {
    if (!data || data.length === 0) return 0;
    const validData = data.filter((item) => item.average !== null);
    return Math.max(...validData.map((item) => item.average)).toFixed(2);
  };

  const formatDate = (date) => {
    const options = { day: "numeric", month: "long", year: "numeric" };
    return new Date(date).toLocaleDateString("en-US", options);
  };

  return (
    <div className="w-full max-w-7xl space-y-4 mx-auto p-6">
      {/* Controls */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CalendarDays className="h-5 w-5" />
            CPU Usage Metrics
          </CardTitle>
          <CardDescription>Monitor CPU performance over time with customizable date ranges and intervals</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4 items-end">
            <div className="flex flex-col gap-3">
              <Label htmlFor="date" className="px-1">
                From Date
              </Label>
              <div className="relative flex gap-2">
                <Input
                  value={formatDateForInput(fromDate)}
                  onChange={(e) => setFromDate(new Date(e.target.value))}
                  id="from-date"
                  className="w-40"
                  onKeyDown={(e) => {
                    if (e.key === "ArrowDown") {
                      e.preventDefault();
                      setOpenc1(true);
                    }
                  }}
                />

                <Popover open={openc1} onOpenChange={setOpenc1}>
                  <PopoverTrigger asChild>
                    <Button id="date-picker" variant="ghost" className="absolute top-1/2 right-2 size-6 -translate-y-1/2">
                      <CalendarIcon className="size-3.5" />
                      <span className="sr-only">Select date</span>
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto overflow-hidden p-0" align="end">
                    <Calendar
                      mode="single"
                      selected={fromDate}
                      captionLayout="dropdown"
                      onSelect={(date) => {
                        setFromDate(date);
                        setOpenc1(false);
                      }}
                    />
                  </PopoverContent>
                </Popover>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">From Time</label>
              <Select value={fromTime} onValueChange={setFromTime}>
                <SelectTrigger className="w-20">
                  <SelectValue>{fromTime}:00</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {Array.from({ length: 24 }, (_, i) => (
                    <SelectItem key={i} value={i.toString()}>
                      {i}:00
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="date" className="px-1">
                To Date
              </Label>
              <div className="relative flex gap-2">
                <Input
                  value={formatDateForInput(toDate)}
                  onChange={(e) => setToDate(new Date(e.target.value))}
                  id="to-date"
                  className="w-40"
                  onKeyDown={(e) => {
                    if (e.key === "ArrowDown") {
                      e.preventDefault();
                      setOpenc1(true);
                    }
                  }}
                />

                <Popover open={openc2} onOpenChange={setOpenc2}>
                  <PopoverTrigger asChild>
                    <Button id="date-picker" variant="ghost" className="absolute top-1/2 right-2 size-6 -translate-y-1/2">
                      <CalendarIcon className="size-3.5" />
                      <span className="sr-only">Select date</span>
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto overflow-hidden p-0" align="end">
                    <Calendar
                      mode="single"
                      selected={toDate}
                      captionLayout="dropdown"
                      onSelect={(date) => {
                        setToDate(date);
                        setOpenc2(false);
                      }}
                    />
                  </PopoverContent>
                </Popover>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">To Time {toTime}</label>
              <Select value={toTime} onValueChange={setToTime}>
                <SelectTrigger className="w-20">
                  <SelectValue>{toTime}:00</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {Array.from({ length: 24 }, (_, i) => (
                    <SelectItem key={i} value={i.toString()}>
                      {i}:00
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">Interval</label>
              <Select value={timeRange} onValueChange={setTimeRange}>
                <SelectTrigger className="w-28">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="hour">
                    <div className="flex items-center gap-2">
                      <Clock className="h-4 w-4" />
                      Hour
                    </div>
                  </SelectItem>
                  <SelectItem value="day">
                    <div className="flex items-center gap-2">
                      <CalendarIcon className="h-4 w-4" />
                      Day
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <Button onClick={fetchData} disabled={isLoading}>
              {isLoading ? "Loading..." : "Refresh"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Graph */}
      <Card>
        <CardHeader>
          <CardTitle>CPU Usage Over Time</CardTitle>
          <CardDescription>{timeRange === "hour" ? "Hourly" : "Daily"} CPU usage metrics for project_1</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-96">
            {isLoading ? (
              <div className="flex items-center justify-center h-full">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                  <XAxis dataKey="timeLabel" tick={{ fontSize: 12 }} tickFormatter={formatXAxisLabel} />
                  <YAxis label={{ value: "CPU Usage (%)", angle: -90, position: "insideLeft" }} tick={{ fontSize: 12 }} />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend />
                  <Line
                    type="monotone"
                    dataKey="average"
                    stroke="#2563eb"
                    strokeWidth={2}
                    dot={{ fill: "#2563eb", strokeWidth: 2, r: 4 }}
                    activeDot={{ r: 6, stroke: "#2563eb", strokeWidth: 2 }}
                    name="CPU Usage (%)"
                  />
                </LineChart>
              </ResponsiveContainer>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Average CPU</p>
                <p className="text-2xl font-bold text-blue-600">{getAverageCPU()}%</p>
              </div>
              <Badge variant="secondary">AVG</Badge>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Peak CPU</p>
                <p className="text-2xl font-bold text-red-600">{getMaxCPU()}%</p>
              </div>
              <Badge variant="destructive">MAX</Badge>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};

export default CpuUsageGraph;

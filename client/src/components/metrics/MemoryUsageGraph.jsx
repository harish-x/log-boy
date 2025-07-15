import React, { useState, useEffect } from "react";
import { XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, AreaChart, Area } from "recharts";
import { CalendarIcon, Clock, MemoryStickIcon, RefreshCcwIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useParams } from "react-router-dom";
import { useLazyGetMemoryUsageQuery, useGetMetricsMinMaxdateQuery } from "@/services/metricsServices";
import { useNavigate } from "react-router-dom";
import { Calendar } from "@/components/ui/calendar";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { parseApiDate, getAvailableHours } from "@/lib/helpers";
import ProjectNotFound from "../ProjectNotFound";

const MemoryusageGraph = () => {
  const [getMemoryusage, { data: apiData, isLoading, isFetching, isError, error }] = useLazyGetMemoryUsageQuery();
  const { projectName } = useParams();
  const { data: minMaxData } = useGetMetricsMinMaxdateQuery(projectName);
  const minDate = minMaxData?.data?.[0]?.minDate ? parseApiDate(minMaxData.data[0].minDate) : null;
  const maxDate = minMaxData?.data?.[0]?.maxDate ? parseApiDate(minMaxData.data[0].maxDate) : null;

  const navigate = useNavigate();

  // State for date and time pickers
  const [fromDate, setFromDate] = useState(new Date(Date.now() - 24 * 60 * 60 * 1000)); // Default to 24 hours ago
  const [toDate, setToDate] = useState(new Date(Date.now())); // Default to today
  const [fromTime, setFromTime] = useState("0");
  const [toTime, setToTime] = useState(new Date().getHours().toString());

  // Other component state
  const [data, setData] = useState(apiData?.data);
  const [timeRange, setTimeRange] = useState("hour");
  const [openc1, setOpenc1] = useState(false);
  const [openc2, setOpenc2] = useState(false);
  const [lastUpdateTime, setLastUpdateTime] = useState(new Date().toLocaleString());

  // Effect to adjust selected dates if they fall outside the available range from the API
  useEffect(() => {
    if (minDate && maxDate) {
      if (fromDate < minDate || fromDate > maxDate) {
        setFromDate(minDate);
        setFromTime(minDate.getHours().toString());
      }
      if (toDate < minDate || toDate > maxDate) {
        setToDate(maxDate);
        setToTime(maxDate.getHours().toString());
      }
    }
  }, [minDate, maxDate]);

  // Handle navigation for critical errors
  if (isError && error?.data?.message !== "Project not found") {
    navigate("/404");
  }

  const transformData = (apiResponse) => {
    return apiResponse?.map((bucket) => ({ timeLabel: bucket.timeLabel, average: bucket.average }));
  };

  const formatDateForInput = (date) => {
    return date.toISOString().split("T")[0];
  };

  const fetchData = () => {
    const fromDateTime = new Date(fromDate);
    fromDateTime.setHours(parseInt(fromTime, 10), 0, 0, 0);

    const toDateTime = new Date(toDate);
    toDateTime.setHours(parseInt(toTime, 10), 59, 59, 999);

    const fromEpoch = fromDateTime.getTime();
    const toEpoch = toDateTime.getTime();

    getMemoryusage({
      project: projectName,
      from: fromEpoch,
      to: toEpoch,
      groupBy: timeRange,
    })
      .unwrap()
      .then((response) => {
        setData(transformData(response.data));
        setLastUpdateTime(new Date().toLocaleString());
      });
  };

  function formatLocalTime(timeString) {
    if (timeRange === "hour") {
      const date = new Date(timeString);
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    } else if (timeRange === "day") {
      const date = new Date(timeString);
      return date.toLocaleDateString([], { day: "numeric", month: "long", year: "numeric" });
    }
  }

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
        <div className="bg-foreground p-3 border rounded-lg shadow-lg">
          <p className="font-medium text-accent">{`Time: ${formatLocalTime(data.timeLabel)}`}</p>
          <p className="text-accent">{`Memory Usage: ${payload[0].value.toFixed(2)}%`}</p>
        </div>
      );
    }
    return null;
  };

  // Calculation functions for stats cards
  const getAverageMemory = () => {
    if (!data || data.length === 0) return 0;
    const validData = data.filter((item) => item.average !== null);
    if (validData.length === 0) return 0;
    const sum = validData.reduce((acc, item) => acc + item.average, 0);
    return (sum / validData.length).toFixed(2);
  };

  const getmaxMemory = () => {
    if (!data || data.length === 0) return 0;
    const validData = data.filter((item) => item.average !== null);
    if (validData.length === 0) return 0;
    return Math.max(...validData.map((item) => item.average)).toFixed(2);
  };

  const availableFromHours = minDate && maxDate ? getAvailableHours(fromDate, minDate, maxDate) : [];
  const availableToHours = minDate && maxDate ? getAvailableHours(toDate, minDate, maxDate) : [];

  const handleFromDateChange = (date) => {
    setFromDate(date);
    const availableHours = getAvailableHours(date, minDate, maxDate);
    if (!availableHours.includes(parseInt(fromTime))) {
      setFromTime(availableHours[0].toString());
    }
  };

  const handleToDateChange = (date) => {
    setToDate(date);
    const availableHours = getAvailableHours(date, minDate, maxDate);
    if (!availableHours.includes(parseInt(toTime))) {
      setToTime(availableHours[availableHours.length - 1].toString());
    }
  };

  const handleRefresh = () => {
    const now = new Date();
    setToDate(now);
    setToTime(now.getHours().toString());
  };

  // Render states for error, no data, and the main component
  if (isError && error?.data?.message === "Project not found") {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-5rem)] mx-auto rounded-2xl border border-primary/[0.20]">
        <ProjectNotFound />
      </div>
    );
  }

  if (apiData?.data === null || apiData?.data?.length === 0) {
    return (
      <div className="w-full mx-auto p-6">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <MemoryStickIcon className="h-5 w-5" />
              Memory Usage Metrics
            </CardTitle>
            <CardDescription className={"mt-1"}>Monitor Memory performance over time with customizable date ranges and intervals</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-4 items-end">
              <div className="flex flex-col gap-3 ">
                <Label htmlFor="date " className="px-1 text-center text-xl">
                  No Data Available
                </Label>
                <Button onClick={handleRefresh} disabled={isLoading} variant="outline">
                  <RefreshCcwIcon
                    className={cn("transition-transform", {
                      "animate-spin": isLoading || isFetching,
                    })}
                  />{" "}
                  Refresh
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="w-full mx-auto p-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MemoryStickIcon className="h-5 w-5" />
            Memory Usage Metrics
          </CardTitle>
          <CardDescription className={"mt-1"}>Monitor Memory performance over time with customizable date ranges and intervals</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-4 items-end">
            <div className="flex flex-col gap-3">
              <Label htmlFor="from-date" className="px-1">
                From Date
              </Label>
              <div className="relative flex gap-2">
                <Input value={formatDateForInput(fromDate)} readOnly id="from-date" className="w-40" />
                <Popover open={openc1} onOpenChange={setOpenc1}>
                  <PopoverTrigger asChild>
                    <Button id="date-picker-from" variant="ghost" className="absolute top-1/2 right-2 size-6 -translate-y-1/2">
                      <CalendarIcon className="size-3.5" />
                      <span className="sr-only">Select date</span>
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto overflow-hidden p-0" align="end">
                    <Calendar
                      mode="single"
                      selected={fromDate}
                      onSelect={(date) => {
                        if (date) handleFromDateChange(date);
                        setOpenc1(false);
                      }}
                      disabled={(date) => (minDate && date < minDate) || (maxDate && date > maxDate)}
                      fromDate={minDate}
                      toDate={maxDate}
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
                  {availableFromHours.map((hour) => (
                    <SelectItem key={hour} value={hour.toString()}>
                      {hour}:00
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-3">
              <Label htmlFor="to-date" className="px-1">
                To Date
              </Label>
              <div className="relative flex gap-2">
                <Input value={formatDateForInput(toDate)} readOnly id="to-date" className="w-40" />
                <Popover open={openc2} onOpenChange={setOpenc2}>
                  <PopoverTrigger asChild>
                    <Button id="date-picker-to" variant="ghost" className="absolute top-1/2 right-2 size-6 -translate-y-1/2">
                      <CalendarIcon className="size-3.5" />
                      <span className="sr-only">Select date</span>
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto overflow-hidden p-0" align="end">
                    <Calendar
                      mode="single"
                      selected={toDate}
                      onSelect={(date) => {
                        if (date) handleToDateChange(date);
                        setOpenc2(false);
                      }}
                      disabled={(date) => (minDate && date < minDate) || (maxDate && date > maxDate)}
                      fromDate={minDate}
                      toDate={maxDate}
                    />
                  </PopoverContent>
                </Popover>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium">To Time</label>
              <Select value={toTime} onValueChange={setToTime}>
                <SelectTrigger className="w-20">
                  <SelectValue>{toTime}:00</SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {availableToHours.map((hour) => (
                    <SelectItem key={hour} value={hour.toString()}>
                      {hour}:00
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

            <Button onClick={handleRefresh} disabled={isLoading} variant="outline" className="flex h-auto">
              <RefreshCcwIcon
                className={cn("transition-transform", {
                  "animate-spin": isLoading || isFetching,
                })}
              />
              <span className="text-xs text-muted-foreground text-center">Last Updated: {lastUpdateTime}</span>
            </Button>
          </div>
        </CardContent>

        {/* Graph */}
        <CardHeader className={"mt-5"}>
          <CardTitle>Memory Usage Over Time</CardTitle>
          <CardDescription>
            {timeRange === "hour" ? "Hourly" : "Daily"} memory usage metrics for {projectName}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-96">
            {isLoading ? (
              <div className="flex items-center justify-center h-full">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
                  <defs>
                    <linearGradient id="lineFillGradient2" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#82ca9d" stopOpacity={0.4} />
                      <stop offset="100%" stopColor="#82ca9d" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                  <XAxis dataKey="timeLabel" tick={{ fontSize: 12 }} tickFormatter={formatXAxisLabel} />
                  <YAxis label={{ value: "Memory Usage (%)", angle: -90, position: "insideLeft" }} tick={{ fontSize: 12 }} />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend />
                  <Area
                    type="monotone"
                    dataKey="average"
                    stroke="#82ca9d"
                    strokeWidth={2}
                    fill="url(#lineFillGradient2)"
                    fillOpacity={1}
                    dot={{ fill: "#82ca9d", strokeWidth: 2, r: 4 }}
                    activeDot={{ r: 6, stroke: "#82ca9d", strokeWidth: 2 }}
                    name="Memory Usage (%)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </div>
        </CardContent>

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 px-4 pb-4">
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Average Memory</p>
                  <p className="text-2xl font-bold text-[#82ca9d]">{getAverageMemory()}%</p>
                </div>
                <Badge variant="secondary">AVG</Badge>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Peak Memory</p>
                  <p className="text-2xl font-bold text-destructive">{getmaxMemory()}%</p>
                </div>
                <Badge variant="destructive">MAX</Badge>
              </div>
            </CardContent>
          </Card>
        </div>
      </Card>
    </div>
  );
};

export default MemoryusageGraph;

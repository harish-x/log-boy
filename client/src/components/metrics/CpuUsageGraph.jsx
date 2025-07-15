import React, { useState, useEffect } from "react";
import { XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, Area, AreaChart } from "recharts";
import { CalendarDays, CalendarIcon, Clock, RefreshCcwIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useParams } from "react-router-dom";
import { useGetMetricsMinMaxdateQuery, useLazyGetCpuUsageQuery } from "@/services/metricsServices";
import { Calendar } from "@/components/ui/calendar";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { parseApiDate, getAvailableHours } from "@/lib/helpers";


const CpuUsageGraph = () => {
  const [getCpuUsage, { data: apiData, isLoading, isFetching, isError }] = useLazyGetCpuUsageQuery();
  const { projectName } = useParams();

  const { data: minMaxData } = useGetMetricsMinMaxdateQuery(projectName);

  const minDate = minMaxData?.data?.[0]?.minDate ? parseApiDate(minMaxData.data[0].minDate) : null;
  const maxDate = minMaxData?.data?.[0]?.maxDate ? parseApiDate(minMaxData.data[0].maxDate) : null;

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

  const transformData = (apiResponse) => {
    return apiResponse.map((bucket) => ({ timeLabel: bucket.timeLabel, average: bucket.average }));
  };

  const [data, setData] = useState(apiData?.data);
  const [timeRange, setTimeRange] = useState("hour");
  const [fromDate, setFromDate] = useState(new Date(Date.now() - 24 * 60 * 60 * 1000)); // Default to 24 hours ago
  const [toDate, setToDate] = useState(new Date(Date.now())); // Default to today
  const [fromTime, setFromTime] = useState("0");
  const [toTime, setToTime] = useState(new Date().getHours().toString());
  const [openc1, setOpenc1] = useState(false);
  const [openc2, setOpenc2] = useState(false);
  const [lastUpdateTime, setLastUpdateTime] = useState(new Date().toLocaleString());
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
        setLastUpdateTime(new Date().toLocaleString());
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
        <div className="bg-foreground p-3 border rounded-lg shadow-lg">
          <p className="font-medium text-accent">{`Time: ${data.timeLabel}`}</p>
          <p className="text-blue-600">{`Memory Usage: ${payload[0].value.toFixed(2)}%`}</p>
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

  const availableFromHours = minDate && maxDate ? getAvailableHours(fromDate, minDate, maxDate) : [];
  const availableToHours = minDate && maxDate ? getAvailableHours(toDate, minDate, maxDate) : [];

  // Validate and update time when date changes
  const handleFromDateChange = (date) => {
    setFromDate(date);
    const availableHours = getAvailableHours(date, minDate, maxDate);

    // If current time is not available for this date, set to first available hour
    if (!availableHours.includes(parseInt(fromTime))) {
      setFromTime(availableHours[0].toString());
    }
  };

  const handleToDateChange = (date) => {
    setToDate(date);
    const availableHours = getAvailableHours(date, minDate, maxDate);

    // If current time is not available for this date, set to last available hour
    if (!availableHours.includes(parseInt(toTime))) {
      setToTime(availableHours[availableHours.length - 1].toString());
    }
  };

  return (
    <div className="w-full  mx-auto p-6">
      {/* Controls */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CalendarDays className="h-5 w-5" />
            CPU Usage Metrics
          </CardTitle>
          <CardDescription className={"mt-1"}>Monitor CPU performance over time with customizable date ranges and intervals</CardDescription>
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
                  id="from-date"
                  className="w-40"
                  type={"readonly"}
                  readOnly
                  min={minDate ? formatDateForInput(minDate) : undefined}
                  max={maxDate ? formatDateForInput(maxDate) : undefined}
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
                      onSelect={(date) => {
                        handleFromDateChange(date);
                        setOpenc1(false);
                      }}
                      disabled={(date) => {
                        if (!minDate || !maxDate) return false;
                        return date < minDate || date > maxDate;
                      }}
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
              <Label htmlFor="date" className="px-1">
                To Date
              </Label>
              <div className="relative flex gap-2">
                <Input
                  value={formatDateForInput(toDate)}
                  id="to-date"
                  className="w-40"
                  type={"readonly"}
                  readOnly
                  min={minDate ? formatDateForInput(minDate) : undefined}
                  max={maxDate ? formatDateForInput(maxDate) : undefined}
                  onKeyDown={(e) => {
                    if (e.key === "ArrowDown") {
                      e.preventDefault();
                      setOpenc2(true);
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
                      onSelect={(date) => {
                        handleToDateChange(date);
                        setOpenc2(false);
                      }}
                      disabled={(date) => {
                        if (!minDate || !maxDate) return false;
                        return date < minDate || date > maxDate;
                      }}
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

            <Button onClick={fetchData} disabled={isLoading} variant="outline">
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
          <CardTitle>CPU Usage Over Time</CardTitle>
          <CardDescription>
            {timeRange === "hour" ? "Hourly" : "Daily"} CPU usage metrics for {projectName}
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
                    <linearGradient id="lineFillGradient" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#2563eb" stopOpacity={0.4} />
                      <stop offset="100%" stopColor="#2563eb" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
                  <XAxis dataKey="timeLabel" tick={{ fontSize: 12 }} tickFormatter={formatXAxisLabel} />
                  <YAxis label={{ value: "CPU Usage (%)", angle: -90, position: "insideLeft" }} tick={{ fontSize: 12 }} />
                  <Tooltip content={<CustomTooltip />} />
                  <Legend />

                  <Area
                    type="monotone"
                    dataKey="average"
                    stroke="#2563eb"
                    strokeWidth={2}
                    fill="url(#lineFillGradient)"
                    fillOpacity={1}
                    dot={{ fill: "#2563eb", strokeWidth: 2, r: 4 }}
                    activeDot={{ r: 6, stroke: "#2563eb", strokeWidth: 2 }}
                    name="CPU Usage (%)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </div>
        </CardContent>

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 px-4">
          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Average CPU</p>
                  <p className="text-2xl font-bold text-[#2563eb]">{getAverageCPU()}%</p>
                </div>
                <Badge variant="secondary">AVG</Badge>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Peak CPU</p>
                  <p className="text-2xl font-bold text-destructive">{getMaxCPU()}%</p>
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

export default CpuUsageGraph;

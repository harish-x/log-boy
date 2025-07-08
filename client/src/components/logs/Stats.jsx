import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { AlertTriangle, CheckCircle2, Info, CalendarDays, Clock, DatabaseZap, AlertCircle } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";

export default function Stats({ profile, isFetchingProject, isLoadingProject }) {
  // loading skeleton
  if (isLoadingProject || isFetchingProject) {
    return (
      <>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 p-4 md:p-6">
          <Skeleton className="w-full h-60" />
          <Skeleton className="w-full h-60" />
          <Skeleton className="w-full h-60" />
          <Skeleton className="w-full h-60" />
          <Skeleton className="w-full h-60 sm:col-span-2 lg:col-span-1 xl:col-span-2" />
          <Skeleton className="w-full h-60" />
        </div>
      </>
    );
  }

  const {
    name = "Unnamed Profile",
    description = "No description provided.",
    active = false,
    error_ratio = 0,
    total_errors = 0,
    total_requests = 0,
    retention_period = 0,
    total_warnings = 0,
    created_at = null,
  } = profile; // Destructure the profile object

  // Function to format date with time
  function formatDatewithTime(data) {
    if (!data) {
      return "Not specified";
    }
    const date = new Date(data);

    if (isNaN(date.getTime())) {
      return "Invalid Date";
    }
    return date.toLocaleString("en-US", {
      day: "numeric",
      month: "long",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  const success_ratio = 100 - error_ratio;

  return (
    <Card className="w-full p-4 rounded-2xl shadow-md">
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-xl font-bold">{name}</h2>
          <p className="text-sm text-gray-500">{description}</p>
        </div>
        <Badge variant={active ? "default" : "secondary"} className="text-xs mr-2">
          {active ? "Active" : "Inactive"}
        </Badge>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6 p-4 md:p-6">
        <Card className="sm:col-span-1 lg:col-span-1 xl:col-span-1 flex flex-col justify-between">
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center text-sm text-muted-foreground">
              <AlertCircle className="mr-2 h-4 w-4 text-destructive" />
              Error Ratio
            </CardDescription>
            <CardTitle className="text-4xl font-bold text-destructive">{error_ratio.toFixed(1)}%</CardTitle>
          </CardHeader>
          <CardContent className="flex justify-center items-center pt-0">
            <div className="w-32 h-32">
              <svg viewBox="0 0 36 36" className="w-full h-full">
                <path
                  className="text-slate-200 dark:text-slate-700"
                  strokeWidth="3.8"
                  fill="none"
                  stroke="currentColor"
                  d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                />
                <path
                  className="text-destructive"
                  strokeWidth="3.8"
                  fill="none"
                  strokeLinecap="round"
                  stroke="currentColor"
                  strokeDasharray={`${error_ratio}, 100`}
                  d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                />
                <text
                  x="18"
                  y="18"
                  dy="1.5"
                  textAnchor="middle"
                  fill="currentColor"
                  className="font-semibold text-slate-700 dark:text-slate-200"
                  fontSize="5.6"
                >
                  {error_ratio.toFixed(1)}%
                </text>
              </svg>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center text-sm text-muted-foreground">
              <DatabaseZap className="mr-2 h-4 w-4 text-blue-500" />
              Total Requests
            </CardDescription>
            <CardTitle className="text-3xl font-bold">{total_requests.toLocaleString()}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">Overall requests processed.</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center text-sm text-muted-foreground">
              <AlertCircle className="mr-2 h-4 w-4 text-destructive" />
              Total Errors
            </CardDescription>
            <CardTitle className="text-3xl font-bold text-destructive">{total_errors.toLocaleString()}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">Critical issues encountered.</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center text-sm text-muted-foreground">
              <AlertTriangle className="mr-2 h-4 w-4 text-warn" />
              Total Warnings
            </CardDescription>
            <CardTitle className="text-3xl font-bold text-warn">{total_warnings.toLocaleString()}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-muted-foreground">Non-critical issues.</p>
          </CardContent>
        </Card>

        <Card className="sm:col-span-2 lg:col-span-1 xl:col-span-2">
          <CardHeader>
            <CardTitle className="text-xl flex items-center">
              <Info className="mr-2 h-5 w-5 text-sky-600" />
              Instance Details
            </CardTitle>
            <CardDescription>Key information about this data instance.</CardDescription>
          </CardHeader>
          <CardContent className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground font-medium flex items-center mb-1">
                <Clock className="mr-2 h-4 w-4 text-gray-500" />
                Retention Period
              </p>
              <p className="font-semibold text-lg">
                {retention_period} <span className="text-sm font-normal text-muted-foreground">days</span>
              </p>
            </div>
            <div>
              <p className="text-muted-foreground font-medium flex items-center mb-1">
                <CalendarDays className="mr-2 h-4 w-4 text-gray-500" />
                Created At
              </p>
              <p className="font-semibold text-base">{formatDatewithTime(created_at)}</p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription className="flex items-center text-sm text-muted-foreground">
              <CheckCircle2 className="mr-2 h-4 w-4 text-green-500" />
              Success Ratio
            </CardDescription>
            <CardTitle className="text-3xl font-bold text-green-600">{success_ratio.toFixed(2)}%</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="w-24 h-24 mx-auto">
              <svg viewBox="0 0 36 36" className="w-full h-full">
                <path
                  className="text-slate-200 dark:text-slate-700"
                  strokeWidth="3.8"
                  fill="none"
                  stroke="currentColor"
                  d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                />
                <path
                  className="text-green-500"
                  strokeWidth="3.8"
                  fill="none"
                  strokeLinecap="round"
                  stroke="currentColor"
                  strokeDasharray={`${success_ratio}, 100`}
                  d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                />
                <text
                  x="18"
                  y="18"
                  dy="1.5"
                  textAnchor="middle"
                  fill="currentColor"
                  className="font-semibold text-slate-700 dark:text-slate-200"
                  fontSize="5.6"
                >
                  {success_ratio.toFixed(1)}%
                </text>
              </svg>
            </div>
          </CardContent>
        </Card>
      </div>
    </Card>
  );
}

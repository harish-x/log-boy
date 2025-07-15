import React, { useEffect, useState } from "react";
import { Calendar } from "@/components/ui/calendar";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { AlertTriangle, Bug, CalendarIcon, CircleX, FileWarning, Info, MoveDown, MoveUp, RefreshCcw } from "lucide-react";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import { useLazyGetLogsQuery, useLazyGetMinMaxDateQuery } from "@/services/LogServices";
import { useGetProjectByNameQuery } from "@/services/ProjectService";
import { useParams } from "react-router-dom";
import LogLists from "./LogLists";
import { useDispatch } from "react-redux";
import { addRecentProjects } from "@/slice/globalState";
import ProjectNotFound from "../ProjectNotFound";
import ProjectError from "../ProjectError";

const Logs = () => {
  const [startDate, setStartDate] = useState(null);
  const [endDate, setEndDate] = useState(null);
  const [sortDate, setSortDate] = useState("desc");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(100);
  const [logType, setLogType] = useState(null);
  const [refresh, setRefresh] = useState(false);
  const dispatch = useDispatch();
  const [getLogs, { data: logsData, isLoading, isFetching: isFetchingLogs, error: logError }] = useLazyGetLogsQuery();
  const { projectName } = useParams();
  const { isError: isErrorProject, error: projectError } = useGetProjectByNameQuery(projectName);
  const [getMinMaxDate, { data: minMaxDate, isLoading: isLoadingDate, isError: isErrorDate, isFetching: isFetchingDate }] =
    useLazyGetMinMaxDateQuery();

  {
    /*------------pagination------------*/
  }
  const total = logsData?.data.total || 0; // total number of logs
  const totalPages = Math.ceil(total / limit); // total number of pages

  // function to get the page numbers
  const getPageNumbers = (currentPage, totalPages, maxPages = 5) => {
    const pages = [];
    if (totalPages <= maxPages) {
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      let start = Math.max(1, currentPage - Math.floor(maxPages / 2));
      let end = Math.min(totalPages, start + maxPages - 1);
      if (end - start + 1 < maxPages) {
        start = Math.max(1, end - maxPages + 1);
      }
      for (let i = start; i <= end; i++) {
        pages.push(i);
      }
    }
    return pages;
  };

  const pageNumbers = getPageNumbers(page, totalPages);

  // function to clear filters
  function clearFilters() {
    setStartDate(null);
    setEndDate(null);
    setSortDate("desc");
    setPage(1);
    setLimit(100);
    setLogType(null);
    setRefresh((prev) => !prev);
  }

  useEffect(() => {
    const fetchLogs = async () => {
      try {
        await getLogs({
          fromdate: startDate ? startDate.toISOString() : null,
          todate: endDate ? endDate.toISOString() : null,
          SortByDate: sortDate,
          page: page,
          limit: limit,
          project: projectName,
          mode: logType,
        })
          .unwrap()
          .finally(() => {
            dispatch(addRecentProjects(projectName));
          });
        await getMinMaxDate(projectName);
      } catch (error) {
        if (error?.data?.message === "No logs found") {
          toast.error("No logs found");
        } else {
          toast.error("Failed to fetch logs", {
            description: "Could not fetch logs. Please try again.",
            variant: "destructive",
            richColors: true,
            action: <AlertTriangle className="text-red-500" />,
          });
        }
      }
    };

    fetchLogs();
  }, [page, limit, startDate, endDate, sortDate, getLogs, logType, refresh, projectName]);
  if (isErrorDate) {
    toast.error(`${minMaxDate?.data?.message}`);
  }
  const formatDate = (date) => {
    const options = { day: "numeric", month: "long", year: "numeric" };
    return new Date(date).toLocaleDateString("en-US", options);
  };

  function toggleDateSort() {
    if (sortDate === "desc") {
      setSortDate("asc");
    } else {
      setSortDate("desc");
    }
  }

  if (isErrorProject && projectError.data.message === "Project not found") {
    return <ProjectNotFound projectName={projectName} />;
  }

  if (isErrorProject && projectError.data.message !== "Project not found") {
    toast.error(`${projectError?.data?.message}` || "An error occurred while loading the project.");
    return <ProjectError projectName={projectName} />;
  }
  return (
    <div>
      <div
        className="flex items-center justify-between px-4 py-2  border-t border-r border-l border-primary/[0.20] mt-5 rounded-t-2xl"
        id="logs-header"
      >
        <div className="flex items-center gap-2">
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className={cn("w-[280px] justify-start text-left font-normal", !startDate && "text-muted-foreground")}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {startDate ? formatDate(startDate) : <span>Pick a date</span>}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <Calendar
                selected={startDate}
                onSelect={setStartDate}
                mode="single"
                fromDate={minMaxDate?.data?.oldest}
                toDate={minMaxDate?.data?.latest}
              />
            </PopoverContent>
          </Popover>
          <span className="text-muted-foreground">to</span>
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                size="sm"
                className={cn("w-[280px] justify-start text-left font-normal", !endDate && "text-muted-foreground")}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {endDate ? formatDate(endDate) : <span>Today</span>}
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <Calendar
                mode="single"
                selected={endDate}
                onSelect={setEndDate}
                fromDate={minMaxDate?.data?.oldest}
                toDate={minMaxDate?.data?.latest}
              />
            </PopoverContent>
          </Popover>
        </div>
        <div className="flex items-center gap-2 text-muted-foreground">
          <Button variant="outline" size="sm" className={"cursor-pointer"} onClick={() => setRefresh((prev) => !prev)}>
            <RefreshCcw
              className={cn("transition-transform", {
                "animate-spin": isLoading || isFetchingLogs || isFetchingDate || isLoadingDate,
              })}
            />
          </Button>
          <Button variant="outline" size="sm" onClick={toggleDateSort}>
            {sortDate === "desc" ? <MoveUp /> : <MoveDown />} Date
          </Button>

          <Select onValueChange={setLogType} value={logType}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="Log Mode" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={null}>All</SelectItem>
              <SelectItem value="info">
                <Info />
                Info
              </SelectItem>
              <SelectItem value="debug">
                <Bug /> Debug
              </SelectItem>
              <SelectItem value="warn">
                <FileWarning /> Warn
              </SelectItem>
              <SelectItem value="error">
                <CircleX /> Error
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <LogLists logError={logError} isLoading={isLoading} isFetchingLogs={isFetchingLogs} logsData={logsData} clearFilters={clearFilters} />

      {totalPages > 1 && (
        <div className="mt-5">
          <Pagination>
            <PaginationContent>
              <PaginationPrevious onClick={() => setPage((prev) => Math.max(prev - 1, 1))} disabled={page === 1}>
                Previous
              </PaginationPrevious>

              {pageNumbers[0] > 1 && (
                <>
                  <PaginationItem>
                    <PaginationLink onClick={() => setPage(1)}>
                      <span className="cursor-pointer">1</span>
                    </PaginationLink>
                  </PaginationItem>
                  {pageNumbers[0] > 2 && <PaginationEllipsis />}
                </>
              )}

              {pageNumbers.map((p) => (
                <PaginationItem key={p}>
                  <PaginationLink isActive={page === p} onClick={() => setPage(p)} className={page === p ? "font-bold" : ""}>
                    <span className="cursor-pointer">{p}</span>
                  </PaginationLink>
                </PaginationItem>
              ))}

              {pageNumbers[pageNumbers.length - 1] < totalPages && (
                <>
                  {pageNumbers[pageNumbers.length - 1] < totalPages - 1 && <PaginationEllipsis />}
                  <PaginationItem>
                    <PaginationLink onClick={() => setPage(totalPages)}>
                      <span className="cursor-pointer">{totalPages}</span>
                    </PaginationLink>
                  </PaginationItem>
                </>
              )}

              <PaginationNext onClick={() => setPage((prev) => Math.min(prev + 1, totalPages))} disabled={page === totalPages}>
                Next
              </PaginationNext>
            </PaginationContent>
          </Pagination>
        </div>
      )}
    </div>
  );
};

export default Logs;

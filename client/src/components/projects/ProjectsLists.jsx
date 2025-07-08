import React from "react";
import { ScrollArea } from "../ui/scroll-area";
import { Calendar, EllipsisVertical } from "lucide-react";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "../ui/pagination";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { useGetProjectsQuery } from "@/services/ProjectService";
import { Skeleton } from "../ui/skeleton";
import { Link, useNavigate } from "react-router-dom";

const ProjectsLists = () => {
  const [page, setPage] = React.useState(1);
  const [limit, setLimit] = React.useState(15);
  const { data: projects, isLoading } = useGetProjectsQuery({
    page: page,
    limit: limit,
  });

  /*---------------------pagination--------------------------------------*/

  const total = projects?.data?.total || 0; // total number of projects
  const totalPages = Math.ceil(total / limit); // total number of pages

  const getPageNumbers = (currentPage, totalPages, maxPages = 5) => {
    // function to get the page numbers
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

  const navigate = useNavigate();
  const FormatDate = (date) => {
    const options = { day: "numeric", month: "long", year: "numeric" };
    return new Date(date).toLocaleDateString("en-US", options);
  };
  return (
    <div className="projects-lists mt-5 w-[98%] mx-auto ">
      {isLoading ? (
        <Skeleton className="w-full h-[60vh] rounded p-2 "></Skeleton> // skeleton loader
      ) : projects.data.projects.length > 0 ? (
        <ScrollArea className="bg-accent w-full max-h-[60vh] rounded p-2 ">
          <ul>
            {projects?.data?.projects.map((project, index) => (
              <Link to={`/dashboard/project/${project.name}`} key={index}>
                <li className="flex justify-between p-2 last:border-0 border-b border-primary/[0.20] first:rounded-t-md last:rounded-b-md hover:bg-primary/[0.20] cursor-pointer ">
                  <span className="font-semibold text-accent-foreground">{project.name}</span>
                  <div className="flex gap-2 items-center justify-end">
                    <Calendar opacity={0.5} /> <span>{FormatDate(project.created_at)}</span>
                  </div>

                  <DropdownMenu>
                    <DropdownMenuTrigger className="hover:bg-secondary p-1  rounded-full cursor-pointer">
                      <EllipsisVertical opacity={0.7} />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent>
                      <DropdownMenuLabel>{project.name}</DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => navigate(`/dashboard/project/${project.name}`)}>View</DropdownMenuItem>
                      <DropdownMenuItem onClick={() => navigate(`/dashboard/project/${project.name}/settings`)}>Settings</DropdownMenuItem>
                      <DropdownMenuItem className={"bg-destructive text-destructive-foreground mt-1 hover:bg-red-600"}>Delete</DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </li>
              </Link>
            ))}
          </ul>
        </ScrollArea>
      ) : (
        <>
          <div className="flex items-center justify-center h-[60vh] text-2xl capitalize font-bold">No projects found</div>
        </>
      )}

      {/* pagination */}

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

export default ProjectsLists;

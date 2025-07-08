import React from "react";
import { useParams } from "react-router-dom";
import ListArchiveAvailableLogs from "@/components/logs/ListArchiveAvailableLogs";
import { useGetArchiveLogListQuery } from "@/services/LogServices";

const ColdStorage = () => {
  const { projectName } = useParams();

  const {
    isLoading: isLoadingProject,
    isError: isErrorProject,
    isFetching: isFetchingProject,
    data: availableLogs,
  } = useGetArchiveLogListQuery(projectName);
  return (
    <div className="projects-container bg-background px-2 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] h-[calc(100vh-5rem)] overflow-hidden">
      <ListArchiveAvailableLogs isLoading={isLoadingProject} isError={isErrorProject} isFetching={isFetchingProject} data={availableLogs} />
    </div>
  );
};

export default ColdStorage;

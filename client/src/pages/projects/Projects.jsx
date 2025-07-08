import React from "react";
import RecentProjects from "@/components/projects/RecentProjects";
import ProjectsLists from "@/components/projects/ProjectsLists";

const Projects = () => {
  return (
    <div className="projects-container bg-background px-2 w-[98%] mx-auto rounded-2xl border border-primary/[0.20] h-[calc(100vh-5rem)] overflow-hidden">
      <RecentProjects />
      <ProjectsLists />
    </div>
  );
};

export default Projects;

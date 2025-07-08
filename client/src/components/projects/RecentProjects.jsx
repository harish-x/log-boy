import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import {
  useLazyGetRecentProjectsQuery,
  useLazyGetProjectsQuery,
} from "@/services/ProjectService";

const RecentProjects = () => {
  const navigate = useNavigate();
  const [displayProjects, setDisplayProjects] = useState([]);
  const [projectsToShow, setProjectsToShow] = useState(
    getProjectByScreenSize()
  );

  // Function to handle project click
  const handleProjectClick = (projectName = null) => {
    if (projectName) {
      navigate(`project/${projectName}`);
    } else {
      navigate(`create`);
    }
  };

  const [getRecentProjects, { data: recentProjectsData }] =
    useLazyGetRecentProjectsQuery(); // Get recent projects
  const [getProjects, { data: allProjectsData }] = useLazyGetProjectsQuery(); // Get all projects

  const recentProjects = useSelector(
    // Get recently viewed projects from local storage
    (state) => state.globalState.recentProjects
  );

  const parseProjectNames = (projects) => {
    if (!projects) return [];

    if (typeof projects === "string") {
      return projects
        .split(",")
        .map((name) => name.trim())
        .filter((name) => name);
    }

    if (Array.isArray(projects)) {
      return projects.filter((name) => name && name.trim());
    }

    return [];
  };

  // Get the number of projects to display based on screen size
  function getProjectByScreenSize() {
    if (window.innerWidth >= 2560) {
      return 5;
    }
    if (window.innerWidth >= 2048) {
      return 4;
    }
    if (window.innerWidth >= 1536) {
      return 3;
    } else if (window.innerWidth >= 768) {
      return 2;
    } else {
      return 1;
    }
  }

  // Function to handle window resize
  useEffect(() => {
    const handleResize = () => {
      const newProjectsToShow = getProjectByScreenSize();
      setProjectsToShow(newProjectsToShow);
    };

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  // Fetch recent projects via API
  useEffect(() => {
    const parsedRecentProjects = parseProjectNames(recentProjects);

    if (parsedRecentProjects.length > 0) {
      const recentProjectsString = parsedRecentProjects.join(",");
      getRecentProjects({ p: recentProjectsString });
    } else {
      getRecentProjects({});
    }
  }, [getRecentProjects, recentProjects]);


  // Fetch all projects if the number of projects to display is greater than the number of recent projects
  useEffect(() => {
    const updateDisplayProjects = async () => {
      let projectsToDisplay = [];
      if (recentProjectsData?.data && Array.isArray(recentProjectsData.data)) {
        projectsToDisplay = [...recentProjectsData.data];
      }

      if (projectsToDisplay.length < projectsToShow) {
        const neededProjects = projectsToShow - projectsToDisplay.length;
        const result = await getProjects({
          limit: neededProjects,
          page: 1,
        });

        if (result.data?.data && Array.isArray(result.data.data.projects)) {
          const existingProjectIds = projectsToDisplay.map((p) => p.id);
          const filteredAdditionalProjects = result.data.data.projects.filter(
            (project) => !existingProjectIds.includes(project.id)
          );

         
          projectsToDisplay = [
            ...projectsToDisplay,
            ...filteredAdditionalProjects.slice(0, neededProjects),
          ];
        }
      }

      setDisplayProjects(projectsToDisplay.slice(0, projectsToShow));
    };

    updateDisplayProjects();
  }, [recentProjectsData, allProjectsData, projectsToShow, getProjects]);

  const renderProjectCard = (project, index) => {
    return (
      <div
        key={project.id || index}
        className="h-36 bg-accent p-5 flex-1 mx-2 rounded border-primary/[0.20] cursor-pointer hover:bg-accent/80 transition-colors"
        onClick={() => handleProjectClick(project.name)}
      >
        <p className="font-semibold text-primary text-lg">{project.name}</p>
        <p className="text-sm mt-2 text-muted-foreground">
          {project.description}
        </p>
      </div>
    );
  };

  return (
    <div>
      <div className="recent-projects flex justify-evenly mt-5">
        <div
          className="h-36 bg-accent p-5 mx-2 rounded flex-1 border-primary/[0.20] cursor-pointer hover:bg-accent/80 transition-colors"
          onClick={() => handleProjectClick()}
        >
          <div className="flex justify-center items-center">
            <svg
              width="49"
              height="49"
              viewBox="0 0 49 49"
              fill="none"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path
                d="M24.497.668v47.667M.664 24.5h47.667"
                stroke="currentColor"
                className="stroke-primary"
                strokeWidth="2"
              ></path>
            </svg>
          </div>
          <p className="mt-5 font-semibold text-center self-baseline text-primary">
            Create New Project
          </p>
        </div>
        {displayProjects.map((project, index) =>
          renderProjectCard(project, index)
        )}
      </div>
    </div>
  );
};

export default RecentProjects;

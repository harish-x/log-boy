import { GalleryVerticalEnd, FolderCode, ChevronDown, HomeIcon, GitBranchPlus, User2 } from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarMenuSub,
  SidebarMenuSubItem,
  SidebarMenuSubButton,
} from "./ui/sidebar";

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "./ui/collapsible";
import { useGetProjectsQuery } from "@/services/ProjectService";
import { useNavigate, useLocation, Link } from "react-router-dom";
import { useMsal } from "@azure/msal-react";
import { NavUser } from "./projects/nav-user";
import { Skeleton } from "./ui/skeleton";

export function ProjectSidebar({ ...props }) {
  const { data: projects, isLoading } = useGetProjectsQuery({
    page: 1,
    limit: 5,
  });
  const navigate = useNavigate();
  const location = useLocation();
  const { accounts, instance } = useMsal();
  const userName = accounts[0].name || "User";
  const userEmail = accounts[0].username || "User";

  const handleLogout = () => {
    instance.logoutRedirect({ postLogoutRedirectUri: "/" });
  };
  return (
    <Sidebar {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <a href="#">
                <div className="flex aspect-square size-10 items-center justify-center rounded-lg  text-sidebar-primary-foreground">
                  <img src="/logboy.png" />
                </div>
                <div className="flex flex-col gap-0.5 leading-none">
                  <span className="font-semibold">LOG BOY</span>
                </div>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarMenuButton size="icon" className={"py-2"}>
            <div className="flex items-center ">
              <HomeIcon />
              <span className="ml-2 text-base">Home</span>
            </div>
          </SidebarMenuButton>
          <SidebarMenuButton
            isActive={location.pathname === "/dashboard/projects/create"}
            size="icon"
            className={"mt-4 cursor-pointer"}
            onClick={() => navigate("/dashboard/projects/create")}
          >
            <div className="flex items-center ">
              <GitBranchPlus />
              <span className="ml-2 text-base">Create Project</span>
            </div>
          </SidebarMenuButton>
        </SidebarGroup>

        <SidebarGroup className={"projects-group"}>
          <SidebarMenu>
            <Collapsible defaultOpen className="group/collapsible">
              <SidebarMenuItem>
                <CollapsibleTrigger asChild>
                  <SidebarMenuButton isActive={location.pathname === "/dashboard/projects"} size="icon">
                    <div className="flex items-center gap-2">
                      <FolderCode />
                      <span className="text-base">projects</span>
                    </div>
                    <ChevronDown className="ml-auto" />
                  </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                  <SidebarMenuSub>
                    {isLoading ? (
                      <SidebarMenuSubItem>
                      
                          <div className="flex flex-col w-full gap-2 mt-3">
                            <Skeleton className="h-4 w-full"></Skeleton>
                            <Skeleton className="h-4 w-full"></Skeleton>
                            <Skeleton className="h-4 w-full"></Skeleton>
                            <Skeleton className="h-4 w-full"></Skeleton>
                          </div>
                
                      </SidebarMenuSubItem>
                    ) : (
                      <>
                        {projects?.data?.projects?.map((project) => (
                          <Link to={`/dashboard/project/${project.name}`} key={project.id}>
                            <SidebarMenuSubItem key={project.id}>
                              <SidebarMenuSubButton className={"cursor-pointer"}>{project.name}</SidebarMenuSubButton>
                            </SidebarMenuSubItem>
                          </Link>
                        ))}
                      </>
                    )}
                  </SidebarMenuSub>
                </CollapsibleContent>
              </SidebarMenuItem>
            </Collapsible>
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarGroup>
          <NavUser user={{ email: userEmail, avatar: "", name: userName }} handleLogout={handleLogout} />
        </SidebarGroup>
      </SidebarFooter>
    </Sidebar>
  );
}

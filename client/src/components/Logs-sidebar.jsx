import {
  GalleryVerticalEnd,
  Tv,
  HomeIcon,
  Cog,
  NotepadText,
  ChartNoAxesCombined,
  TriangleAlert,
  ActivityIcon,
  RefrigeratorIcon,
  AlertTriangleIcon,
} from "lucide-react";
import { Sidebar, SidebarContent, SidebarFooter, SidebarGroup, SidebarHeader, SidebarMenu, SidebarMenuItem, SidebarMenuButton } from "./ui/sidebar";

import { useLocation, useParams, useMatch, Link } from "react-router-dom";

export function LogsSidebar({ ...props }) {
  const { projectName } = useParams();
  const location = useLocation();
  const match = useMatch("/dashboard/project/:projectName");
  return (
    <Sidebar {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" asChild>
              <a href="#">
                <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                  <GalleryVerticalEnd className="size-4" />
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
          <Link to={`/dashboard/projects`}>
            <SidebarMenuButton size="icon" className={"py-2"}>
              <div className="flex items-center ">
                <HomeIcon />
                <span className="ml-2 text-base">Home</span>
              </div>
            </SidebarMenuButton>
          </Link>
        </SidebarGroup>

        <SidebarGroup className={"projects-group"}>
          <Link to={`/dashboard/project/${projectName}/live-watch`}>
            <SidebarMenuButton size="icon" className={"py-2 cursor-pointer"} isActive={location.pathname.includes("live-watch")}>
              <div className="flex items-center ">
                <Tv />
                <span className="ml-2 text-base">Live Watch </span>
              </div>
            </SidebarMenuButton>
          </Link>

          <Link to={`/dashboard/project/${projectName}`}>
            <SidebarMenuButton size="icon" className={"py-2 mt-2 cursor-pointer"} isActive={match !== null}>
              <div className="flex items-center ">
                <NotepadText />
                <span className="ml-2 text-base">Logs</span>
              </div>
            </SidebarMenuButton>
          </Link>

          <Link to={`/dashboard/project/${projectName}/metrics`}>
            <SidebarMenuButton size="icon" className={"py-2 mt-2 cursor-pointer"} isActive={location.pathname.includes("metrics")}>
              <div className="flex items-center ">
                <ActivityIcon />
                <span className="ml-2 text-base">Metrics</span>
              </div>
            </SidebarMenuButton>
          </Link>

          <Link to={`/dashboard/project/${projectName}/alert`}>
            <SidebarMenuButton size="icon" className={"py-2 mt-2 cursor-pointer"} isActive={location.pathname.includes("alert")}>
              <div className="flex items-center ">
                <AlertTriangleIcon />
                <span className="ml-2 text-base">Alert</span>
              </div>
            </SidebarMenuButton>
          </Link>

          <Link to={`/dashboard/project/${projectName}/statistics`}>
            <SidebarMenuButton size="icon" className={"py-2 mt-2 cursor-pointer"} isActive={location.pathname.includes("statistics")}>
              <div className="flex items-center ">
                <ChartNoAxesCombined />
                <span className="ml-2 text-base">Statistics</span>
              </div>
            </SidebarMenuButton>
          </Link>
          <Link to={`/dashboard/project/${projectName}/cold_storage`}>
            <SidebarMenuButton size="icon" className={"py-2 mt-2 cursor-pointer"} isActive={location.pathname.includes("cold_storage")}>
              <div className="flex items-center ">
                <RefrigeratorIcon />
                <span className="ml-2 text-base">Cold Storage</span>
              </div>
            </SidebarMenuButton>
          </Link>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <SidebarGroup>
          <Link to={`/dashboard/project/${projectName}/settings`}>
            <SidebarMenuButton
              size="icon"
              className={"py-2 cursor-pointer"}
              isActive={location.pathname.startsWith("/dashboard/project/") && location.pathname.includes("settings")}
            >
              <div className="flex items-center ">
                <Cog />
                <span className="ml-2 text-base">Settings</span>
              </div>
            </SidebarMenuButton>
          </Link>
        </SidebarGroup>
      </SidebarFooter>
    </Sidebar>
  );
}

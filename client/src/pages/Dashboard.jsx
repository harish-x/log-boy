import React, { useEffect } from "react";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { ProjectSidebar } from "@/components/Project-sidebar";
import { Outlet } from "react-router-dom";
import { Separator } from "@radix-ui/react-dropdown-menu";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { ModeToggle } from "@/components/mode-toggle";
import { useLocation } from "react-router-dom";
import { LogsSidebar } from "@/components/Logs-sidebar";
import { useNavigate } from "react-router-dom";

export default function Dashboard({ children }) {
  const location = useLocation();

  const routeLabels = {
    "/": "Home",
    "/projects": "Projects",
    "/projects/create": "Create Project",
  };
  const navigate = useNavigate();
  // Redirect to project page if user is on /dashboard
  useEffect(() => {
    if (location.pathname == "/dashboard") {
      navigate("/dashboard/projects");
    }
  }, []);

  // Generate breadcrumbs
  const generateBreadcrumbs = () => {
    const pathnames = location.pathname.split("/").filter((x) => x);

    const breadcrumbs = [{ label: "Home", href: "/", isLast: pathnames.length === 0 }];

    let currentPath = "";
    pathnames.forEach((pathname, index) => {
      currentPath += `/${pathname}`;
      const isLast = index === pathnames.length - 1;
      const label = routeLabels[currentPath] || decodeURIComponent(pathname.replace("-", " ").replace(/\b\w/g, (l) => l.toUpperCase()));

      breadcrumbs.push({
        label,
        href: currentPath,
        isLast,
      });
    });

    return breadcrumbs;
  };

  const breadcrumbs = generateBreadcrumbs();
  return (
    <SidebarProvider>
      {/* if location starts with /dashboard/project/ then show logs sidebar else show project sidebar */}
      {location.pathname.startsWith("/dashboard/project/") ? <LogsSidebar /> : <ProjectSidebar />}

      <main className="bg-sidebar flex-1">
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4 w-full">
            <SidebarTrigger className="-ml-1 bg-primary text-primary-foreground" />
            <Separator orientation="vertical" className="mr-2 h-4 bg-border" />
            <Breadcrumb className="">
              <BreadcrumbList>
                {breadcrumbs.map((breadcrumb, index) => (
                  <React.Fragment key={breadcrumb.href}>
                    <BreadcrumbItem className="hidden md:block">
                      {breadcrumb.isLast ? (
                        <BreadcrumbPage>{breadcrumb.label}</BreadcrumbPage>
                      ) : (
                        <BreadcrumbLink
                          href={
                            breadcrumb.href === "/dashboard"
                              ? "/dashboard/projects"
                              : breadcrumb.href === "/dashboard/project"
                              ? "/dashboard/projects"
                              : breadcrumb.href
                          }
                        >
                          {breadcrumb.label}
                        </BreadcrumbLink>
                      )}
                    </BreadcrumbItem>
                    {!breadcrumb.isLast && <BreadcrumbSeparator className="hidden md:block" />}
                  </React.Fragment>
                ))}
              </BreadcrumbList>
            </Breadcrumb>
            <div className="ml-auto flex items-center space-x-4">
              <ModeToggle />
            </div>
          </div>
        </header>
        {children}
        <Outlet />
      </main>
    </SidebarProvider>
  );
}

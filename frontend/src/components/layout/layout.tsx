import { Outlet, useLocation } from "react-router-dom";

import { ThemeToggle } from "@/components/theme-toggle";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "./app-sidebar";
import { getRouteMeta } from "./navigation";

export default function Layout() {
  const currentRoute = getRouteMeta(useLocation().pathname);

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset className="min-h-svh bg-transparent">
        <header className="sticky top-0 z-20 border-b bg-background/90 backdrop-blur-xl">
          <div className="flex h-14 items-center gap-3 px-4 md:px-5">
            <SidebarTrigger />
            <Separator orientation="vertical" className="hidden h-4 md:block" />
            <Breadcrumb>
              <BreadcrumbList>
                <BreadcrumbItem className="hidden text-muted-foreground sm:block">
                  {currentRoute.section}
                </BreadcrumbItem>
                <BreadcrumbSeparator className="hidden sm:block" />
                <BreadcrumbItem>
                  <BreadcrumbPage>{currentRoute.title}</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
            <div className="ml-auto flex items-center gap-2">
              <ThemeToggle />
            </div>
          </div>
        </header>
        <div className="flex flex-1 flex-col p-4 md:p-5">
          <div className="mx-auto w-full max-w-7xl">
            <Outlet />
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

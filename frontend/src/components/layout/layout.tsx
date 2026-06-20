import { Bell, Command, Search, Sparkles } from "lucide-react";
import { Outlet, useLocation } from "react-router-dom";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Input } from "@/components/ui/input";
import { SidebarInset, SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import { ThemeToggle } from "@/components/theme-toggle";
import { AppSidebar } from "./app-sidebar";
import { getRouteMeta } from "./navigation";

export default function Layout() {
  const location = useLocation();
  const currentRoute = getRouteMeta(location.pathname);

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset className="min-h-svh bg-transparent">
        <header className="sticky top-0 z-20 border-b border-border/60 bg-background/80 backdrop-blur-xl">
          <div className="flex h-[4.5rem] items-center gap-3 px-4 md:px-6">
            <SidebarTrigger className="-ml-1 rounded-xl" />
            <Separator orientation="vertical" className="mr-1 hidden h-4 md:block" />
            <div className="min-w-0">
              <Breadcrumb>
                <BreadcrumbList>
                  <BreadcrumbItem className="text-muted-foreground">
                    {currentRoute.section}
                  </BreadcrumbItem>
                  <BreadcrumbSeparator />
                  <BreadcrumbItem>
                    <BreadcrumbPage>{currentRoute.title}</BreadcrumbPage>
                  </BreadcrumbItem>
                </BreadcrumbList>
              </Breadcrumb>
              <div className="mt-1 hidden items-center gap-2 text-xs text-muted-foreground md:flex">
                <Sparkles className="size-3.5" />
                <span>灵感参考自 shadcn-admin 的轻后台布局</span>
              </div>
            </div>
            <div className="ml-auto hidden max-w-sm flex-1 items-center md:flex">
              <div className="relative w-full">
                <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  aria-label="全局搜索"
                  placeholder="Search dashboards, components, users..."
                  className="rounded-2xl border-border/60 bg-card/80 pl-9 shadow-sm"
                />
              </div>
            </div>
            <Button
              variant="outline"
              className="hidden rounded-2xl border-border/60 lg:inline-flex"
            >
              <Command />
              快速命令
            </Button>
            <Button variant="ghost" size="icon" className="rounded-xl">
              <Bell />
              <span className="sr-only">通知中心</span>
            </Button>
            <ThemeToggle />
            <Avatar className="size-10 rounded-2xl border border-border/60 shadow-sm">
              <AvatarFallback className="rounded-2xl bg-primary/10 text-primary">LH</AvatarFallback>
            </Avatar>
          </div>
        </header>
        <div className="flex flex-1 flex-col px-4 py-6 md:px-6">
          <Outlet />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

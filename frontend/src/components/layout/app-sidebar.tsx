import { Blocks, ChevronsUpDown, Search } from "lucide-react";
import { NavLink, useLocation } from "react-router-dom";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInput,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { navigationGroups } from "./navigation";

export function AppSidebar() {
  const location = useLocation();

  return (
    <Sidebar
      collapsible="icon"
      className="sticky top-0 h-svh border-r border-sidebar-border/70 bg-sidebar/95"
    >
      <SidebarHeader className="gap-3 px-3 py-4">
        <div className="flex items-center gap-3 rounded-2xl border border-sidebar-border/60 bg-background/80 px-3 py-3 shadow-sm">
          <div className="flex size-10 items-center justify-center rounded-2xl bg-primary text-primary-foreground shadow-sm">
            <Blocks className="size-5" />
          </div>
          <div className="min-w-0 group-data-[collapsible=icon]:hidden">
            <p className="truncate text-sm font-semibold text-sidebar-foreground">Shadcn Admin</p>
            <p className="truncate text-xs text-muted-foreground">React Starter Demo</p>
          </div>
        </div>
        <div className="relative group-data-[collapsible=icon]:hidden">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <SidebarInput
            aria-label="搜索组件"
            placeholder="搜索页面..."
            className="rounded-xl border-sidebar-border/60 bg-background/80 pl-9"
          />
        </div>
      </SidebarHeader>
      <SidebarSeparator />
      <SidebarContent className="px-2 py-3">
        {navigationGroups.map((group) => (
          <SidebarGroup key={group.label}>
            <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {group.items.map((item) => {
                  const isActive =
                    item.url === "/"
                      ? location.pathname === item.url
                      : location.pathname.startsWith(item.url);

                  return (
                    <SidebarMenuItem key={item.url}>
                      <SidebarMenuButton asChild isActive={isActive} tooltip={item.title}>
                        <NavLink to={item.url} end={item.url === "/"}>
                          <item.icon />
                          <span>{item.title}</span>
                        </NavLink>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  );
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
      <SidebarSeparator />
      <SidebarFooter className="p-3">
        <div className="flex items-center gap-3 rounded-2xl border border-sidebar-border/60 bg-background/80 px-3 py-3 shadow-sm">
          <Avatar className="size-10 rounded-xl">
            <AvatarFallback className="rounded-xl bg-primary/10 text-primary">LH</AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1 group-data-[collapsible=icon]:hidden">
            <p className="truncate text-sm font-semibold text-sidebar-foreground">Liyuhang</p>
            <p className="truncate text-xs text-muted-foreground">Design Ops Workspace</p>
          </div>
          <ChevronsUpDown className="size-4 text-muted-foreground group-data-[collapsible=icon]:hidden" />
        </div>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}

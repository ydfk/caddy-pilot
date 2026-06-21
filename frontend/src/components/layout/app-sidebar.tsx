import { LogOut } from "lucide-react";
import { NavLink, useLocation, useNavigate } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { BrandLogo } from "@/components/brand-logo";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { useAuthStore } from "@/store/auth-store";
import { navigationItems } from "./navigation";

export function AppSidebar() {
  const location = useLocation();
  const navigate = useNavigate();
  const clear = useAuthStore((state) => state.clear);

  return (
    <Sidebar collapsible="icon" className="border-r border-sidebar-border/70">
      <SidebarHeader className="p-3">
        <div className="flex items-center gap-3 rounded-lg px-2 py-2">
          <BrandLogo className="size-9 shrink-0 shadow-sm shadow-primary/15" eager />
          <div className="min-w-0 group-data-[collapsible=icon]:hidden">
            <p className="truncate font-mono text-sm font-semibold tracking-[0.12em]">CADDYPILOT</p>
          </div>
        </div>
      </SidebarHeader>
      <SidebarSeparator />
      <SidebarContent className="py-3">
        <SidebarGroup>
          <SidebarGroupLabel>管理</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {navigationItems.map((item) => (
                <SidebarMenuItem key={item.url}>
                  <SidebarMenuButton
                    asChild
                    isActive={location.pathname.startsWith(item.url)}
                    tooltip={item.title}
                  >
                    <NavLink to={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </NavLink>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarSeparator />
      <SidebarFooter className="p-3">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => {
            clear();
            navigate("/login", { replace: true });
          }}
        >
          <LogOut data-icon="inline-start" />
          <span className="group-data-[collapsible=icon]:hidden">退出登录</span>
        </Button>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}

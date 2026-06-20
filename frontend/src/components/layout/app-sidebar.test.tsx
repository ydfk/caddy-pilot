import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { SidebarProvider } from "@/components/ui/sidebar";
import { AppSidebar } from "./app-sidebar";

test("展示 CaddyPilot 管理导航", () => {
  render(
    <MemoryRouter>
      <SidebarProvider>
        <AppSidebar />
      </SidebarProvider>
    </MemoryRouter>
  );

  expect(screen.getByText("CADDYPILOT")).toBeInTheDocument();
  for (const item of ["仪表盘", "代理站点", "配置版本", "Caddy 状态", "系统设置"]) {
    expect(screen.getByText(item)).toBeInTheDocument();
  }
});

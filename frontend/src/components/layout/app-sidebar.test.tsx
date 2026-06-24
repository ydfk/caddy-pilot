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
  for (const item of [
    "仪表盘",
    "代理站点",
    "Caddy",
    "证书与访问",
    "证书",
    "DNS Provider",
    "密码本",
    "日志",
  ]) {
    expect(screen.getByText(item)).toBeInTheDocument();
  }
  expect(screen.queryByText("系统设置")).not.toBeInTheDocument();
  expect(screen.queryByText("运行状态")).not.toBeInTheDocument();
  expect(screen.queryByText("配置版本")).not.toBeInTheDocument();
  const password = screen.getByText("密码本");
  const logs = screen.getByText("日志");
  expect(password.compareDocumentPosition(logs) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
});

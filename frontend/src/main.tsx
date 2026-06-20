import { lazy, StrictMode, Suspense } from "react";
import { createRoot } from "react-dom/client";
import { Navigate, RouterProvider, createBrowserRouter } from "react-router-dom";

import { ProtectedRoute } from "@/components/auth/protected-route";
import Layout from "@/components/layout/layout";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { Spinner } from "@/components/ui/spinner";
import App from "./App";
import "./index.css";

const LoginPage = lazy(() => import("./pages/login"));
const PlaceholderPage = lazy(() => import("./pages/placeholder"));
const ProxySitesPage = lazy(() => import("./pages/proxy-sites"));
const ProxySiteFormPage = lazy(() => import("./pages/proxy-sites/form"));

const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    element: <ProtectedRoute />,
    children: [
      {
        path: "/",
        element: <Layout />,
        children: [
          { index: true, element: <App /> },
          {
            path: "dashboard",
            element: <PlaceholderPage title="仪表盘" description="站点与 Caddy 运行状态总览。" />,
          },
          {
            path: "proxy-sites",
            element: <ProxySitesPage />,
          },
          {
            path: "proxy-sites/new",
            element: <ProxySiteFormPage />,
          },
          {
            path: "proxy-sites/:id/edit",
            element: <ProxySiteFormPage />,
          },
          {
            path: "proxy-sites/:id/clone",
            element: <ProxySiteFormPage />,
          },
          {
            path: "config-versions",
            element: <PlaceholderPage title="配置版本" description="查看发布历史与执行回滚。" />,
          },
          {
            path: "config-versions/:id",
            element: <PlaceholderPage title="版本详情" description="检查业务配置与 Caddy JSON。" />,
          },
          {
            path: "caddy",
            element: (
              <PlaceholderPage title="Caddy 状态" description="检查连接、预览并发布配置。" />
            ),
          },
          {
            path: "settings",
            element: <PlaceholderPage title="系统设置" description="查看本地节点与部署参数。" />,
          },
        ],
      },
    ],
  },
  { path: "*", element: <Navigate to="/dashboard" replace /> },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ThemeProvider>
      <Suspense
        fallback={
          <div className="flex min-h-svh items-center justify-center gap-2 text-sm text-muted-foreground">
            <Spinner /> 正在加载
          </div>
        }
      >
        <RouterProvider router={router} />
      </Suspense>
      <Toaster />
    </ThemeProvider>
  </StrictMode>
);

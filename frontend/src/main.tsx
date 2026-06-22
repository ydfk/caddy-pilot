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
const ProxySitesPage = lazy(() => import("./pages/proxy-sites"));
const ProxySiteFormPage = lazy(() => import("./pages/proxy-sites/form"));
const ConfigVersionDetailPage = lazy(() => import("./pages/config-versions/detail"));
const DashboardPage = lazy(() => import("./pages/dashboard"));
const CaddyPage = lazy(() => import("./pages/caddy"));
const BasicAuthPage = lazy(() => import("./pages/basic-auth"));
const CertificatesPage = lazy(() => import("./pages/certificates"));
const DNSProvidersPage = lazy(() => import("./pages/dns-providers"));
const LogsPage = lazy(() => import("./pages/logs"));

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
            element: <DashboardPage />,
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
            path: "certificates",
            element: <CertificatesPage />,
          },
          {
            path: "dns-providers",
            element: <DNSProvidersPage />,
          },
          {
            path: "basic-auth",
            element: <BasicAuthPage />,
          },
          {
            path: "config-versions",
            element: <Navigate to="/caddy" replace />,
          },
          {
            path: "config-versions/:id",
            element: <ConfigVersionDetailPage />,
          },
          {
            path: "caddy",
            element: <CaddyPage />,
          },
          {
            path: "logs",
            element: <LogsPage />,
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

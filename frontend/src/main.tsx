/*
 * @Description: Copyright (c) ydfk. All rights reserved
 * @Author: ydfk
 * @Date: 2025-03-11 10:49:14
 * @LastEditors: ydfk
 * @LastEditTime: 2025-03-11 11:00:10
 */
import { lazy, StrictMode, Suspense } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider, createBrowserRouter } from "react-router-dom";

import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";

import "./index.css";
import App from "./App";
import Unauthorized from "./components/error/unauthorized";
import Layout from "./components/layout/layout";

const ButtonDemo = lazy(() => import("./pages/components/button-demo"));
const CalendarDemo = lazy(() => import("./pages/components/calendar-demo"));
const CardDemo = lazy(() => import("./pages/components/card-demo"));
const CheckboxDemo = lazy(() => import("./pages/components/checkbox-demo"));
const DialogDemo = lazy(() => import("./pages/components/dialog-demo"));
const DropdownMenuDemo = lazy(() => import("./pages/components/dropdown-demo"));
const FormDemo = lazy(() => import("./pages/components/form-demo"));
const InputDemo = lazy(() => import("./pages/components/input-demo"));
const MagicUiDemo = lazy(() => import("./pages/components/magicui-demo"));
const ProgressDemo = lazy(() => import("./pages/components/progress-demo"));
const RadioGroupDemo = lazy(() => import("./pages/components/radio-group-demo"));
const SelectDemo = lazy(() => import("./pages/components/select-demo"));
const SliderDemo = lazy(() => import("./pages/components/slider-demo"));
const SonnerDemo = lazy(() => import("./pages/components/sonner-demo"));
const SwitchDemo = lazy(() => import("./pages/components/switch-demo"));
const TableDemo = lazy(() => import("./pages/components/table-demo"));
const TabsDemo = lazy(() => import("./pages/components/tabs-demo"));
const DashboardDemo = lazy(() => import("./pages/dashboard-demo"));
const UserManagement = lazy(() => import("./pages/user-management"));

// 创建路由
const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    children: [
      {
        index: true,
        element: <DashboardDemo />,
      },
      {
        path: "user-management",
        element: <UserManagement />,
      },
      {
        path: "components/button",
        element: <ButtonDemo />,
      },
      {
        path: "components/input",
        element: <InputDemo />,
      },
      {
        path: "components/checkbox",
        element: <CheckboxDemo />,
      },
      {
        path: "components/radio-group",
        element: <RadioGroupDemo />,
      },
      {
        path: "components/switch",
        element: <SwitchDemo />,
      },
      {
        path: "components/slider",
        element: <SliderDemo />,
      },
      {
        path: "components/tabs",
        element: <TabsDemo />,
      },
      {
        path: "components/progress",
        element: <ProgressDemo />,
      },
      {
        path: "components/card",
        element: <CardDemo />,
      },
      {
        path: "components/dialog",
        element: <DialogDemo />,
      },
      {
        path: "components/form",
        element: <FormDemo />,
      },
      {
        path: "components/table",
        element: <TableDemo />,
      },
      {
        path: "components/dropdown",
        element: <DropdownMenuDemo />,
      },
      {
        path: "components/sonner",
        element: <SonnerDemo />,
      },
      {
        path: "components/select",
        element: <SelectDemo />,
      },
      {
        path: "components/calendar",
        element: <CalendarDemo />,
      },
      {
        path: "components/magicui",
        element: <MagicUiDemo />,
      },
    ],
  },
  {
    path: "/401",
    element: <Unauthorized />,
  },
  {
    path: "/auth/login",
    element: <App />,
  },
]);

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <ThemeProvider>
      <Suspense
        fallback={
          <div className="flex min-h-screen items-center justify-center text-sm text-muted-foreground">
            正在加载页面...
          </div>
        }
      >
        <RouterProvider router={router} />
      </Suspense>
      <Toaster />
    </ThemeProvider>
  </StrictMode>
);

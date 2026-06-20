import type { LucideIcon } from "lucide-react";
import { Activity, FileClock, Gauge, Route, Settings2 } from "lucide-react";

export type NavigationItem = { title: string; url: string; icon: LucideIcon };

export const navigationItems: NavigationItem[] = [
  { title: "仪表盘", url: "/dashboard", icon: Gauge },
  { title: "代理站点", url: "/proxy-sites", icon: Route },
  { title: "配置版本", url: "/config-versions", icon: FileClock },
  { title: "Caddy 状态", url: "/caddy", icon: Activity },
  { title: "系统设置", url: "/settings", icon: Settings2 },
];

export function getRouteMeta(pathname: string) {
  if (pathname === "/proxy-sites/new") return { section: "代理站点", title: "新增站点" };
  if (pathname.endsWith("/clone")) return { section: "代理站点", title: "克隆站点" };
  if (pathname.startsWith("/proxy-sites/")) return { section: "代理站点", title: "编辑站点" };
  if (pathname.startsWith("/config-versions/")) return { section: "配置版本", title: "版本详情" };
  const current = navigationItems.find((item) => pathname.startsWith(item.url));
  return { section: "CaddyPilot", title: current?.title ?? "控制台" };
}

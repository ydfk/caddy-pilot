import type { LucideIcon } from "lucide-react";
import { BadgeCheck, CloudCog, Gauge, KeyRound, Route, ServerCog, ShieldCheck } from "lucide-react";

export type NavigationItem = { title: string; url: string; icon: LucideIcon };

export const primaryNavigation: NavigationItem[] = [
  { title: "仪表盘", url: "/dashboard", icon: Gauge },
  { title: "代理站点", url: "/proxy-sites", icon: Route },
  { title: "Caddy 管理", url: "/caddy", icon: ServerCog },
];

export const navigationGroups = [
  {
    title: "证书与访问",
    icon: ShieldCheck,
    children: [
      { title: "证书", url: "/certificates", icon: BadgeCheck },
      { title: "DNS Provider", url: "/dns-providers", icon: CloudCog },
      { title: "密码本", url: "/basic-auth", icon: KeyRound },
    ],
  },
] satisfies Array<{ title: string; icon: LucideIcon; children: NavigationItem[] }>;

const navigationItems = [
  ...primaryNavigation,
  ...navigationGroups.flatMap((group) => group.children),
];

export function getRouteMeta(pathname: string) {
  if (pathname === "/proxy-sites/new") return { section: "代理站点", title: "新增站点" };
  if (pathname.endsWith("/clone")) return { section: "代理站点", title: "克隆站点" };
  if (pathname.startsWith("/proxy-sites/")) return { section: "代理站点", title: "编辑站点" };
  if (pathname.startsWith("/config-versions/")) return { section: "Caddy 管理", title: "版本详情" };
  const group = navigationGroups.find((item) =>
    item.children.some((child) => pathname.startsWith(child.url))
  );
  const current = navigationItems.find((item) => pathname.startsWith(item.url));
  return { section: group?.title ?? "CaddyPilot", title: current?.title ?? "控制台" };
}

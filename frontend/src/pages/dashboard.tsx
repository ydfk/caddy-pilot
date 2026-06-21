import { useEffect, useState } from "react";
import { Activity, FileClock, Globe2, LockKeyhole, Power, Route } from "lucide-react";
import { Link } from "react-router-dom";
import { toast } from "sonner";

import { getDashboardSummary, type DashboardSummary } from "@/api/dashboard";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const metrics = [
  { key: "site_count", label: "站点总数", icon: Route },
  { key: "enabled_site_count", label: "启用站点", icon: Power },
  { key: "disabled_site_count", label: "停用站点", icon: Activity },
  { key: "https_site_count", label: "HTTPS 站点", icon: LockKeyhole },
] as const;

export default function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);

  useEffect(() => {
    getDashboardSummary()
      .then(setSummary)
      .catch((error) => toast.error(error instanceof Error ? error.message : "读取仪表盘失败"));
  }, []);

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow="CONTROL / OVERVIEW"
        title="仪表盘"
        description="本地代理站点与配置发布状态的即时概览。"
      />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {metrics.map((metric) => (
          <Card key={metric.key}>
            <CardHeader className="flex-row items-center justify-between">
              <CardDescription>{metric.label}</CardDescription>
              <metric.icon className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {summary ? (
                <p className="font-mono text-3xl font-semibold tracking-tight">
                  {summary[metric.key]}
                </p>
              ) : (
                <Skeleton className="h-9 w-20" />
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe2 className="size-5" /> 本地 Caddy
            </CardTitle>
            <CardDescription>Admin API 仅用于容器内部通信</CardDescription>
          </CardHeader>
          <CardContent className="flex items-center justify-between gap-4">
            <div>
              <p className="font-mono text-sm">{summary?.caddy_admin_api ?? "正在读取…"}</p>
              <p className="mt-1 text-xs text-muted-foreground">固定管理入口 :8080</p>
            </div>
            <Badge variant={summary?.caddy_online ? "default" : "destructive"}>
              {summary?.caddy_online ? "在线" : "离线"}
            </Badge>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileClock className="size-5" /> 最近发布
            </CardTitle>
            <CardDescription>最近一次成功发布或回滚</CardDescription>
          </CardHeader>
          <CardContent className="flex items-center justify-between gap-4">
            <p className="font-mono text-sm">
              {summary?.last_publish_time
                ? new Intl.DateTimeFormat("zh-CN", {
                    dateStyle: "medium",
                    timeStyle: "short",
                  }).format(new Date(summary.last_publish_time))
                : "暂无成功发布"}
            </p>
            <Button variant="outline" size="sm" asChild>
              <Link to="/config-versions">查看历史</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

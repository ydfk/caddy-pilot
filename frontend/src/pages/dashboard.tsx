import { useEffect, useState } from "react";
import {
  Activity,
  ArrowUpDown,
  FileClock,
  Globe2,
  LockKeyhole,
  Power,
  Route,
  ServerCrash,
} from "lucide-react";
import { Link } from "react-router-dom";
import { toast } from "sonner";

import { getDashboardSummary, type DashboardSummary } from "@/api/dashboard";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const metrics = [
  { key: "site_count", label: "站点总数", icon: Route, format: formatNumber },
  { key: "enabled_site_count", label: "启用站点", icon: Power, format: formatNumber },
  { key: "https_site_count", label: "HTTPS 站点", icon: LockKeyhole, format: formatNumber },
  { key: "request_count_24h", label: "24h 访问量", icon: Activity, format: formatNumber },
  { key: "error_count_24h", label: "24h 5xx", icon: ServerCrash, format: formatNumber },
  { key: "traffic_bytes_24h", label: "24h 出站流量", icon: ArrowUpDown, format: formatBytes },
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
        description="代理站点与配置发布状态的即时概览。"
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        {metrics.map((metric) => (
          <Card key={metric.key}>
            <CardHeader className="flex-row items-center justify-between">
              <CardDescription>{metric.label}</CardDescription>
              <metric.icon className="size-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {summary ? (
                <p className="font-mono text-3xl font-semibold tracking-tight">
                  {metric.format(summary[metric.key])}
                </p>
              ) : (
                <Skeleton className="h-9 w-20" />
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card className="lg:col-span-2">
          <CardHeader className="flex-row items-start justify-between gap-4">
            <div>
              <CardTitle>访问最多的站点</CardTitle>
              <CardDescription>最近 24 小时，数据来自已开启的站点访问日志</CardDescription>
            </div>
            <Badge variant={summary?.error_count_24h ? "destructive" : "outline"}>
              <ServerCrash className="mr-1 size-3" /> {summary?.error_count_24h ?? 0} 次 5xx
            </Badge>
          </CardHeader>
          <CardContent>
            {summary?.top_sites_24h.length ? (
              <div className="grid gap-3">
                {summary.top_sites_24h.map((site) => {
                  const maximum = summary.top_sites_24h[0]?.request_count || 1;
                  return (
                    <div key={site.domain} className="grid gap-1.5">
                      <div className="flex items-center justify-between gap-4 text-sm">
                        <span className="truncate font-mono text-xs">{site.domain}</span>
                        <span className="shrink-0 text-xs text-muted-foreground">
                          {formatNumber(site.request_count)} 次 · {formatBytes(site.bytes)}
                        </span>
                      </div>
                      <div className="h-1.5 overflow-hidden rounded-full bg-muted">
                        <div
                          className="h-full rounded-full bg-primary transition-[width]"
                          style={{ width: `${Math.max(3, (site.request_count / maximum) * 100)}%` }}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <p className="py-6 text-center text-sm text-muted-foreground">
                暂无访问数据，发布开启访问日志的站点后开始统计。
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe2 className="size-5" /> Caddy 运行状态
            </CardTitle>
            <CardDescription>由 CaddyPilot 后端统一托管</CardDescription>
          </CardHeader>
          <CardContent className="flex items-center justify-between gap-4">
            <div>
              <p className="text-sm font-medium">
                {summary?.caddy_online ? "服务运行正常" : "服务暂不可用"}
              </p>
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
              <Link to="/caddy">查看历史</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("zh-CN", { notation: "compact", maximumFractionDigits: 1 }).format(
    value
  );
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let size = value / 1024;
  let unit = units[0];
  for (let index = 1; index < units.length && size >= 1024; index += 1) {
    size /= 1024;
    unit = units[index];
  }
  return `${size.toFixed(size >= 10 ? 0 : 1)} ${unit}`;
}

import { Fragment } from "react";
import { Database, HardDrive, Network, ShieldAlert } from "lucide-react";

import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";

const settings = [
  { icon: Network, label: "管理界面", value: ":8080" },
  { icon: Network, label: "后端监听", value: "127.0.0.1:25610" },
  { icon: ShieldAlert, label: "Caddy Admin API", value: "http://127.0.0.1:2019" },
  { icon: Database, label: "SQLite", value: "/data/caddypilot.db" },
  { icon: HardDrive, label: "前端目录", value: "/app/frontend" },
];

export default function SettingsPage() {
  return (
    <div className="flex max-w-4xl flex-col gap-4">
      <PageHeader
        eyebrow="SYSTEM / LOCAL"
        title="系统设置"
        description="MVP 采用单节点、单用户和环境变量配置；此页面只读。"
      />
      <Alert>
        <ShieldAlert />
        <AlertTitle>Admin API 安全边界</AlertTitle>
        <AlertDescription>
          不要将 Caddy Admin API 2019 端口暴露到公网或映射到宿主机。
        </AlertDescription>
      </Alert>
      <Card>
        <CardHeader>
          <CardTitle>运行参数</CardTitle>
          <CardDescription>单镜像容器内的固定服务地址</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col">
          {settings.map((setting, index) => (
            <Fragment key={setting.label}>
              {index > 0 ? <Separator /> : null}
              <div className="flex items-center gap-4 py-4 first:pt-0 last:pb-0">
                <setting.icon className="size-5 text-muted-foreground" />
                <p className="flex-1 text-sm font-medium">{setting.label}</p>
                <code className="break-all text-right font-mono text-xs text-muted-foreground">
                  {setting.value}
                </code>
              </div>
            </Fragment>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}

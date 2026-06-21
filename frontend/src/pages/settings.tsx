import { Database, PackageCheck } from "lucide-react";

import { PageHeader } from "@/components/page-header";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function SettingsPage() {
  return (
    <div className="flex max-w-4xl flex-col gap-4">
      <PageHeader
        eyebrow="SYSTEM / SETTINGS"
        title="系统设置"
        description="CaddyPilot 运行方式与数据持久化说明。"
      />
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2"><PackageCheck className="size-4" /> 托管运行</CardTitle>
          <CardDescription>Caddy 与管理服务作为一个系统运行</CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          后端负责准备、启动、检查、更新和关闭 Caddy，无需额外安装或维护进程。
        </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2"><Database className="size-4" /> 数据持久化</CardTitle>
            <CardDescription>站点、密码本和配置版本统一保存</CardDescription>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            请持续备份数据目录；重新部署时挂载原数据即可恢复管理状态。
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

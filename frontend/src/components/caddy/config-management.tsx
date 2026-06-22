import { useState } from "react";
import { AlertTriangle, RotateCcw } from "lucide-react";
import { toast } from "sonner";

import type { CaddyChangeStatus } from "@/api/caddy";
import { rollbackConfigVersion } from "@/api/config-versions";
import { ConfigHistory } from "@/components/caddy/config-history";
import { DeploymentPipeline } from "@/components/caddy/deployment-pipeline";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";

export function ConfigManagement({
  status,
  refreshKey,
  onChanged,
}: {
  status: CaddyChangeStatus;
  refreshKey: number;
  onChanged: () => Promise<void>;
}) {
  const [repairing, setRepairing] = useState(false);

  async function repair() {
    if (!status.latest_version_id) return;
    setRepairing(true);
    try {
      const version = await rollbackConfigVersion(status.latest_version_id);
      toast.success(`已重新应用配置 v${version.version}`);
      await onChanged();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "重新应用配置失败");
    } finally {
      setRepairing(false);
    }
  }

  const drifted = status.state === "runtime_drift";
  return (
    <Card className="overflow-hidden border-primary/20">
      <CardHeader className="border-b bg-primary/[0.035]">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <CardTitle>配置管理</CardTitle>
            <CardDescription>发布流程、运行一致性和历史版本集中管理。</CardDescription>
          </div>
          <Badge variant={drifted ? "destructive" : status.dirty ? "secondary" : "outline"}>
            {statusLabel(status)}
          </Badge>
        </div>
      </CardHeader>
      {drifted ? (
        <CardContent className="border-b pt-5">
          <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-amber-500/30 bg-amber-500/5 p-3">
            <div className="flex gap-2 text-sm">
              <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600" />
              <div>
                <p className="font-medium">Caddy 当前 JSON 与持久化版本不一致</p>
                <p className="text-xs text-muted-foreground">
                  重新应用最近成功版本可恢复受管状态。
                </p>
              </div>
            </div>
            <Button
              size="sm"
              onClick={() => void repair()}
              disabled={repairing || !status.latest_version_id}
            >
              {repairing ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <RotateCcw data-icon="inline-start" />
              )}
              重新应用最新版本
            </Button>
          </div>
        </CardContent>
      ) : null}
      <DeploymentPipeline
        embedded
        dirty={status.dirty}
        latestVersion={status.latest_version}
        onPublished={onChanged}
      />
      <ConfigHistory
        embedded
        refreshKey={refreshKey}
        activeVersion={status.active_version}
        onRollback={onChanged}
      />
    </Card>
  );
}

function statusLabel(status: CaddyChangeStatus) {
  switch (status.state) {
    case "runtime_drift":
      return "运行配置漂移";
    case "offline":
      return "Caddy 离线";
    case "unpublished_changes":
      return "存在未发布修改";
    case "in_sync":
      return status.active_version ? `运行 v${status.active_version}` : "运行一致";
    default:
      return "暂无发布";
  }
}

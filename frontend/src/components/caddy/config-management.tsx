import { useState } from "react";
import { AlertTriangle, Braces, Download, FileText, RotateCcw } from "lucide-react";
import { toast } from "sonner";

import { getCurrentCaddyConfig, previewCaddyfile, type CaddyChangeStatus } from "@/api/caddy";
import { rollbackConfigVersion } from "@/api/config-versions";
import { ConfigHistory } from "@/components/caddy/config-history";
import { DeploymentPipeline } from "@/components/caddy/deployment-pipeline";
import { JSONDialog } from "@/components/json-dialog";
import { TextPreviewDialog } from "@/components/text-preview-dialog";
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
  const [currentJSON, setCurrentJSON] = useState<unknown>(null);
  const [caddyfile, setCaddyfile] = useState("");

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

  async function loadCurrentJSON(show = true) {
    try {
      const value = (await getCurrentCaddyConfig()).caddy_json;
      if (show) setCurrentJSON(value);
      return value;
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取当前 JSON 失败");
      return null;
    }
  }

  async function loadCaddyfile(show = true) {
    try {
      const response = await previewCaddyfile();
      if (show) setCaddyfile(response.caddyfile);
      return response.caddyfile;
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "生成 Caddyfile 失败");
      return "";
    }
  }

  async function exportJSON() {
    const value = await loadCurrentJSON(false);
    if (value !== null)
      downloadText("caddy-current.json", JSON.stringify(value, null, 2), "application/json");
  }

  async function exportCaddyfile() {
    const value = await loadCaddyfile(false);
    if (value) downloadText("Caddyfile", value, "text/plain");
  }

  const drifted = status.state === "runtime_drift";
  return (
    <Card className="overflow-hidden border-primary/20">
      <CardHeader className="border-b bg-primary/[0.035]">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <CardTitle>配置管理</CardTitle>
            <CardDescription>同步状态、校验发布、只读预览与版本记录。</CardDescription>
          </div>
          <div className="flex flex-wrap items-center justify-end gap-2">
            <Badge variant={drifted ? "destructive" : status.dirty ? "secondary" : "outline"}>
              {statusLabel(status)}
            </Badge>
            <Button variant="outline" size="sm" onClick={() => void loadCurrentJSON()}>
              <Braces /> 当前 JSON
            </Button>
            <Button variant="outline" size="sm" onClick={() => void loadCaddyfile()}>
              <FileText /> 预览 Caddyfile
            </Button>
            <Button variant="ghost" size="sm" onClick={() => void exportJSON()}>
              <Download /> 导出 JSON
            </Button>
            <Button variant="ghost" size="sm" onClick={() => void exportCaddyfile()}>
              <Download /> 导出 Caddyfile
            </Button>
          </div>
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
              {repairing ? <Spinner /> : <RotateCcw />}
              重新应用最新版本
            </Button>
          </div>
        </CardContent>
      ) : null}
      <DeploymentPipeline dirty={status.dirty} onPublished={onChanged} />
      <ConfigHistory
        refreshKey={refreshKey}
        activeVersion={status.active_version}
        onRollback={onChanged}
      />
      <JSONDialog
        open={currentJSON !== null}
        onOpenChange={(open) => !open && setCurrentJSON(null)}
        title="当前运行 JSON"
        description="从托管 Caddy 实时读取；发布、回滚和一致性校验均以 JSON 为准。"
        value={currentJSON}
      />
      <TextPreviewDialog
        open={Boolean(caddyfile)}
        onOpenChange={(open) => !open && setCaddyfile("")}
        title="Caddyfile 阅读视图"
        description="由同一站点模型完整生成，并通过托管 Caddy adapt 回 JSON 校验；不支持编辑导入。"
        value={caddyfile}
        filename="Caddyfile"
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

function downloadText(filename: string, content: string, type: string) {
  const url = URL.createObjectURL(new Blob([content], { type: `${type};charset=utf-8` }));
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}

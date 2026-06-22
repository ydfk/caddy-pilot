import { useCallback, useEffect, useState } from "react";
import { Braces, Download, ExternalLink, PackageCheck, RefreshCw, Settings2 } from "lucide-react";
import { toast } from "sonner";

import {
  getCaddyChangeStatus,
  getCaddySettings,
  getCaddyStatus,
  getCaddyVersion,
  getCurrentCaddyConfig,
  saveCaddySettings,
  updateManagedCaddy,
  type CaddyChangeStatus,
  type CaddySettings,
  type CaddyStatus,
  type CaddyVersion,
} from "@/api/caddy";
import { CaddySettingsDialog } from "@/components/caddy/caddy-settings-dialog";
import { ConfigHistory } from "@/components/caddy/config-history";
import { DeploymentPipeline } from "@/components/caddy/deployment-pipeline";
import { JSONDialog } from "@/components/json-dialog";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";

export default function CaddyPage() {
  const [status, setStatus] = useState<CaddyStatus | null>(null);
  const [version, setVersion] = useState<CaddyVersion | null>(null);
  const [settings, setSettings] = useState<CaddySettings | null>(null);
  const [changeStatus, setChangeStatus] = useState<CaddyChangeStatus>({ dirty: false });
  const [currentConfig, setCurrentConfig] = useState<unknown>(null);
  const [loading, setLoading] = useState(true);
  const [checkingVersion, setCheckingVersion] = useState(false);
  const [updatingVersion, setUpdatingVersion] = useState(false);
  const [historyKey, setHistoryKey] = useState(0);

  const refresh = useCallback(async () => {
    setLoading(true);
    const results = await Promise.allSettled([
      getCaddyStatus(),
      getCaddyVersion(),
      getCaddySettings(),
      getCaddyChangeStatus(),
    ]);
    if (results[0].status === "fulfilled") setStatus(results[0].value);
    if (results[1].status === "fulfilled") setVersion(results[1].value);
    if (results[2].status === "fulfilled") setSettings(results[2].value);
    if (results[3].status === "fulfilled") setChangeStatus(results[3].value);
    const failure = results.find((result) => result.status === "rejected");
    if (failure?.status === "rejected")
      toast.error(
        failure.reason instanceof Error ? failure.reason.message : "读取 Caddy 工作台失败"
      );
    setLoading(false);
  }, []);

  useEffect(() => void refresh(), [refresh]);

  async function showCurrentConfig() {
    try {
      setCurrentConfig((await getCurrentCaddyConfig()).caddy_json);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取当前配置失败");
    }
  }

  async function checkVersion() {
    setCheckingVersion(true);
    try {
      setVersion(await getCaddyVersion());
      toast.success("版本信息已刷新");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "检查版本失败");
    } finally {
      setCheckingVersion(false);
    }
  }

  async function saveSettings(next: CaddySettings) {
    try {
      const saved = await saveCaddySettings(next);
      setSettings(saved);
      setVersion(await getCaddyVersion());
      toast.success("更新源设置已保存");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存更新源失败");
      throw error;
    }
  }

  async function updateVersion() {
    if (!version?.latest_version) return;
    setUpdatingVersion(true);
    try {
      const task = await updateManagedCaddy(version.latest_version);
      toast.info(`正在更新 Caddy 到 ${task.target_version}`);
      for (let attempt = 0; attempt < 30; attempt += 1) {
        await new Promise((resolve) => window.setTimeout(resolve, 1000));
        try {
          const current = await getCaddyVersion();
          if (current.current_version === task.target_version) {
            setVersion(current);
            toast.success(`Caddy 已更新到 ${task.target_version}`);
            await refresh();
            return;
          }
        } catch {
          /* Caddy 重启期间继续等待。 */
        }
      }
      toast.error("等待 Caddy 更新完成超时，请刷新确认状态");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "启动 Caddy 更新失败");
    } finally {
      setUpdatingVersion(false);
    }
  }

  async function refreshAfterConfigChange() {
    await refresh();
    setHistoryKey((current) => current + 1);
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-4">
      <PageHeader
        eyebrow="CADDY / CONTROL DESK"
        title="Caddy 管理"
        description="运行状态、发布流程、更新源与配置版本集中在一个工作台。"
        actions={
          <>
            <Button variant="outline" onClick={() => void showCurrentConfig()}>
              <Braces data-icon="inline-start" />
              当前 JSON
            </Button>
            <Button variant="outline" onClick={() => void refresh()} disabled={loading}>
              {loading ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <RefreshCw data-icon="inline-start" />
              )}
              刷新
            </Button>
          </>
        }
      />

      {status?.error_message ? (
        <Alert variant="destructive">
          <AlertTitle>Caddy 服务不可用</AlertTitle>
          <AlertDescription>{status.error_message}</AlertDescription>
        </Alert>
      ) : null}

      <Card className="overflow-hidden">
        <CardContent className="grid gap-0 p-0 sm:grid-cols-3">
          <Metric
            label="CaddyPilot"
            value={version?.system_version ?? (loading ? "读取中…" : "dev")}
            detail="系统版本"
          />
          <Metric
            label="托管 Caddy"
            value={version?.current_version ?? "不可用"}
            detail={status?.online ? "运行在线" : "运行离线"}
            online={status?.online}
          />
          <Metric
            label="最新稳定版"
            value={version?.latest_version ?? "暂不可用"}
            detail={version?.update_available ? "发现可用更新" : "当前无需更新"}
          />
        </CardContent>
      </Card>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1.35fr)_minmax(20rem,0.65fr)]">
        <DeploymentPipeline
          dirty={changeStatus.dirty}
          latestVersion={changeStatus.latest_version}
          onPublished={refreshAfterConfigChange}
        />

        <Card>
          <CardHeader>
            <div className="flex items-start justify-between gap-3">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <PackageCheck className="size-4" />
                  运行时版本
                </CardTitle>
                <CardDescription>检查和更新内部托管的 Caddy。</CardDescription>
              </div>
              <CaddySettingsDialog
                settings={settings}
                onSave={saveSettings}
                trigger={
                  <Button variant="ghost" size="icon" aria-label="更新源设置">
                    <Settings2 />
                  </Button>
                }
              />
            </div>
          </CardHeader>
          <CardContent className="grid gap-4">
            <div className="rounded-xl border bg-muted/20 p-4">
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm text-muted-foreground">更新状态</span>
                <Badge variant={version?.update_available ? "secondary" : "outline"}>
                  {version?.update_available ? "可更新" : "已是目标版本"}
                </Badge>
              </div>
              {version?.error_message ? (
                <p className="mt-2 text-xs text-muted-foreground">{version.error_message}</p>
              ) : null}
            </div>
            <div className="flex flex-wrap gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => void checkVersion()}
                disabled={checkingVersion}
              >
                {checkingVersion ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <RefreshCw data-icon="inline-start" />
                )}
                检查更新
              </Button>
              {version?.update_available && version.latest_version ? (
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button size="sm" disabled={updatingVersion}>
                      {updatingVersion ? (
                        <Spinner data-icon="inline-start" />
                      ) : (
                        <Download data-icon="inline-start" />
                      )}
                      更新到 {version.latest_version}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>更新托管 Caddy？</AlertDialogTitle>
                      <AlertDialogDescription>
                        系统会下载目标版本、保留当前配置并自动重启，管理入口可能短暂重连。
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>取消</AlertDialogCancel>
                      <AlertDialogAction onClick={() => void updateVersion()}>
                        确认更新
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              ) : null}
              {version?.release_url ? (
                <Button variant="ghost" size="sm" asChild>
                  <a href={version.release_url} target="_blank" rel="noreferrer">
                    <ExternalLink data-icon="inline-start" />
                    发布说明
                  </a>
                </Button>
              ) : null}
            </div>
          </CardContent>
        </Card>
      </div>

      <ConfigHistory refreshKey={historyKey} onRollback={refreshAfterConfigChange} />
      <JSONDialog
        open={currentConfig !== null}
        onOpenChange={(open) => !open && setCurrentConfig(null)}
        title="当前 Caddy JSON"
        description="从托管 Caddy 实时读取。"
        value={currentConfig}
      />
    </div>
  );
}

function Metric({
  label,
  value,
  detail,
  online,
}: {
  label: string;
  value: string;
  detail: string;
  online?: boolean;
}) {
  return (
    <div className="border-b p-5 last:border-b-0 sm:border-r sm:border-b-0 sm:last:border-r-0">
      <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-[0.12em] text-muted-foreground">
        {online !== undefined ? (
          <span className={`size-2 rounded-full ${online ? "bg-emerald-500" : "bg-destructive"}`} />
        ) : null}
        {label}
      </div>
      <p className="mt-2 font-mono text-xl font-semibold tracking-tight">{value}</p>
      <p className="mt-1 text-xs text-muted-foreground">{detail}</p>
    </div>
  );
}

import { useCallback, useEffect, useState } from "react";
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
import { ConfigHistory } from "@/components/caddy/config-history";
import { DeploymentPipeline } from "@/components/caddy/deployment-pipeline";
import { RuntimeCard } from "@/components/caddy/runtime-card";
import { JSONDialog } from "@/components/json-dialog";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

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
      />

      {status?.error_message ? (
        <Alert variant="destructive">
          <AlertTitle>Caddy 服务不可用</AlertTitle>
          <AlertDescription>{status.error_message}</AlertDescription>
        </Alert>
      ) : null}

      <RuntimeCard
        status={status}
        version={version}
        settings={settings}
        loading={loading}
        checkingVersion={checkingVersion}
        updatingVersion={updatingVersion}
        onRefresh={refresh}
        onShowConfig={showCurrentConfig}
        onCheckVersion={checkVersion}
        onUpdateVersion={updateVersion}
        onSaveSettings={saveSettings}
      />

      <div className="grid gap-4">
        <DeploymentPipeline
          dirty={changeStatus.dirty}
          latestVersion={changeStatus.latest_version}
          onPublished={refreshAfterConfigChange}
        />
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

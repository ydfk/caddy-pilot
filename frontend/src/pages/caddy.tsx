import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import {
  getCaddyChangeStatus,
  getCaddySettings,
  getCaddyStatus,
  getCaddyUpdateTask,
  getCaddyVersion,
  saveCaddySettings,
  updateManagedCaddy,
  uploadManagedCaddy,
  type CaddyChangeStatus,
  type CaddySettings,
  type CaddyStatus,
  type CaddyUpdateTask,
  type CaddyVersion,
} from "@/api/caddy";
import { ConfigManagement } from "@/components/caddy/config-management";
import { RuntimeCard } from "@/components/caddy/runtime-card";
import { PageHeader } from "@/components/page-header";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export default function CaddyPage() {
  const [status, setStatus] = useState<CaddyStatus | null>(null);
  const [version, setVersion] = useState<CaddyVersion | null>(null);
  const [settings, setSettings] = useState<CaddySettings | null>(null);
  const [changeStatus, setChangeStatus] = useState<CaddyChangeStatus>({
    dirty: false,
    state: "no_version",
    runtime_in_sync: false,
    persistent_config_in_sync: false,
  });
  const [checkingVersion, setCheckingVersion] = useState(false);
  const [updatingVersion, setUpdatingVersion] = useState(false);
  const [updateTask, setUpdateTask] = useState<CaddyUpdateTask | null>(null);
  const [historyKey, setHistoryKey] = useState(0);

  const refresh = useCallback(async () => {
    const results = await Promise.allSettled([
      getCaddyStatus(),
      getCaddyVersion(),
      getCaddySettings(),
      getCaddyChangeStatus(),
      getCaddyUpdateTask(),
    ]);
    if (results[0].status === "fulfilled") setStatus(results[0].value);
    if (results[1].status === "fulfilled") setVersion(results[1].value);
    if (results[2].status === "fulfilled") setSettings(results[2].value);
    if (results[3].status === "fulfilled") setChangeStatus(results[3].value);
    if (results[4].status === "fulfilled") {
      const task = results[4].value;
      setUpdateTask(task);
      setUpdatingVersion(!["idle", "succeeded", "failed"].includes(task.status));
    }
    const failure = results.find((result) => result.status === "rejected");
    if (failure?.status === "rejected")
      toast.error(
        failure.reason instanceof Error ? failure.reason.message : "读取 Caddy 工作台失败"
      );
  }, []);

  useEffect(() => void refresh(), [refresh]);

  useEffect(() => {
    const timer = window.setInterval(() => {
      getCaddyStatus()
        .then(setStatus)
        .catch(() => undefined);
    }, 10000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!updatingVersion) return;
    const timer = window.setInterval(() => {
      getCaddyUpdateTask()
        .then(async (task) => {
          setUpdateTask(task);
          if (task.status === "succeeded" || task.status === "failed") {
            setUpdatingVersion(false);
            if (task.status === "succeeded") {
              toast.success(`Caddy 已更新到 ${task.target_version}`);
              await refresh();
            } else {
              toast.error(task.error_message || "Caddy 更新失败");
            }
          }
        })
        .catch((error) => toast.error(error instanceof Error ? error.message : "读取更新任务失败"));
    }, 2000);
    return () => window.clearInterval(timer);
  }, [refresh, updatingVersion]);

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
      setUpdateTask({
        status: task.status as CaddyUpdateTask["status"],
        progress: 0,
        target_version: task.target_version,
      });
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "启动 Caddy 更新失败");
      setUpdatingVersion(false);
    }
  }

  async function uploadVersion(file: File) {
    setUpdatingVersion(true);
    try {
      const task = await uploadManagedCaddy(file);
      setUpdateTask({ status: task.status as CaddyUpdateTask["status"], progress: 0 });
      toast.info(`已上传 ${file.name}，正在校验并安装`);
    } catch (error) {
      setUpdatingVersion(false);
      toast.error(error instanceof Error ? error.message : "上传 Caddy 失败");
    }
  }

  async function refreshAfterConfigChange() {
    await refresh();
    setHistoryKey((current) => current + 1);
  }

  return (
    <div className="mx-auto flex w-full max-w-7xl flex-col gap-4">
      <PageHeader title="Caddy 管理" />

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
        checkingVersion={checkingVersion}
        updatingVersion={updatingVersion}
        updateTask={updateTask}
        onCheckVersion={checkVersion}
        onUpdateVersion={updateVersion}
        onUploadVersion={uploadVersion}
        onSaveSettings={saveSettings}
      />

      <ConfigManagement
        status={changeStatus}
        refreshKey={historyKey}
        onChanged={refreshAfterConfigChange}
      />
    </div>
  );
}

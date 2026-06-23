import { Download, ExternalLink, RefreshCw, Settings2 } from "lucide-react";
import type { ReactNode } from "react";

import type { CaddySettings, CaddyStatus, CaddyUpdateTask, CaddyVersion } from "@/api/caddy";
import { CaddySettingsDialog } from "@/components/caddy/caddy-settings-dialog";
import { CaddyUploadButton } from "@/components/caddy/caddy-upload-button";
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
import { Progress } from "@/components/ui/progress";
import { Spinner } from "@/components/ui/spinner";

type Props = {
  status: CaddyStatus | null;
  version: CaddyVersion | null;
  settings: CaddySettings | null;
  checkingVersion: boolean;
  updatingVersion: boolean;
  updateTask: CaddyUpdateTask | null;
  onCheckVersion: () => Promise<void>;
  onUpdateVersion: () => Promise<void>;
  onUploadVersion: (file: File) => Promise<void>;
  onSaveSettings: (settings: CaddySettings) => Promise<void>;
};

export function RuntimeCard({
  status,
  version,
  settings,
  checkingVersion,
  updatingVersion,
  updateTask,
  onCheckVersion,
  onUpdateVersion,
  onUploadVersion,
  onSaveSettings,
}: Props) {
  return (
    <Card className="overflow-hidden border-primary/20">
      <CardHeader className="border-b bg-primary/[0.035]">
        <div className="flex items-start justify-between gap-3">
          <div>
            <CardTitle>Caddy 运行时</CardTitle>
            <CardDescription>版本、安装方式与托管进程状态。</CardDescription>
          </div>
          <CaddySettingsDialog
            settings={settings}
            onSave={onSaveSettings}
            trigger={
              <Button variant="outline" size="sm">
                <Settings2 data-icon="inline-start" />
                更新源
              </Button>
            }
          />
        </div>
      </CardHeader>
      <CardContent className="grid gap-4 pt-5">
        <div className="grid gap-3 rounded-xl border p-4 md:grid-cols-[1fr_auto] md:items-center">
          <div className="grid gap-3 sm:grid-cols-3">
            <RuntimeValue label="服务状态">
              <span className={`size-2 rounded-full ${status?.online ? "bg-emerald-500" : "bg-destructive"}`} />
              <strong>{status?.online ? "在线" : "离线"}</strong>
            </RuntimeValue>
            <RuntimeValue label="当前版本">
              <strong className="font-mono">{version?.current_version ?? "不可用"}</strong>
              <Button
                variant="ghost"
                size="sm"
                className="h-7 px-2"
                onClick={() => void onCheckVersion()}
                disabled={checkingVersion}
              >
                {checkingVersion ? <Spinner /> : <RefreshCw />}
                检查更新
              </Button>
            </RuntimeValue>
            <RuntimeValue label="最新版本">
              <strong className="font-mono">{version?.latest_version ?? "暂不可用"}</strong>
              {version?.update_available ? <Badge variant="outline">可更新</Badge> : null}
              {version?.release_url ? (
                <a
                  className="inline-flex items-center gap-1 text-xs text-primary hover:underline"
                  href={version.release_url}
                  target="_blank"
                  rel="noreferrer"
                >
                  发布说明 <ExternalLink className="size-3" />
                </a>
              ) : null}
            </RuntimeValue>
          </div>
          <div className="flex flex-wrap gap-2 md:justify-end">
            <OnlineUpdateButton
              disabled={!version?.update_available || !version.latest_version || updatingVersion}
              updating={updatingVersion}
              target={version?.latest_version}
              onUpdate={onUpdateVersion}
            />
            <CaddyUploadButton disabled={updatingVersion} onUpload={onUploadVersion} />
          </div>
        </div>

        <p className="text-xs text-muted-foreground">
          上传包须匹配当前系统与 amd64 架构，并包含 <code>dns.providers.alidns</code> 模块。
        </p>
        {version?.error_message ? (
          <p className="rounded-lg border border-dashed p-3 text-xs text-destructive">
            {version.error_message}
          </p>
        ) : null}
        {updateTask && updateTask.status !== "idle" ? <UpdateTaskStatus task={updateTask} /> : null}
      </CardContent>
    </Card>
  );
}

function OnlineUpdateButton({
  disabled,
  updating,
  target,
  onUpdate,
}: {
  disabled: boolean;
  updating: boolean;
  target?: string;
  onUpdate: () => Promise<void>;
}) {
  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button size="sm" disabled={disabled}>
          {updating ? <Spinner data-icon="inline-start" /> : <Download data-icon="inline-start" />}
          {target ? `在线更新到 ${target}` : "在线更新"}
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>更新托管 Caddy？</AlertDialogTitle>
          <AlertDialogDescription>
            系统会断点下载并校验 SHA-512，安装成功后保留当前配置重启 Caddy。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction onClick={() => void onUpdate()}>确认更新</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

const taskLabels: Record<CaddyUpdateTask["status"], string> = {
  idle: "空闲",
  queued: "等待开始",
  downloading: "正在下载",
  verifying: "正在校验",
  installing: "正在安装",
  restarting: "正在重启",
  succeeded: "安装完成",
  failed: "安装失败",
};

function UpdateTaskStatus({ task }: { task: CaddyUpdateTask }) {
  const active = !["succeeded", "failed"].includes(task.status);
  return (
    <div className="grid gap-2 rounded-lg border bg-muted/20 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2 text-sm">
        <span className="font-medium">{taskLabels[task.status]}</span>
        <div className="flex gap-3 text-xs text-muted-foreground">
          {task.attempt ? <span>尝试 {task.attempt}/3</span> : null}
          {task.http_status ? <span>HTTP {task.http_status}</span> : null}
          {task.target_version ? <span className="font-mono">v{task.target_version}</span> : null}
        </div>
      </div>
      {active ? <Progress value={task.progress || undefined} /> : null}
      {task.downloaded_bytes ? (
        <p className="text-xs text-muted-foreground">已下载 {formatBytes(task.downloaded_bytes)}</p>
      ) : null}
      {task.effective_url ? (
        <p className="break-all text-xs text-muted-foreground">有效地址：{task.effective_url}</p>
      ) : null}
      {task.error_message ? <p className="text-xs text-destructive">{task.error_message}</p> : null}
    </div>
  );
}

function RuntimeValue({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="grid content-start gap-1">
      <span className="text-xs text-muted-foreground">{label}</span>
      <div className="flex min-h-7 flex-wrap items-center gap-2">{children}</div>
    </div>
  );
}

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KiB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MiB`;
}

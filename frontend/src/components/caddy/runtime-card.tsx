import { Braces, Download, ExternalLink, RefreshCw, Settings2 } from "lucide-react";

import type { CaddySettings, CaddyStatus, CaddyVersion } from "@/api/caddy";
import { CaddySettingsDialog } from "@/components/caddy/caddy-settings-dialog";
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

type Props = {
  status: CaddyStatus | null;
  version: CaddyVersion | null;
  settings: CaddySettings | null;
  loading: boolean;
  checkingVersion: boolean;
  updatingVersion: boolean;
  onRefresh: () => Promise<void>;
  onShowConfig: () => Promise<void>;
  onCheckVersion: () => Promise<void>;
  onUpdateVersion: () => Promise<void>;
  onSaveSettings: (settings: CaddySettings) => Promise<void>;
};

export function RuntimeCard({
  status,
  version,
  settings,
  loading,
  checkingVersion,
  updatingVersion,
  onRefresh,
  onShowConfig,
  onCheckVersion,
  onUpdateVersion,
  onSaveSettings,
}: Props) {
  return (
    <Card className="overflow-hidden border-primary/20">
      <CardHeader className="border-b bg-primary/[0.035]">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <CardTitle>Caddy 运行时</CardTitle>
            <CardDescription>运行状态、版本与维护操作集中管理。</CardDescription>
          </div>
          <CaddySettingsDialog
            settings={settings}
            onSave={onSaveSettings}
            trigger={
              <Button variant="outline">
                <Settings2 data-icon="inline-start" />
                更新源设置
              </Button>
            }
          />
        </div>
      </CardHeader>
      <CardContent className="grid gap-5 pt-5">
        <div className="grid overflow-hidden rounded-xl border sm:grid-cols-3">
          <RuntimeMetric
            label="服务状态"
            value={status?.online ? "在线" : "离线"}
            online={status?.online}
          />
          <RuntimeMetric label="当前 Caddy" value={version?.current_version ?? "不可用"} />
          <RuntimeMetric
            label="最新稳定版"
            value={version?.latest_version ?? "暂不可用"}
            badge={version?.update_available ? "可更新" : "已是目标版本"}
          />
        </div>

        {version?.error_message ? (
          <p className="rounded-lg border border-dashed p-3 text-xs text-muted-foreground">
            {version.error_message}
          </p>
        ) : null}

        <div className="flex flex-wrap items-center justify-between gap-3 border-t pt-4">
          <div className="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" onClick={() => void onRefresh()} disabled={loading}>
              {loading ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <RefreshCw data-icon="inline-start" />
              )}
              刷新状态
            </Button>
            <Button variant="outline" size="sm" onClick={() => void onShowConfig()}>
              <Braces data-icon="inline-start" />
              当前 JSON
            </Button>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => void onCheckVersion()}
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
                    <AlertDialogAction onClick={() => void onUpdateVersion()}>
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
        </div>
      </CardContent>
    </Card>
  );
}

function RuntimeMetric({
  label,
  value,
  online,
  badge,
}: {
  label: string;
  value: string;
  online?: boolean;
  badge?: string;
}) {
  return (
    <div className="border-b p-4 last:border-b-0 sm:border-r sm:border-b-0 sm:last:border-r-0">
      <p className="text-xs text-muted-foreground">{label}</p>
      <div className="mt-2 flex items-center gap-2">
        {online !== undefined ? (
          <span className={`size-2 rounded-full ${online ? "bg-emerald-500" : "bg-destructive"}`} />
        ) : null}
        <span className="font-mono text-lg font-semibold">{value}</span>
        {badge ? <Badge variant="outline">{badge}</Badge> : null}
      </div>
    </div>
  );
}

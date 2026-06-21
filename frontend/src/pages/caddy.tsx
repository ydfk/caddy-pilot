import { useCallback, useEffect, useState } from "react";
import {
  Braces,
  Copy,
  ExternalLink,
  PackageCheck,
  RefreshCw,
  Rocket,
  ShieldCheck,
} from "lucide-react";
import { toast } from "sonner";

import {
  getCaddyStatus,
  getCaddyVersion,
  getCurrentCaddyConfig,
  previewCaddyConfig,
  publishCaddyConfig,
  validateCaddyConfig,
  type CaddyStatus,
  type CaddyVersion,
} from "@/api/caddy";
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
  const [currentConfig, setCurrentConfig] = useState<unknown>(null);
  const [preview, setPreview] = useState<unknown>(null);
  const [version, setVersion] = useState<CaddyVersion | null>(null);
  const [loading, setLoading] = useState(true);
  const [publishing, setPublishing] = useState(false);
  const [checkingVersion, setCheckingVersion] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    const [statusResult, configResult, versionResult] = await Promise.allSettled([
      getCaddyStatus(),
      getCurrentCaddyConfig(),
      getCaddyVersion(),
    ]);
    if (statusResult.status === "fulfilled") setStatus(statusResult.value);
    else
      toast.error(
        statusResult.reason instanceof Error ? statusResult.reason.message : "读取 Caddy 状态失败"
      );
    if (configResult.status === "fulfilled") setCurrentConfig(configResult.value.caddy_json);
    else setCurrentConfig(null);
    if (versionResult.status === "fulfilled") setVersion(versionResult.value);
    else {
      setVersion(null);
      toast.error(
        versionResult.reason instanceof Error ? versionResult.reason.message : "读取 Caddy 版本失败"
      );
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function showPreview() {
    try {
      const result = await previewCaddyConfig();
      setPreview(result.caddy_json);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "生成配置预览失败");
    }
  }

  async function validate() {
    try {
      await validateCaddyConfig();
      toast.success("生成配置通过基础校验");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "配置校验失败");
    }
  }

  async function publish() {
    setPublishing(true);
    try {
      const version = await publishCaddyConfig();
      toast.success(`配置 v${version.version} 已发布`);
      await refresh();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "发布配置失败");
    } finally {
      setPublishing(false);
    }
  }

  async function checkVersion() {
    setCheckingVersion(true);
    try {
      setVersion(await getCaddyVersion());
      toast.success("Caddy 版本信息已刷新");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "检查 Caddy 版本失败");
    } finally {
      setCheckingVersion(false);
    }
  }

  async function copyUpdateCommand() {
    if (!version?.update_command) return;
    try {
      await navigator.clipboard.writeText(version.update_command);
      toast.success("更新命令已复制");
    } catch {
      toast.error("浏览器拒绝访问剪贴板，请手动复制更新命令");
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow="CADDY / RUNTIME"
        title="Caddy 状态"
        description="检查本地 Admin API、预览生成结果并显式发布。"
        actions={
          <>
            <Button variant="outline" onClick={() => void refresh()} disabled={loading}>
              {loading ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <RefreshCw data-icon="inline-start" />
              )}
              刷新
            </Button>
            <Button variant="outline" onClick={() => void showPreview()}>
              <Braces data-icon="inline-start" /> 预览
            </Button>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button disabled={publishing}>
                  {publishing ? (
                    <Spinner data-icon="inline-start" />
                  ) : (
                    <Rocket data-icon="inline-start" />
                  )}
                  发布
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>发布当前业务配置？</AlertDialogTitle>
                  <AlertDialogDescription>
                    后端会重新生成完整 JSON、注入 :8080 管理入口并调用本地 Caddy /load。
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>取消</AlertDialogCancel>
                  <AlertDialogAction onClick={() => void publish()}>确认发布</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </>
        }
      />

      {status?.error_message ? (
        <Alert variant="destructive">
          <AlertTitle>Caddy Admin API 不可用</AlertTitle>
          <AlertDescription>{status.error_message}</AlertDescription>
        </Alert>
      ) : null}

      <div className="grid gap-4 lg:grid-cols-[minmax(18rem,0.7fr)_minmax(0,1.3fr)]">
        <div className="flex flex-col gap-4">
          <Card>
            <CardHeader>
              <CardTitle>连接信息</CardTitle>
              <CardDescription>Admin API 不应暴露到宿主机</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <div>
                <p className="text-xs text-muted-foreground">Admin API 地址</p>
                <p className="mt-1 break-all font-mono text-sm">
                  {status?.admin_api ?? "正在读取…"}
                </p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground">连接状态</p>
                <Badge className="mt-2" variant={status?.online ? "default" : "destructive"}>
                  {status?.online ? "在线" : "离线"}
                </Badge>
              </div>
              <Button variant="secondary" onClick={() => void validate()}>
                <ShieldCheck data-icon="inline-start" /> 校验生成配置
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <PackageCheck className="size-4" /> 版本管理
              </CardTitle>
              <CardDescription>容器镜像固定版本，更新通过重建完成</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <p className="text-xs text-muted-foreground">当前版本</p>
                  <p className="mt-1 font-mono text-sm">
                    {version?.current_version ?? (loading ? "读取中…" : "不可用")}
                  </p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground">最新稳定版</p>
                  <p className="mt-1 font-mono text-sm">{version?.latest_version ?? "暂不可用"}</p>
                </div>
              </div>
              <div className="grid gap-2 rounded-md border bg-muted/30 p-3 text-xs">
                <div>
                  <p className="text-muted-foreground">Caddy 可执行文件</p>
                  <p className="mt-1 break-all font-mono">
                    {version?.binary_path ?? "由 PATH 解析"}
                  </p>
                </div>
                <div>
                  <p className="text-muted-foreground">版本校验地址</p>
                  <p className="mt-1 break-all font-mono">
                    {version?.version_check_url ?? "暂不可用"}
                  </p>
                </div>
                {version?.update_url ? (
                  <div>
                    <p className="text-muted-foreground">自定义更新地址</p>
                    <p className="mt-1 break-all font-mono">{version.update_url}</p>
                  </div>
                ) : null}
              </div>
              <Badge variant={version?.update_available ? "secondary" : "outline"}>
                {!version?.latest_version
                  ? "尚未获取最新版本"
                  : version.update_available
                    ? "有可用更新"
                    : "当前已是目标版本"}
              </Badge>
              {version?.error_message ? (
                <p className="text-xs text-muted-foreground">{version.error_message}</p>
              ) : null}
              {version?.update_available && version.update_command ? (
                <code className="break-all rounded-md bg-muted p-3 font-mono text-xs">
                  {version.update_command}
                </code>
              ) : null}
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
                {version?.update_available && version.update_command ? (
                  <Button size="sm" onClick={() => void copyUpdateCommand()}>
                    <Copy data-icon="inline-start" /> 复制更新命令
                  </Button>
                ) : null}
                {version?.release_url ? (
                  <Button variant="ghost" size="sm" asChild>
                    <a href={version.release_url} target="_blank" rel="noreferrer">
                      <ExternalLink data-icon="inline-start" /> 发布说明
                    </a>
                  </Button>
                ) : null}
              </div>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>当前 Caddy JSON</CardTitle>
            <CardDescription>从 GET /config/ 实时读取</CardDescription>
          </CardHeader>
          <CardContent>
            <pre className="max-h-[62vh] overflow-auto rounded-lg bg-muted p-4 font-mono text-xs leading-relaxed">
              {currentConfig ? JSON.stringify(currentConfig, null, 2) : "当前配置不可用"}
            </pre>
          </CardContent>
        </Card>
      </div>

      <JSONDialog
        open={preview !== null}
        onOpenChange={(open) => !open && setPreview(null)}
        title="完整 Caddy JSON 预览"
        description="仅预览，不会调用 /load。"
        value={preview}
      />
    </div>
  );
}

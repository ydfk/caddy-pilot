import { useCallback, useEffect, useState } from "react";
import { Braces, RefreshCw, Rocket, ShieldCheck } from "lucide-react";
import { toast } from "sonner";

import {
  getCaddyStatus,
  getCurrentCaddyConfig,
  previewCaddyConfig,
  publishCaddyConfig,
  validateCaddyConfig,
  type CaddyStatus,
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
  const [loading, setLoading] = useState(true);
  const [publishing, setPublishing] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    const [statusResult, configResult] = await Promise.allSettled([
      getCaddyStatus(),
      getCurrentCaddyConfig(),
    ]);
    if (statusResult.status === "fulfilled") setStatus(statusResult.value);
    else
      toast.error(
        statusResult.reason instanceof Error ? statusResult.reason.message : "读取 Caddy 状态失败"
      );
    if (configResult.status === "fulfilled") setCurrentConfig(configResult.value.caddy_json);
    else setCurrentConfig(null);
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

  return (
    <div className="flex flex-col gap-6">
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

      <div className="grid gap-4 lg:grid-cols-[minmax(0,0.65fr)_minmax(0,1.35fr)]">
        <Card className="h-fit">
          <CardHeader>
            <CardTitle>连接信息</CardTitle>
            <CardDescription>Admin API 不应暴露到宿主机</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-5">
            <div>
              <p className="text-xs text-muted-foreground">Admin API 地址</p>
              <p className="mt-1 break-all font-mono text-sm">{status?.admin_api ?? "正在读取…"}</p>
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

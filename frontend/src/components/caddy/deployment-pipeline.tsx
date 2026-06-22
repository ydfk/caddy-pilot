import { useState } from "react";
import { Braces, CheckCircle2, Rocket, ShieldCheck } from "lucide-react";
import { toast } from "sonner";

import { previewCaddyConfig, publishCaddyConfig, validateCaddyConfig } from "@/api/caddy";
import { JSONDialog } from "@/components/json-dialog";
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
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

type Props = {
  dirty: boolean;
  latestVersion?: number;
  onPublished: () => Promise<void>;
};

export function DeploymentPipeline({ dirty, latestVersion, onPublished }: Props) {
  const [validated, setValidated] = useState(false);
  const [validating, setValidating] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [preview, setPreview] = useState<unknown>(null);
  const [reason, setReason] = useState("");

  async function showPreview() {
    try {
      setPreview((await previewCaddyConfig()).caddy_json);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "生成预览失败");
    }
  }

  async function validate() {
    setValidating(true);
    try {
      await validateCaddyConfig();
      setValidated(true);
      toast.success("配置校验通过，可以发布");
    } catch (error) {
      setValidated(false);
      toast.error(error instanceof Error ? error.message : "配置校验失败");
    } finally {
      setValidating(false);
    }
  }

  async function publish() {
    setPublishing(true);
    try {
      const created = await publishCaddyConfig(reason.trim() || "手动发布");
      toast.success(`配置 v${created.version} 已发布`);
      setValidated(false);
      setReason("");
      await onPublished();
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "发布配置失败");
    } finally {
      setPublishing(false);
    }
  }

  return (
    <Card className="overflow-hidden border-primary/20">
      <CardHeader className="border-b bg-primary/[0.035]">
        <div className="flex items-center justify-between gap-3">
          <div>
            <CardTitle>配置发布</CardTitle>
            <CardDescription>校验通过后才开放发布，两个动作保持连续。</CardDescription>
          </div>
          <Badge variant={dirty ? "secondary" : "outline"}>
            {dirty ? "存在未发布修改" : latestVersion ? `已同步 v${latestVersion}` : "暂无发布"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="pt-5">
        {dirty ? (
          <div className="grid gap-4">
            <div className="grid gap-3 md:grid-cols-[1fr_auto_1fr] md:items-stretch">
              <div
                className={`rounded-xl border p-4 ${validated ? "border-emerald-500/40 bg-emerald-500/5" : "bg-muted/20"}`}
              >
                <div className="flex items-center gap-2 text-sm font-semibold">
                  <span className="flex size-6 items-center justify-center rounded-full bg-foreground text-xs text-background">
                    1
                  </span>
                  校验配置
                </div>
                <p className="mt-2 text-xs text-muted-foreground">
                  生成完整 JSON 并执行结构与管理入口检查。
                </p>
                <Button
                  className="mt-4 w-full"
                  variant={validated ? "outline" : "secondary"}
                  onClick={() => void validate()}
                  disabled={validating}
                >
                  {validating ? (
                    <Spinner data-icon="inline-start" />
                  ) : validated ? (
                    <CheckCircle2 data-icon="inline-start" />
                  ) : (
                    <ShieldCheck data-icon="inline-start" />
                  )}
                  {validated ? "重新校验" : "开始校验"}
                </Button>
              </div>
              <div className="hidden items-center text-muted-foreground md:flex">→</div>
              <div
                className={`rounded-xl border p-4 ${validated ? "bg-primary/[0.035]" : "bg-muted/10 opacity-70"}`}
              >
                <div className="flex items-center gap-2 text-sm font-semibold">
                  <span className="flex size-6 items-center justify-center rounded-full bg-foreground text-xs text-background">
                    2
                  </span>
                  发布配置
                </div>
                <Input
                  className="mt-3"
                  value={reason}
                  onChange={(event) => setReason(event.target.value)}
                  placeholder="发布说明（可选）"
                  disabled={!validated}
                />
                <AlertDialog>
                  <AlertDialogTrigger asChild>
                    <Button className="mt-3 w-full" disabled={!validated || publishing}>
                      {publishing ? (
                        <Spinner data-icon="inline-start" />
                      ) : (
                        <Rocket data-icon="inline-start" />
                      )}
                      发布到 Caddy
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>发布已校验的配置？</AlertDialogTitle>
                      <AlertDialogDescription>
                        系统会创建配置版本并加载到托管 Caddy，管理入口会受到保护。
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>取消</AlertDialogCancel>
                      <AlertDialogAction onClick={() => void publish()}>确认发布</AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </div>
            <Button variant="ghost" className="w-fit" onClick={() => void showPreview()}>
              <Braces data-icon="inline-start" />
              预览待发布 JSON
            </Button>
          </div>
        ) : (
          <div className="flex flex-col items-start gap-3 rounded-xl border border-dashed p-5">
            <div className="flex items-center gap-2 font-medium text-emerald-600">
              <CheckCircle2 className="size-4" />
              站点配置与已发布版本一致
            </div>
            <p className="text-sm text-muted-foreground">
              只有站点内容发生变化后，才需要重新校验和发布。
            </p>
            <Button variant="outline" size="sm" onClick={() => void showPreview()}>
              <Braces data-icon="inline-start" />
              查看生成 JSON
            </Button>
          </div>
        )}
      </CardContent>
      <JSONDialog
        open={preview !== null}
        onOpenChange={(open) => !open && setPreview(null)}
        title="待发布 Caddy JSON"
        description="当前业务配置生成结果，预览不会发布。"
        value={preview}
      />
    </Card>
  );
}

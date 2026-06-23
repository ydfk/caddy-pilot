import { useState } from "react";
import { CheckCircle2, Rocket, ShieldCheck } from "lucide-react";
import { toast } from "sonner";

import { publishCaddyConfig, validateCaddyConfig } from "@/api/caddy";
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
import { Button } from "@/components/ui/button";
import { CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";

export function DeploymentPipeline({
  dirty,
  onPublished,
}: {
  dirty: boolean;
  onPublished: () => Promise<void>;
}) {
  const [validated, setValidated] = useState(false);
  const [validating, setValidating] = useState(false);
  const [publishing, setPublishing] = useState(false);
  const [reason, setReason] = useState("");

  if (!dirty) return null;

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
    <CardContent className="border-b pt-5">
      <div className="mb-3 flex items-center gap-2 text-sm font-medium text-amber-700">
        <span className="size-2 rounded-full bg-amber-500" />
        检测到站点修改，请连续完成校验与发布
      </div>
      <div className="grid gap-3 md:grid-cols-[1fr_auto_1fr] md:items-center">
        <div className="flex items-center gap-3 rounded-lg border bg-muted/20 p-3">
          <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-foreground text-xs text-background">
            1
          </span>
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium">校验配置</p>
            <p className="text-xs text-muted-foreground">检查 JSON 结构和管理入口</p>
          </div>
          <Button size="sm" variant={validated ? "outline" : "secondary"} onClick={() => void validate()} disabled={validating}>
            {validating ? <Spinner /> : validated ? <CheckCircle2 /> : <ShieldCheck />}
            {validated ? "重新校验" : "校验"}
          </Button>
        </div>
        <span className="hidden text-muted-foreground md:block">→</span>
        <div className={`flex items-center gap-2 rounded-lg border p-3 ${validated ? "bg-primary/[0.035]" : "bg-muted/10 opacity-70"}`}>
          <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-foreground text-xs text-background">
            2
          </span>
          <Input
            value={reason}
            onChange={(event) => setReason(event.target.value)}
            placeholder="发布说明（可选）"
            disabled={!validated}
          />
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button size="sm" disabled={!validated || publishing}>
                {publishing ? <Spinner /> : <Rocket />}
                发布
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>发布已校验的配置？</AlertDialogTitle>
                <AlertDialogDescription>
                  系统会创建配置版本并加载 JSON 到托管 Caddy，管理入口会受到保护。
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
    </CardContent>
  );
}

import { RefreshCw, Rocket, ShieldCheck } from "lucide-react";

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
import { Spinner } from "@/components/ui/spinner";

type Props = {
  hasChanges: boolean;
  validated: boolean;
  validating: boolean;
  publishing: boolean;
  onValidate: () => void;
  onPublish: () => void;
  onRegenerate: () => void;
};

export function ProxySitePublishActions({
  hasChanges,
  validated,
  validating,
  publishing,
  onValidate,
  onPublish,
  onRegenerate,
}: Props) {
  if (!hasChanges) {
    return (
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button variant="outline" disabled={publishing}>
            {publishing ? <Spinner /> : <RefreshCw />}
            重新生成配置
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>重新生成并发布完整配置？</AlertDialogTitle>
            <AlertDialogDescription>
              即使站点没有变化，也会重新生成、校验并发布一份新的 Caddy 配置版本。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={onRegenerate}>确认重新生成</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    );
  }

  return (
    <>
      <Button variant="outline" onClick={onValidate} disabled={validating}>
        {validating ? <Spinner /> : <ShieldCheck />}
        1. 校验配置
      </Button>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button disabled={!validated || publishing}>
            {publishing ? <Spinner /> : <Rocket />}
            2. 发布
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>发布已校验的站点配置？</AlertDialogTitle>
            <AlertDialogDescription>
              后端会重新生成完整配置，并安全发布到托管的 Caddy。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={onPublish}>确认发布</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

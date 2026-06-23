import { useState, type ChangeEvent, type ReactNode } from "react";
import { Upload } from "lucide-react";

import type { NginxImportResult } from "@/api/proxy-sites";
import { DialogError } from "@/components/dialog-error";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Textarea } from "@/components/ui/textarea";

type Props = {
  trigger: ReactNode;
  onImport: (config: string) => Promise<NginxImportResult>;
  onImported: (result: NginxImportResult) => void;
};

export function NginxImportDialog({ trigger, onImport, onImported }: Props) {
  const [open, setOpen] = useState(false);
  const [config, setConfig] = useState("");
  const [filename, setFilename] = useState("");
  const [pending, setPending] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  async function readFile(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    if (file.size > 2 * 1024 * 1024) {
      setErrorMessage("Nginx 配置不能超过 2 MiB");
      return;
    }
    setFilename(file.name);
    setConfig(await file.text());
    setErrorMessage("");
  }

  async function submit() {
    if (!config.trim()) {
      setErrorMessage("请上传或粘贴 Nginx 配置");
      return;
    }
    setPending(true);
    setErrorMessage("");
    try {
      const result = await onImport(config);
      onImported(result);
      setOpen(false);
      setConfig("");
      setFilename("");
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "导入 Nginx 配置失败");
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>导入 Nginx 配置</DialogTitle>
          <DialogDescription>
            识别 server_name、listen、proxy_pass、upstream 和 HTTPS 跳转；导入后默认停用。
          </DialogDescription>
        </DialogHeader>
        <FieldGroup className="py-4">
          {errorMessage ? <DialogError message={errorMessage} /> : null}
          <Field>
            <FieldLabel htmlFor="nginx-config-file">配置文件</FieldLabel>
            <div className="flex items-center gap-3">
              <Input
                id="nginx-config-file"
                type="file"
                accept=".conf,.nginx,text/plain"
                onChange={(event) => void readFile(event)}
              />
              {filename ? (
                <span className="shrink-0 text-xs text-muted-foreground">{filename}</span>
              ) : null}
            </div>
          </Field>
          <Field>
            <FieldLabel htmlFor="nginx-config-content">配置内容</FieldLabel>
            <Textarea
              id="nginx-config-content"
              rows={15}
              className="font-mono text-xs"
              value={config}
              onChange={(event) => setConfig(event.target.value)}
              placeholder={
                "server {\n    listen 80;\n    server_name example.com;\n    location / { proxy_pass http://127.0.0.1:3000; }\n}"
              }
            />
            <FieldDescription>
              可以导入完整 nginx.conf，也可以只粘贴 server/upstream 片段。
            </FieldDescription>
          </Field>
        </FieldGroup>
        <DialogFooter>
          <Button onClick={() => void submit()} disabled={pending}>
            {pending ? <Spinner data-icon="inline-start" /> : <Upload data-icon="inline-start" />}
            导入为停用站点
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

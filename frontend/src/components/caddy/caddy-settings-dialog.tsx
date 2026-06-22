import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import { Save } from "lucide-react";

import type { CaddySettings } from "@/api/caddy";
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

type Props = {
  settings: CaddySettings | null;
  trigger: ReactNode;
  onSave: (settings: CaddySettings) => Promise<void>;
};

export function CaddySettingsDialog({ settings, trigger, onSave }: Props) {
  const [open, setOpen] = useState(false);
  const [values, setValues] = useState<CaddySettings>({
    version_check_url: "",
    download_url: "",
    checksum_url: "",
  });
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (open && settings) setValues(settings);
  }, [open, settings]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setPending(true);
    try {
      await onSave(values);
      setOpen(false);
    } finally {
      setPending(false);
    }
  }

  function update(key: keyof CaddySettings, value: string) {
    setValues((current) => ({ ...current, [key]: value }));
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-2xl">
        <form onSubmit={submit}>
          <DialogHeader>
            <DialogTitle>更新源设置</DialogTitle>
            <DialogDescription>
              保存在系统数据库中，修改后立即用于版本检查和下一次托管更新。
            </DialogDescription>
          </DialogHeader>
          <FieldGroup className="py-5">
            <Field>
              <FieldLabel htmlFor="version-check-url">版本校验地址</FieldLabel>
              <Input
                id="version-check-url"
                className="font-mono text-xs"
                value={values.version_check_url}
                onChange={(event) => update("version_check_url", event.target.value)}
                required
              />
              <FieldDescription>
                支持 GitHub Release 响应，或包含 version、update_url 的 JSON。
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="download-url">Caddy 下载地址</FieldLabel>
              <Input
                id="download-url"
                className="font-mono text-xs"
                value={values.download_url}
                onChange={(event) => update("download_url", event.target.value)}
                required
              />
              <FieldDescription>
                支持 version、os、arch、ext 占位符，目标文件必须包含阿里云 DNS 模块。
              </FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="checksum-url">SHA-512 清单地址</FieldLabel>
              <Input
                id="checksum-url"
                className="font-mono text-xs"
                value={values.checksum_url}
                onChange={(event) => update("checksum_url", event.target.value)}
                placeholder="可选"
              />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button type="submit" disabled={pending}>
              {pending ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
              保存设置
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

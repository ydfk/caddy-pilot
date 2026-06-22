import { useEffect, useState, type FormEvent, type ReactNode } from "react";
import { Save } from "lucide-react";

import type { DNSProvider, DNSProviderPayload } from "@/api/dns-providers";
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
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";

type Props = {
  provider?: DNSProvider | null;
  trigger: ReactNode;
  onSave: (payload: DNSProviderPayload) => Promise<unknown>;
};

export function DNSProviderDialog({ provider, trigger, onSave }: Props) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [accessKeyID, setAccessKeyID] = useState("");
  const [accessKeySecret, setAccessKeySecret] = useState("");
  const [regionID, setRegionID] = useState("cn-hangzhou");
  const [enabled, setEnabled] = useState(true);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    if (!open) return;
    setName(provider?.name ?? "");
    setAccessKeyID("");
    setAccessKeySecret("");
    setRegionID(provider?.region_id ?? "cn-hangzhou");
    setEnabled(provider?.enabled ?? true);
  }, [open, provider]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setPending(true);
    try {
      await onSave({
        name: name.trim(),
        provider_type: "alidns",
        access_key_id: accessKeyID.trim(),
        access_key_secret: accessKeySecret,
        region_id: regionID.trim(),
        enabled,
      });
      setOpen(false);
    } finally {
      setPending(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <form onSubmit={submit}>
          <DialogHeader>
            <DialogTitle>{provider ? "编辑 DNS Provider" : "新增 DNS Provider"}</DialogTitle>
            <DialogDescription>凭据加密保存，页面不会再次显示 AccessKey Secret。</DialogDescription>
          </DialogHeader>
          <FieldGroup className="py-5">
            <div className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel>服务商</FieldLabel>
                <Select value="alidns" disabled>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="alidns">阿里云 DNS</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel htmlFor="dns-name">名称</FieldLabel>
                <Input
                  id="dns-name"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="生产账号"
                  required
                />
              </Field>
            </div>
            <Field>
              <FieldLabel htmlFor="access-key-id">AccessKey ID</FieldLabel>
              <Input
                id="access-key-id"
                value={accessKeyID}
                onChange={(event) => setAccessKeyID(event.target.value)}
                placeholder={provider?.access_key_id_hint || "LTAI..."}
                required={!provider}
              />
              {provider ? <FieldDescription>留空时保留当前 AccessKey ID。</FieldDescription> : null}
            </Field>
            <Field>
              <FieldLabel htmlFor="access-key-secret">AccessKey Secret</FieldLabel>
              <Input
                id="access-key-secret"
                type="password"
                autoComplete="new-password"
                value={accessKeySecret}
                onChange={(event) => setAccessKeySecret(event.target.value)}
                required={!provider}
              />
              {provider ? <FieldDescription>留空时保留当前密钥。</FieldDescription> : null}
            </Field>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor="region-id">Region ID</FieldLabel>
                <Input
                  id="region-id"
                  value={regionID}
                  onChange={(event) => setRegionID(event.target.value)}
                />
              </Field>
              <Field orientation="horizontal">
                <div className="flex-1">
                  <FieldLabel htmlFor="dns-enabled">启用</FieldLabel>
                  <FieldDescription>停用后不能用于签发。</FieldDescription>
                </div>
                <Switch id="dns-enabled" checked={enabled} onCheckedChange={setEnabled} />
              </Field>
            </div>
          </FieldGroup>
          <DialogFooter>
            <Button type="submit" disabled={pending}>
              {pending ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
              保存
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

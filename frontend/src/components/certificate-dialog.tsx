import { useEffect, useMemo, useState, type FormEvent, type ReactNode } from "react";
import { Plus, Save, Trash2 } from "lucide-react";

import type { CertificateProfile, CertificateProfilePayload } from "@/api/certificates";
import type { DNSProvider, DNSProviderPayload } from "@/api/dns-providers";
import { DNSProviderDialog } from "@/components/dns-provider-dialog";
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
  profile?: CertificateProfile | null;
  providers: DNSProvider[];
  trigger: ReactNode;
  onSave: (payload: CertificateProfilePayload) => Promise<unknown>;
  onCreateProvider: (payload: DNSProviderPayload) => Promise<DNSProvider>;
};

export function CertificateDialog({
  profile,
  providers,
  trigger,
  onSave,
  onCreateProvider,
}: Props) {
  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [type, setType] = useState<"single" | "wildcard">("wildcard");
  const [subjects, setSubjects] = useState([""]);
  const [challenge, setChallenge] = useState<"http" | "dns">("dns");
  const [providerID, setProviderID] = useState("");
  const [enabled, setEnabled] = useState(true);
  const [pending, setPending] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const enabledProviders = useMemo(() => providers.filter((item) => item.enabled), [providers]);

  useEffect(() => {
    if (!open) return;
    setName(profile?.name ?? "");
    setType(profile?.certificate_type ?? "wildcard");
    setSubjects(profile?.subjects.length ? profile.subjects : [""]);
    setChallenge(profile?.challenge_type ?? "dns");
    setProviderID(profile?.dns_provider_id ?? enabledProviders[0]?.id ?? "");
    setEnabled(profile?.enabled ?? true);
    setErrorMessage("");
  }, [open, profile]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setPending(true);
    setErrorMessage("");
    try {
      const challengeType = type === "wildcard" ? "dns" : challenge;
      await onSave({
        name: name.trim(),
        certificate_type: type,
        subjects: subjects.map((item) => item.trim()).filter(Boolean),
        challenge_type: challengeType,
        dns_provider_id: challengeType === "dns" ? providerID : undefined,
        enabled,
      });
      setOpen(false);
    } catch (error) {
      setErrorMessage(error instanceof Error ? error.message : "保存证书配置失败");
    } finally {
      setPending(false);
    }
  }

  function updateSubject(index: number, value: string) {
    setSubjects((current) =>
      current.map((item, itemIndex) => (itemIndex === index ? value : item))
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{trigger}</DialogTrigger>
      <DialogContent className="sm:max-w-xl">
        <form onSubmit={submit}>
          <DialogHeader>
            <DialogTitle>{profile ? "编辑证书配置" : "新增证书配置"}</DialogTitle>
            <DialogDescription>Caddy 按此配置自动签发并续期证书。</DialogDescription>
          </DialogHeader>
          <FieldGroup className="py-5">
            {errorMessage ? <DialogError message={errorMessage} /> : null}
            <div className="grid gap-4 sm:grid-cols-2">
              <Field>
                <FieldLabel htmlFor="certificate-name">名称</FieldLabel>
                <Input
                  id="certificate-name"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="example.com 泛域名"
                  required
                />
              </Field>
              <Field>
                <FieldLabel>证书类型</FieldLabel>
                <Select value={type} onValueChange={(value) => setType(value as typeof type)}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="single">域名证书</SelectItem>
                      <SelectItem value="wildcard">通配符证书</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
            </div>
            <Field>
              <FieldLabel>证书域名</FieldLabel>
              <div className="flex flex-col gap-2">
                {subjects.map((subject, index) => (
                  <div key={index} className="flex gap-2">
                    <Input
                      className="font-mono"
                      value={subject}
                      onChange={(event) => updateSubject(index, event.target.value)}
                      placeholder={type === "wildcard" ? "*.example.com" : "example.com"}
                      required
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon"
                      disabled={subjects.length === 1}
                      onClick={() =>
                        setSubjects((current) =>
                          current.filter((_, itemIndex) => itemIndex !== index)
                        )
                      }
                    >
                      <Trash2 />
                    </Button>
                  </div>
                ))}
              </div>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setSubjects((current) => [...current, ""])}
              >
                <Plus data-icon="inline-start" />
                添加域名
              </Button>
            </Field>
            {type === "single" ? (
              <Field>
                <FieldLabel>验证方式</FieldLabel>
                <Select
                  value={challenge}
                  onValueChange={(value) => setChallenge(value as typeof challenge)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="http">自动验证（HTTP / TLS-ALPN）</SelectItem>
                      <SelectItem value="dns">DNS-01</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
            ) : null}
            {type === "wildcard" || challenge === "dns" ? (
              <Field>
                <FieldLabel>DNS Provider</FieldLabel>
                {enabledProviders.length ? (
                  <Select value={providerID} onValueChange={setProviderID}>
                    <SelectTrigger>
                      <SelectValue placeholder="选择 DNS Provider" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {enabledProviders.map((provider) => (
                          <SelectItem key={provider.id} value={provider.id}>
                            {provider.name} · 阿里云
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                ) : (
                  <FieldDescription>暂无可用 Provider，可直接在这里新增。</FieldDescription>
                )}
                <DNSProviderDialog
                  trigger={
                    <Button type="button" variant="outline" size="sm" className="w-fit">
                      <Plus data-icon="inline-start" />
                      新增 DNS Provider
                    </Button>
                  }
                  onSave={async (payload) => {
                    const created = await onCreateProvider(payload);
                    setProviderID(created.id);
                  }}
                />
              </Field>
            ) : null}
            <Field orientation="horizontal">
              <div className="flex-1">
                <FieldLabel htmlFor="certificate-enabled">启用</FieldLabel>
                <FieldDescription>停用后不能被新站点选择。</FieldDescription>
              </div>
              <Switch id="certificate-enabled" checked={enabled} onCheckedChange={setEnabled} />
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button
              type="submit"
              disabled={pending || ((type === "wildcard" || challenge === "dns") && !providerID)}
            >
              {pending ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
              保存
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

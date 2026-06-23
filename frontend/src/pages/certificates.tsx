import { useCallback, useEffect, useState } from "react";
import { BadgeCheck, CalendarClock, Pencil, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";

import {
  createCertificate,
  deleteCertificate,
  listCertificates,
  updateCertificate,
  type CertificateProfile,
  type CertificateProfilePayload,
} from "@/api/certificates";
import {
  createDNSProvider,
  listDNSProviders,
  type DNSProvider,
  type DNSProviderPayload,
} from "@/api/dns-providers";
import { CertificateDialog } from "@/components/certificate-dialog";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";

export default function CertificatesPage() {
  const [profiles, setProfiles] = useState<CertificateProfile[]>([]);
  const [providers, setProviders] = useState<DNSProvider[]>([]);
  const load = useCallback(async () => {
    try {
      const [nextProfiles, nextProviders] = await Promise.all([
        listCertificates(),
        listDNSProviders(),
      ]);
      setProfiles(nextProfiles);
      setProviders(nextProviders);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "读取证书配置失败");
    }
  }, []);
  useEffect(() => void load(), [load]);

  async function save(payload: CertificateProfilePayload, profile?: CertificateProfile) {
    try {
      if (profile) await updateCertificate(profile.id, payload);
      else await createCertificate(payload);
      await load();
      toast.success(profile ? "证书配置已更新" : "证书配置已创建");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "保存证书配置失败");
      throw error;
    }
  }
  async function createProvider(payload: DNSProviderPayload) {
    try {
      const created = await createDNSProvider(payload);
      setProviders((current) => [created, ...current]);
      toast.success("DNS Provider 已创建并选中");
      return created;
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "创建 DNS Provider 失败");
      throw error;
    }
  }
  async function remove(profile: CertificateProfile) {
    if (!window.confirm(`删除证书配置“${profile.name}”？`)) return;
    try {
      await deleteCertificate(profile.id);
      await load();
      toast.success("证书配置已删除");
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "删除证书配置失败");
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow="CERTIFICATES / PROFILES"
        title="证书"
        description="复用证书签发配置，由 Caddy 自动申请和续期。"
        actions={
          <CertificateDialog
            providers={providers}
            onCreateProvider={createProvider}
            trigger={
              <Button>
                <Plus data-icon="inline-start" />
                新增证书
              </Button>
            }
            onSave={(payload) => save(payload)}
          />
        }
      />
      <Card>
        <CardHeader>
          <CardTitle>证书配置</CardTitle>
          <CardDescription>通配符配置可直接在代理站点中选择。</CardDescription>
        </CardHeader>
        <CardContent>
          {profiles.length === 0 ? (
            <Empty>
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <BadgeCheck />
                </EmptyMedia>
                <EmptyTitle>还没有证书配置</EmptyTitle>
                <EmptyDescription>新增后可在站点 HTTPS 设置中复用。</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <div className="divide-y rounded-lg border">
              {profiles.map((profile) => (
                <div key={profile.id} className="flex items-center gap-3 p-3">
                  <BadgeCheck className="size-4 text-muted-foreground" />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center gap-2">
                      <p className="truncate text-sm font-medium">{profile.name}</p>
                      <Badge variant="outline">
                        {profile.certificate_type === "wildcard" ? "通配符" : "域名"}
                      </Badge>
                      <Badge variant={profile.enabled ? "secondary" : "outline"}>
                        {profile.enabled ? "启用" : "停用"}
                      </Badge>
                      <Badge variant={issuanceBadgeVariant(profile.issuance_state)}>
                        {issuanceLabels[profile.issuance_state]}
                      </Badge>
                    </div>
                    <p className="truncate font-mono text-xs text-muted-foreground">
                      {profile.subjects.join(", ")} · {profile.challenge_type.toUpperCase()}
                    </p>
                    <CertificateRuntimeInfo profile={profile} />
                  </div>
                  <CertificateDialog
                    profile={profile}
                    providers={providers}
                    onCreateProvider={createProvider}
                    trigger={
                      <Button variant="ghost" size="icon" aria-label={`编辑 ${profile.name}`}>
                        <Pencil />
                      </Button>
                    }
                    onSave={(payload) => save(payload, profile)}
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={`删除 ${profile.name}`}
                    onClick={() => void remove(profile)}
                  >
                    <Trash2 />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function CertificateRuntimeInfo({ profile }: { profile: CertificateProfile }) {
  return (
    <div className="mt-2 grid gap-1">
      <p className="text-xs text-muted-foreground">
        {profile.issuance_message} · {profile.usage_count} 个站点引用
      </p>
      {profile.last_error ? (
        <p className="rounded border border-destructive/20 bg-destructive/5 px-2 py-1 text-xs text-destructive">
          {profile.last_error}
        </p>
      ) : null}
      {profile.issued_certificates.map((certificate) => (
        <div
          key={certificate.serial_number}
          className="grid gap-1 rounded-md border bg-muted/20 px-2 py-1.5 text-[11px] text-muted-foreground sm:grid-cols-[1fr_auto]"
        >
          <span className="flex min-w-0 items-center gap-1.5">
            <CalendarClock className="size-3 shrink-0" />
            <span className="truncate">
              签发 {formatCertificateTime(certificate.issued_at)} · 到期{" "}
              {formatCertificateTime(certificate.expires_at)}
            </span>
          </span>
          <span className="truncate">{certificate.issuer}</span>
        </div>
      ))}
    </div>
  );
}

const issuanceLabels: Record<CertificateProfile["issuance_state"], string> = {
  unused: "未被使用",
  pending_publish: "等待发布",
  issuing: "正在签发",
  failed: "签发失败",
  issued: "已签发",
  expiring: "即将到期",
  expired: "已过期",
};

function issuanceBadgeVariant(state: CertificateProfile["issuance_state"]) {
  if (state === "failed" || state === "expired") return "destructive" as const;
  if (state === "issued") return "secondary" as const;
  return "outline" as const;
}

function formatCertificateTime(value: string) {
  return new Date(value).toLocaleDateString("zh-CN");
}

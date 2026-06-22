import { useCallback, useEffect, useState } from "react";
import { BadgeCheck, Pencil, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { createCertificate, deleteCertificate, listCertificates, updateCertificate, type CertificateProfile, type CertificateProfilePayload } from "@/api/certificates";
import { listDNSProviders, type DNSProvider } from "@/api/dns-providers";
import { CertificateDialog } from "@/components/certificate-dialog";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";

export default function CertificatesPage() {
  const [profiles, setProfiles] = useState<CertificateProfile[]>([]);
  const [providers, setProviders] = useState<DNSProvider[]>([]);
  const load = useCallback(async () => {
    try { const [nextProfiles, nextProviders] = await Promise.all([listCertificates(), listDNSProviders()]); setProfiles(nextProfiles); setProviders(nextProviders); }
    catch (error) { toast.error(error instanceof Error ? error.message : "读取证书配置失败"); }
  }, []);
  useEffect(() => void load(), [load]);

  async function save(payload: CertificateProfilePayload, profile?: CertificateProfile) {
    try { if (profile) await updateCertificate(profile.id, payload); else await createCertificate(payload); await load(); toast.success(profile ? "证书配置已更新" : "证书配置已创建"); }
    catch (error) { toast.error(error instanceof Error ? error.message : "保存证书配置失败"); throw error; }
  }
  async function remove(profile: CertificateProfile) {
    if (!window.confirm(`删除证书配置“${profile.name}”？`)) return;
    try { await deleteCertificate(profile.id); await load(); toast.success("证书配置已删除"); }
    catch (error) { toast.error(error instanceof Error ? error.message : "删除证书配置失败"); }
  }

  return <div className="flex flex-col gap-4">
    <PageHeader eyebrow="CERTIFICATES / PROFILES" title="证书" description="复用证书签发配置，由 Caddy 自动申请和续期。" actions={<CertificateDialog providers={providers} trigger={<Button><Plus data-icon="inline-start" />新增证书</Button>} onSave={(payload) => save(payload)} />} />
    <Card><CardHeader><CardTitle>证书配置</CardTitle><CardDescription>通配符配置可直接在代理站点中选择。</CardDescription></CardHeader><CardContent>
      {profiles.length === 0 ? <Empty><EmptyHeader><EmptyMedia variant="icon"><BadgeCheck /></EmptyMedia><EmptyTitle>还没有证书配置</EmptyTitle><EmptyDescription>新增后可在站点 HTTPS 设置中复用。</EmptyDescription></EmptyHeader></Empty> : <div className="divide-y rounded-lg border">{profiles.map((profile) => <div key={profile.id} className="flex items-center gap-3 p-3"><BadgeCheck className="size-4 text-muted-foreground" /><div className="min-w-0 flex-1"><div className="flex items-center gap-2"><p className="truncate text-sm font-medium">{profile.name}</p><Badge variant="outline">{profile.certificate_type === "wildcard" ? "通配符" : "域名"}</Badge><Badge variant={profile.enabled ? "secondary" : "outline"}>{profile.enabled ? "启用" : "停用"}</Badge></div><p className="truncate font-mono text-xs text-muted-foreground">{profile.subjects.join(", ")} · {profile.challenge_type.toUpperCase()}</p></div><CertificateDialog profile={profile} providers={providers} trigger={<Button variant="ghost" size="icon" aria-label={`编辑 ${profile.name}`}><Pencil /></Button>} onSave={(payload) => save(payload, profile)} /><Button variant="ghost" size="icon" aria-label={`删除 ${profile.name}`} onClick={() => void remove(profile)}><Trash2 /></Button></div>)}</div>}
    </CardContent></Card>
  </div>;
}

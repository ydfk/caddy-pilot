import { useCallback, useEffect, useState } from "react";
import { CloudCog, Pencil, Plus, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { createDNSProvider, deleteDNSProvider, listDNSProviders, updateDNSProvider, type DNSProvider, type DNSProviderPayload } from "@/api/dns-providers";
import { DNSProviderDialog } from "@/components/dns-provider-dialog";
import { PageHeader } from "@/components/page-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from "@/components/ui/empty";

export default function DNSProvidersPage() {
  const [providers, setProviders] = useState<DNSProvider[]>([]);
  const load = useCallback(async () => {
    try { setProviders(await listDNSProviders()); }
    catch (error) { toast.error(error instanceof Error ? error.message : "读取 DNS Provider 失败"); }
  }, []);
  useEffect(() => void load(), [load]);

  async function save(payload: DNSProviderPayload, provider?: DNSProvider) {
    try {
      if (provider) await updateDNSProvider(provider.id, payload); else await createDNSProvider(payload);
      toast.success(provider ? "DNS Provider 已更新" : "DNS Provider 已创建");
      await load();
    } catch (error) { toast.error(error instanceof Error ? error.message : "保存 DNS Provider 失败"); throw error; }
  }

  async function remove(provider: DNSProvider) {
    if (!window.confirm(`删除 DNS Provider“${provider.name}”？`)) return;
    try { await deleteDNSProvider(provider.id); await load(); toast.success("DNS Provider 已删除"); }
    catch (error) { toast.error(error instanceof Error ? error.message : "删除 DNS Provider 失败"); }
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader eyebrow="CERTIFICATES / DNS" title="DNS Provider" description="统一保存证书签发所需的 DNS 服务商凭据。" actions={<DNSProviderDialog trigger={<Button><Plus data-icon="inline-start" />新增 Provider</Button>} onSave={(payload) => save(payload)} />} />
      <Card>
        <CardHeader><CardTitle>服务商账号</CardTitle><CardDescription>当前支持阿里云 DNS，数据结构已预留其他服务商。</CardDescription></CardHeader>
        <CardContent>
          {providers.length === 0 ? <Empty><EmptyHeader><EmptyMedia variant="icon"><CloudCog /></EmptyMedia><EmptyTitle>还没有 DNS Provider</EmptyTitle><EmptyDescription>添加阿里云凭据后即可配置 DNS-01。</EmptyDescription></EmptyHeader></Empty> : (
            <div className="divide-y rounded-lg border">
              {providers.map((provider) => <div key={provider.id} className="flex items-center gap-3 p-3"><CloudCog className="size-4 text-muted-foreground" /><div className="min-w-0 flex-1"><div className="flex items-center gap-2"><p className="truncate text-sm font-medium">{provider.name}</p><Badge variant={provider.enabled ? "secondary" : "outline"}>{provider.enabled ? "启用" : "停用"}</Badge></div><p className="font-mono text-xs text-muted-foreground">{provider.access_key_id_hint} · {provider.region_id}</p></div><DNSProviderDialog provider={provider} trigger={<Button variant="ghost" size="icon" aria-label={`编辑 ${provider.name}`}><Pencil /></Button>} onSave={(payload) => save(payload, provider)} /><Button variant="ghost" size="icon" aria-label={`删除 ${provider.name}`} onClick={() => void remove(provider)}><Trash2 /></Button></div>)}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

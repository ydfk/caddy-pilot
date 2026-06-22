import { Plus } from "lucide-react";
import { Controller, type Control, type FieldErrors } from "react-hook-form";

import type { CertificateProfile, CertificateProfilePayload } from "@/api/certificates";
import type { DNSProvider } from "@/api/dns-providers";
import { CertificateDialog } from "@/components/certificate-dialog";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { SiteFormValues } from "./site-form-data";

type Props = {
  control: Control<SiteFormValues>;
  errors: FieldErrors<SiteFormValues>;
  certificateType: SiteFormValues["certificateType"];
  challengeType: SiteFormValues["acmeChallengeType"];
  profiles: CertificateProfile[];
  providers: DNSProvider[];
  onCreateCertificate: (payload: CertificateProfilePayload) => Promise<CertificateProfile>;
};

export function CertificateOptions({ control, errors, certificateType, challengeType, profiles, providers, onCreateCertificate }: Props) {
  const wildcardProfiles = profiles.filter((profile) => profile.enabled && profile.certificate_type === "wildcard");
  const enabledProviders = providers.filter((provider) => provider.enabled);

  return (
    <div className="grid gap-4 rounded-lg border bg-muted/20 p-4 md:grid-cols-2">
      <Controller control={control} name="certificateType" render={({ field }) => (
        <Field><FieldLabel>证书类型</FieldLabel><Select value={field.value} onValueChange={field.onChange}><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectGroup><SelectItem value="single">域名证书</SelectItem><SelectItem value="wildcard">通配符证书</SelectItem></SelectGroup></SelectContent></Select></Field>
      )} />

      {certificateType === "single" ? <Controller control={control} name="acmeChallengeType" render={({ field }) => (
        <Field><FieldLabel>验证方式</FieldLabel><Select value={field.value} onValueChange={field.onChange}><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectGroup><SelectItem value="http">自动验证（HTTP / TLS-ALPN）</SelectItem><SelectItem value="dns">DNS-01</SelectItem></SelectGroup></SelectContent></Select></Field>
      )} /> : <Controller control={control} name="certificateProfileID" render={({ field }) => (
        <Field data-invalid={Boolean(errors.certificateProfileID) || undefined}>
          <FieldLabel>通配符证书</FieldLabel>
          {wildcardProfiles.length ? <Select value={field.value} onValueChange={field.onChange}><SelectTrigger aria-invalid={Boolean(errors.certificateProfileID)}><SelectValue placeholder="选择系统证书" /></SelectTrigger><SelectContent><SelectGroup>{wildcardProfiles.map((profile) => <SelectItem key={profile.id} value={profile.id}>{profile.name} · {profile.subjects.join(", ")}</SelectItem>)}</SelectGroup></SelectContent></Select> : <FieldDescription>系统中还没有通配符证书配置。</FieldDescription>}
          <CertificateDialog providers={providers} trigger={<Button type="button" variant="outline" size="sm"><Plus data-icon="inline-start" />直接新增证书</Button>} onSave={onCreateCertificate} />
          <FieldError errors={[errors.certificateProfileID]} />
        </Field>
      )} />}

      {certificateType === "single" && challengeType === "dns" ? <Controller control={control} name="dnsProviderID" render={({ field }) => (
        <Field data-invalid={Boolean(errors.dnsProviderID) || undefined}>
          <FieldLabel>DNS Provider</FieldLabel>
          <Select value={field.value} onValueChange={field.onChange}><SelectTrigger aria-invalid={Boolean(errors.dnsProviderID)}><SelectValue placeholder="选择 DNS Provider" /></SelectTrigger><SelectContent><SelectGroup>{enabledProviders.map((provider) => <SelectItem key={provider.id} value={provider.id}>{provider.name} · 阿里云 DNS</SelectItem>)}</SelectGroup></SelectContent></Select>
          <FieldDescription>DNS-01 是验证方式，Provider 决定使用哪个账号。</FieldDescription>
          <FieldError errors={[errors.dnsProviderID]} />
        </Field>
      )} /> : null}
    </div>
  );
}

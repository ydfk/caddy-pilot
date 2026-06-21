import { Controller, type Control, type FieldErrors } from "react-hook-form";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { cn } from "@/lib/utils";
import type { SiteFormValues } from "./site-form-data";

const certificateTypes = [
  { value: "single", label: "域名证书", description: "为站点域名逐个签发" },
  { value: "wildcard", label: "通配符证书", description: "签发 *.example.com" },
] as const;

export function CertificateOptions({
  control,
  errors,
  certificateType,
}: {
  control: Control<SiteFormValues>;
  errors: FieldErrors<SiteFormValues>;
  certificateType: SiteFormValues["certificateType"];
}) {
  return (
    <div className="grid gap-4 rounded-lg border bg-muted/20 p-4">
      <Controller
        control={control}
        name="certificateType"
        render={({ field }) => (
          <Field>
            <FieldLabel>证书类型</FieldLabel>
            <RadioGroup value={field.value} onValueChange={field.onChange} className="grid gap-2 sm:grid-cols-2">
              {certificateTypes.map((type) => (
                <label key={type.value} className={cn("flex cursor-pointer items-start gap-3 rounded-lg border bg-background p-3", field.value === type.value && "border-primary bg-primary/5")}>
                  <RadioGroupItem value={type.value} className="mt-0.5" />
                  <span><span className="block text-sm font-medium">{type.label}</span><span className="block text-xs text-muted-foreground">{type.description}</span></span>
                </label>
              ))}
            </RadioGroup>
          </Field>
        )}
      />

      {certificateType === "wildcard" ? (
        <Controller
          control={control}
          name="certificateDomain"
          render={({ field }) => (
            <Field data-invalid={Boolean(errors.certificateDomain) || undefined}>
              <FieldLabel htmlFor="certificateDomain">通配符证书域名</FieldLabel>
              <Input id="certificateDomain" className="font-mono" placeholder="*.example.com" {...field} />
              <FieldError errors={[errors.certificateDomain]} />
              <FieldDescription>通配符证书必须使用 DNS-01 验证。</FieldDescription>
            </Field>
          )}
        />
      ) : (
        <Controller
          control={control}
          name="acmeChallengeType"
          render={({ field }) => (
            <Field>
              <FieldLabel>证书验证方式</FieldLabel>
              <RadioGroup value={field.value} onValueChange={field.onChange} className="grid gap-2 sm:grid-cols-2">
                <label className={cn("flex cursor-pointer items-start gap-3 rounded-lg border bg-background p-3", field.value === "http" && "border-primary bg-primary/5")}><RadioGroupItem value="http" className="mt-0.5" /><span><span className="block text-sm font-medium">HTTP / TLS-ALPN</span><span className="block text-xs text-muted-foreground">可开放标准端口时使用</span></span></label>
                <label className={cn("flex cursor-pointer items-start gap-3 rounded-lg border bg-background p-3", field.value === "dns" && "border-primary bg-primary/5")}><RadioGroupItem value="dns" className="mt-0.5" /><span><span className="block text-sm font-medium">阿里云 DNS-01</span><span className="block text-xs text-muted-foreground">无需开放 80 端口</span></span></label>
              </RadioGroup>
            </Field>
          )}
        />
      )}

      {(certificateType === "wildcard") ? <DNSHint /> : null}
    </div>
  );
}

export function DNSHint() {
  return <FieldDescription>阿里云凭据通过 ALIYUN_ACCESS_KEY_ID 和 ALIYUN_ACCESS_KEY_SECRET 环境变量提供，不会保存到站点配置。</FieldDescription>;
}

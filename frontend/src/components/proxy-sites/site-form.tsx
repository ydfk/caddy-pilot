import { zodResolver } from "@hookform/resolvers/zod";
import { Braces, Save, Send, X } from "lucide-react";
import { useEffect } from "react";
import { Controller, useForm, type UseFormRegisterReturn } from "react-hook-form";
import { Link } from "react-router-dom";

import { PageHeader } from "@/components/page-header";
import type { BasicAuthCredential, BasicAuthCredentialPayload } from "@/api/basic-auth";
import type { CertificateProfile, CertificateProfilePayload } from "@/api/certificates";
import type { DNSProvider, DNSProviderPayload } from "@/api/dns-providers";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Spinner } from "@/components/ui/spinner";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { SiteCoreOptions, SiteOptions } from "./site-options";
import { matchingWildcardProfile, siteFormSchema, type SiteFormValues } from "./site-form-data";
import { SiteModeOptions } from "./site-mode-options";
import { UpstreamOptions, UpstreamTLSOptions } from "./upstream-options";
import { CredentialSelector } from "./credential-selector";
import { CertificateOptions } from "./certificate-options";
import { StringListField } from "./string-list-field";

type SiteFormProps = {
  mode: "new" | "edit" | "clone";
  values: SiteFormValues;
  pending: boolean;
  previewing: boolean;
  credentials: BasicAuthCredential[];
  certificates: CertificateProfile[];
  dnsProviders: DNSProvider[];
  onCreateCertificate: (payload: CertificateProfilePayload) => Promise<CertificateProfile>;
  onCreateDNSProvider: (payload: DNSProviderPayload) => Promise<DNSProvider>;
  onCreateCredential: (payload: BasicAuthCredentialPayload) => Promise<BasicAuthCredential>;
  onSave: (values: SiteFormValues, publish: boolean) => Promise<void>;
  onPreview: (values: SiteFormValues) => Promise<void>;
};

export function SiteForm({
  mode,
  values,
  pending,
  previewing,
  credentials,
  certificates,
  dnsProviders,
  onCreateCertificate,
  onCreateDNSProvider,
  onCreateCredential,
  onSave,
  onPreview,
}: SiteFormProps) {
  const form = useForm<SiteFormValues>({ resolver: zodResolver(siteFormSchema), values });
  const errors = form.formState.errors;
  const cloneMode = mode === "clone";
  const basicAuthEnabled = form.watch("basicAuthEnabled");
  const siteType = form.watch("siteType");
  const upstreamType = form.watch("upstreamType");
  const enableHTTPS = form.watch("enableHTTPS");
  const certificateType = form.watch("certificateType");
  const acmeChallengeType = form.watch("acmeChallengeType");
  const domains = form.watch("domains");
  const domainKey = domains.map((domain) => domain.trim().toLowerCase()).join("|");
  const title =
    mode === "new" ? "新增代理站点" : mode === "clone" ? "克隆代理站点" : "编辑代理站点";
  const description = cloneMode
    ? "克隆结果固定为停用状态，确认后再手动启用。"
    : "先完成核心配置，其余选项按需调整。保存不会自动发布。";

  useEffect(() => {
    if (mode === "edit" || !enableHTTPS) return;
    const profile = matchingWildcardProfile(domainKey.split("|"), certificates);
    if (!profile) {
      form.setValue("certificateType", "single");
      form.setValue("certificateProfileID", "");
      return;
    }
    form.setValue("certificateType", "wildcard");
    form.setValue("certificateProfileID", profile.id, { shouldValidate: true });
    form.setValue("acmeChallengeType", "dns");
  }, [certificates, domainKey, enableHTTPS, form, mode]);

  return (
    <form className="mx-auto flex w-full max-w-6xl flex-col gap-4">
      <PageHeader eyebrow="ROUTES / EDITOR" title={title} description={description} />

      <Card className="border-primary/20 shadow-sm">
        <CardHeader>
          <CardTitle>核心配置</CardTitle>
          <CardDescription>选择站点处理方式，再配置域名、目录或上游。</CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup className="gap-5">
            <SiteModeOptions control={form.control} errors={errors} siteType={siteType} />
            {siteType !== "static" ? <UpstreamOptions control={form.control} /> : null}
            <div className="grid gap-4 md:grid-cols-2">
              <Controller
                control={form.control}
                name="domains"
                render={({ field }) => (
                  <StringListField
                    id="domains"
                    label="域名"
                    value={field.value}
                    onChange={field.onChange}
                    placeholder="example.com"
                    addLabel="添加域名"
                    description="多个域名共用这一套代理和访问控制配置。"
                    error={errors.domains?.message}
                  />
                )}
              />
              {siteType !== "static" ? (
                <Controller
                  control={form.control}
                  name="upstreams"
                  render={({ field }) => (
                    <StringListField
                      id="upstreams"
                      label={siteType === "spa" ? "API 上游地址" : "上游地址"}
                      value={field.value}
                      onChange={field.onChange}
                      placeholder={upstreamType === "unix" ? "/run/app.sock" : "127.0.0.1:3000"}
                      addLabel="添加上游"
                      description={
                        upstreamType === "unix"
                          ? "每项一个套接字路径。"
                          : "多个上游由 Caddy 自动负载分配。每项填写 host:port。"
                      }
                      error={errors.upstreams?.message}
                    />
                  )}
                />
              ) : null}
            </div>

            <Field data-invalid={Boolean(errors.description) || undefined}>
              <FieldLabel htmlFor="description">备注说明</FieldLabel>
              <Textarea
                id="description"
                rows={2}
                placeholder="可选，例如用途、维护人或部署位置"
                {...form.register("description")}
              />
              <FieldDescription>仅用于管理展示，不会写入 Caddy 配置。</FieldDescription>
              <FieldError>{errors.description?.message}</FieldError>
            </Field>

            {siteType !== "static" && upstreamType === "https" ? (
              <UpstreamTLSOptions control={form.control} />
            ) : null}

            <SiteCoreOptions control={form.control} cloneMode={cloneMode} />
            {enableHTTPS ? (
              <CertificateOptions
                control={form.control}
                errors={errors}
                certificateType={certificateType}
                challengeType={acmeChallengeType}
                profiles={certificates}
                providers={dnsProviders}
                onCreateCertificate={async (payload) => {
                  const created = await onCreateCertificate(payload);
                  form.setValue("certificateProfileID", created.id, { shouldValidate: true });
                  return created;
                }}
                onCreateDNSProvider={onCreateDNSProvider}
              />
            ) : null}
          </FieldGroup>
        </CardContent>
      </Card>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>传输与日志</CardTitle>
            <CardDescription>
              WebSocket 由 Caddy 自动支持，这里只保留需要显式控制的能力。
            </CardDescription>
          </CardHeader>
          <CardContent>
            <SiteOptions control={form.control} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>访问控制</CardTitle>
            <CardDescription>限制来源地址或为站点添加基础认证。</CardDescription>
          </CardHeader>
          <CardContent>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="allowedIPs">IP 白名单</FieldLabel>
                <Textarea
                  id="allowedIPs"
                  rows={3}
                  className="font-mono"
                  placeholder={"127.0.0.1\n10.0.0.0/8"}
                  {...form.register("allowedIPs")}
                />
                <FieldDescription>留空表示不限制来源地址。</FieldDescription>
              </Field>
              <Controller
                control={form.control}
                name="basicAuthEnabled"
                render={({ field }) => (
                  <Field orientation="horizontal">
                    <div className="flex-1">
                      <FieldLabel htmlFor="basicAuthEnabled">Basic Auth</FieldLabel>
                      <FieldDescription>从统一密码本选择允许登录的账号。</FieldDescription>
                    </div>
                    <Switch
                      id="basicAuthEnabled"
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </Field>
                )}
              />
              {basicAuthEnabled ? (
                <Controller
                  control={form.control}
                  name="basicAuthCredentialIDs"
                  render={({ field }) => (
                    <CredentialSelector
                      credentials={credentials}
                      selected={field.value}
                      onChange={field.onChange}
                      onCreate={onCreateCredential}
                    />
                  )}
                />
              ) : null}
            </FieldGroup>
          </CardContent>
        </Card>
      </div>

      <Card>
        <Accordion type="single" collapsible>
          <AccordionItem value="advanced" className="border-0 px-6">
            <AccordionTrigger>高级配置 · Header 与扩展 JSON</AccordionTrigger>
            <AccordionContent>
              <FieldGroup>
                <div className="grid gap-4 md:grid-cols-2">
                  <JSONField
                    id="requestHeaders"
                    label="请求头 JSON"
                    error={errors.requestHeaders?.message}
                    register={form.register("requestHeaders")}
                  />
                  <JSONField
                    id="responseHeaders"
                    label="响应头 JSON"
                    error={errors.responseHeaders?.message}
                    register={form.register("responseHeaders")}
                  />
                </div>
                <JSONField
                  id="advancedJSON"
                  label="扩展 JSON"
                  error={errors.advancedJSON?.message}
                  register={form.register("advancedJSON")}
                />
                <FieldDescription>扩展 JSON 当前仅保存，不合并到生成配置。</FieldDescription>
              </FieldGroup>
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </Card>

      <div className="sticky bottom-2 flex flex-wrap justify-end gap-2 rounded-xl border bg-background/95 p-2 shadow-lg backdrop-blur">
        <Button type="button" variant="ghost" asChild>
          <Link to="/proxy-sites">
            <X data-icon="inline-start" /> 取消
          </Link>
        </Button>
        <Button
          type="button"
          variant="outline"
          disabled={previewing}
          onClick={form.handleSubmit((data) => void onPreview(data))}
        >
          {previewing ? <Spinner data-icon="inline-start" /> : <Braces data-icon="inline-start" />}
          预览配置
        </Button>
        <Button
          type="button"
          variant="secondary"
          disabled={pending}
          onClick={form.handleSubmit((data) => onSave(data, false))}
        >
          {pending ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
          保存
        </Button>
        <Button
          type="button"
          disabled={pending}
          onClick={form.handleSubmit((data) => onSave(data, true))}
        >
          {pending ? <Spinner data-icon="inline-start" /> : <Send data-icon="inline-start" />}
          保存并发布
        </Button>
      </div>
    </form>
  );
}

function JSONField({
  id,
  label,
  error,
  register,
}: {
  id: string;
  label: string;
  error?: string;
  register: UseFormRegisterReturn;
}) {
  return (
    <Field data-invalid={Boolean(error) || undefined}>
      <FieldLabel htmlFor={id}>{label}</FieldLabel>
      <Textarea
        id={id}
        rows={4}
        className="font-mono"
        {...register}
        aria-invalid={Boolean(error)}
      />
      <FieldError>{error}</FieldError>
    </Field>
  );
}

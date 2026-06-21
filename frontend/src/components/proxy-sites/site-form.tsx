import { zodResolver } from "@hookform/resolvers/zod";
import { Braces, Save, Send, X } from "lucide-react";
import { Controller, useForm, type UseFormRegisterReturn } from "react-hook-form";
import { Link } from "react-router-dom";

import { PageHeader } from "@/components/page-header";
import type { BasicAuthCredential } from "@/api/basic-auth";
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
import { siteFormSchema, type SiteFormValues } from "./site-form-data";
import { UpstreamOptions, UpstreamTLSOptions } from "./upstream-options";
import { CredentialSelector } from "./credential-selector";

type SiteFormProps = {
  mode: "new" | "edit" | "clone";
  values: SiteFormValues;
  pending: boolean;
  credentials: BasicAuthCredential[];
  onSave: (values: SiteFormValues, publish: boolean) => Promise<void>;
  onPreview: (values: SiteFormValues) => void;
};

export function SiteForm({ mode, values, pending, credentials, onSave, onPreview }: SiteFormProps) {
  const form = useForm<SiteFormValues>({ resolver: zodResolver(siteFormSchema), values });
  const errors = form.formState.errors;
  const cloneMode = mode === "clone";
  const basicAuthEnabled = form.watch("basicAuthEnabled");
  const upstreamType = form.watch("upstreamType");
  const title =
    mode === "new" ? "新增代理站点" : mode === "clone" ? "克隆代理站点" : "编辑代理站点";
  const description = cloneMode
    ? "克隆结果固定为停用状态，确认后再手动启用。"
    : "先完成核心配置，其余选项按需调整。保存不会自动发布。";

  return (
    <form className="mx-auto flex w-full max-w-6xl flex-col gap-4">
      <PageHeader eyebrow="ROUTES / EDITOR" title={title} description={description} />

      <Card className="border-primary/20 shadow-sm">
        <CardHeader>
          <CardTitle>核心配置</CardTitle>
          <CardDescription>域名和上游是代理站点的核心配置。</CardDescription>
        </CardHeader>
        <CardContent>
          <FieldGroup className="gap-5">
            <UpstreamOptions control={form.control} />
            <div className="grid gap-4 md:grid-cols-2">
              <Field data-invalid={Boolean(errors.domains) || undefined}>
                <FieldLabel htmlFor="domains">域名</FieldLabel>
                <Textarea
                  id="domains"
                  rows={3}
                  className="font-mono"
                  placeholder={"example.com\nwww.example.com"}
                  {...form.register("domains")}
                  aria-invalid={Boolean(errors.domains)}
                />
                <FieldError errors={[errors.domains]} />
              </Field>
              <Field data-invalid={Boolean(errors.upstreams) || undefined}>
                <FieldLabel htmlFor="upstreams">上游</FieldLabel>
                <Textarea
                  id="upstreams"
                  rows={3}
                  className="font-mono"
                  placeholder={"127.0.0.1:3000\n10.0.0.8:8080"}
                  {...form.register("upstreams")}
                  aria-invalid={Boolean(errors.upstreams)}
                />
                <FieldDescription>
                  {upstreamType === "unix" ? "每行一个套接字绝对路径。" : "每行一个 host:port 地址。"}
                </FieldDescription>
                <FieldError errors={[errors.upstreams]} />
              </Field>
            </div>

            {upstreamType === "https" ? <UpstreamTLSOptions control={form.control} /> : null}

            <SiteCoreOptions control={form.control} cloneMode={cloneMode} />
          </FieldGroup>
        </CardContent>
      </Card>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>传输与日志</CardTitle>
            <CardDescription>直接开关常用代理能力。</CardDescription>
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
        <Button type="button" variant="outline" onClick={form.handleSubmit(onPreview)}>
          <Braces data-icon="inline-start" /> 预览 Caddy JSON
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

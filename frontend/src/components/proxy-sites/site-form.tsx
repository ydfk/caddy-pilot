import { zodResolver } from "@hookform/resolvers/zod";
import { Braces, Save, Send, X } from "lucide-react";
import { Controller, useForm, type UseFormRegisterReturn } from "react-hook-form";
import { Link } from "react-router-dom";

import { PageHeader } from "@/components/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@/components/ui/field";
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
import { Textarea } from "@/components/ui/textarea";
import { SiteOptions } from "./site-options";
import { siteFormSchema, type SiteFormValues } from "./site-form-data";

type SiteFormProps = {
  mode: "new" | "edit" | "clone";
  values: SiteFormValues;
  pending: boolean;
  onSave: (values: SiteFormValues, publish: boolean) => Promise<void>;
  onPreview: (values: SiteFormValues) => void;
};

export function SiteForm({ mode, values, pending, onSave, onPreview }: SiteFormProps) {
  const form = useForm<SiteFormValues>({ resolver: zodResolver(siteFormSchema), values });
  const errors = form.formState.errors;
  const cloneMode = mode === "clone";
  const basicAuthEnabled = form.watch("basicAuthEnabled");

  const title =
    mode === "new" ? "新增代理站点" : mode === "clone" ? "克隆代理站点" : "编辑代理站点";
  const description = cloneMode
    ? "克隆结果固定为停用状态，确认后再手动启用。"
    : "配置只保存在业务数据库中，除非选择保存并发布。";

  return (
    <form className="flex flex-col gap-4">
      <PageHeader eyebrow="ROUTES / EDITOR" title={title} description={description} />

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1.5fr)_minmax(18rem,0.5fr)]">
        <div className="flex flex-col gap-4">
          <Card>
            <CardHeader>
              <CardTitle>基础信息</CardTitle>
              <CardDescription>用于识别站点，不会写入 Caddy 匹配器。</CardDescription>
            </CardHeader>
            <CardContent>
              <FieldGroup>
                <Field data-invalid={Boolean(errors.name) || undefined}>
                  <FieldLabel htmlFor="name">名称</FieldLabel>
                  <Input id="name" {...form.register("name")} aria-invalid={Boolean(errors.name)} />
                  <FieldError errors={[errors.name]} />
                </Field>
                <Field>
                  <FieldLabel htmlFor="description">描述</FieldLabel>
                  <Textarea id="description" rows={2} {...form.register("description")} />
                </Field>
              </FieldGroup>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>路由目标</CardTitle>
              <CardDescription>
                每行一个值；上游使用 host:port，例如 127.0.0.1:3000。
              </CardDescription>
            </CardHeader>
            <CardContent>
              <FieldGroup>
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
                  <FieldError errors={[errors.upstreams]} />
                </Field>
              </FieldGroup>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>请求与访问控制</CardTitle>
              <CardDescription>Header 与 Basic Auth 用户使用 JSON 对象。</CardDescription>
            </CardHeader>
            <CardContent>
              <FieldGroup>
                <FieldGroup className="grid gap-4 md:grid-cols-2">
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
                </FieldGroup>
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
                <FieldSet>
                  <FieldLegend>Basic Auth</FieldLegend>
                  <Controller
                    control={form.control}
                    name="basicAuthEnabled"
                    render={({ field }) => (
                      <Field>
                        <FieldLabel htmlFor="basicAuthEnabled">认证模式</FieldLabel>
                        <Select
                          value={field.value ? "basic" : "none"}
                          onValueChange={(value) => field.onChange(value === "basic")}
                        >
                          <SelectTrigger id="basicAuthEnabled">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectGroup>
                              <SelectItem value="none">不启用认证</SelectItem>
                              <SelectItem value="basic">启用 Basic Auth</SelectItem>
                            </SelectGroup>
                          </SelectContent>
                        </Select>
                        <FieldDescription>密码值必须是 Caddy 支持的 bcrypt 哈希。</FieldDescription>
                      </Field>
                    )}
                  />
                  {basicAuthEnabled ? (
                    <JSONField
                      id="basicAuthUsers"
                      label="用户 JSON"
                      error={errors.basicAuthUsers?.message}
                      register={form.register("basicAuthUsers")}
                    />
                  ) : null}
                </FieldSet>
                <Field data-invalid={Boolean(errors.advancedJSON) || undefined}>
                  <FieldLabel htmlFor="advancedJSON">advanced_json</FieldLabel>
                  <Textarea
                    id="advancedJSON"
                    rows={4}
                    className="font-mono"
                    placeholder="{}"
                    {...form.register("advancedJSON")}
                    aria-invalid={Boolean(errors.advancedJSON)}
                  />
                  <FieldDescription>MVP 仅保存此字段，不合并到生成配置。</FieldDescription>
                  <FieldError errors={[errors.advancedJSON]} />
                </Field>
              </FieldGroup>
            </CardContent>
          </Card>
        </div>

        <Card className="h-fit xl:sticky xl:top-18">
          <CardHeader>
            <CardTitle>运行行为</CardTitle>
            <CardDescription>控制站点进入发布配置后的行为。</CardDescription>
          </CardHeader>
          <CardContent>
            <SiteOptions control={form.control} cloneMode={cloneMode} />
          </CardContent>
        </Card>
      </div>

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

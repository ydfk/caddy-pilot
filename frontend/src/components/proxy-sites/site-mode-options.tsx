import { Controller, type Control, type FieldErrors } from "react-hook-form";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import type { SiteFormValues } from "./site-form-data";

type Props = {
  control: Control<SiteFormValues>;
  errors: FieldErrors<SiteFormValues>;
  siteType: SiteFormValues["siteType"];
};

export function SiteModeOptions({ control, errors, siteType }: Props) {
  return (
    <div className="grid gap-3 md:grid-cols-3">
      <Controller
        control={control}
        name="siteType"
        render={({ field }) => (
          <Field>
            <FieldLabel>站点类型</FieldLabel>
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="proxy">反向代理</SelectItem>
                  <SelectItem value="static">静态文件目录</SelectItem>
                  <SelectItem value="spa">SPA 静态前端 + API 代理</SelectItem>
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        )}
      />

      {siteType !== "proxy" ? (
        <Controller
          control={control}
          name="rootPath"
          render={({ field }) => (
            <Field data-invalid={Boolean(errors.rootPath) || undefined}>
              <FieldLabel htmlFor="rootPath">静态文件根目录</FieldLabel>
              <Input
                id="rootPath"
                className="font-mono"
                placeholder="/var/www/app/dist"
                {...field}
                aria-invalid={Boolean(errors.rootPath)}
              />
              <FieldError errors={[errors.rootPath]} />
            </Field>
          )}
        />
      ) : null}

      {siteType === "spa" ? (
        <Controller
          control={control}
          name="apiPath"
          render={({ field }) => (
            <Field data-invalid={Boolean(errors.apiPath) || undefined}>
              <FieldLabel htmlFor="apiPath">API 路径</FieldLabel>
              <Input id="apiPath" className="font-mono" placeholder="/api/*" {...field} />
              <FieldError errors={[errors.apiPath]} />
            </Field>
          )}
        />
      ) : null}

      {siteType !== "proxy" ? (
        <div className="grid gap-2 md:col-span-3 md:grid-cols-2">
          <ModeSwitch
            control={control}
            name="enableSecurityHeaders"
            label="安全响应头"
            description="写入 nosniff 与严格来源策略"
          />
          <ModeSwitch
            control={control}
            name="enableAssetCache"
            label="静态资源缓存"
            description="assets 长期缓存，index.html 禁用缓存"
          />
        </div>
      ) : null}
    </div>
  );
}

function ModeSwitch({
  control,
  name,
  label,
  description,
}: {
  control: Control<SiteFormValues>;
  name: "enableSecurityHeaders" | "enableAssetCache";
  label: string;
  description: string;
}) {
  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <Field orientation="horizontal" className="rounded-md border bg-background px-3 py-2.5">
          <div className="flex-1">
            <FieldLabel htmlFor={name}>{label}</FieldLabel>
            <FieldDescription>{description}</FieldDescription>
          </div>
          <Switch id={name} checked={field.value} onCheckedChange={field.onChange} />
        </Field>
      )}
    />
  );
}

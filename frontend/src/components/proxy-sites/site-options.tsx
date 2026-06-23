import { Controller, type Control, useController } from "react-hook-form";

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
import { Switch } from "@/components/ui/switch";
import type { SiteFormValues } from "./site-form-data";

type SiteOptionsProps = { control: Control<SiteFormValues>; cloneMode: boolean };

const transportOptions = [
  { name: "enableGzip", label: "响应压缩", description: "启用 gzip 与 zstd" },
  { name: "enableLog", label: "访问日志", description: "记录站点请求" },
] as const;

export function SiteCoreOptions({ control, cloneMode }: SiteOptionsProps) {
  const https = useController({ control, name: "enableHTTPS" });
  const forceHTTPS = useController({ control, name: "forceHTTPS" });
  const httpsMode = !https.field.value ? "off" : forceHTTPS.field.value ? "redirect" : "on";

  function setHTTPSMode(value: string) {
    https.field.onChange(value !== "off");
    forceHTTPS.field.onChange(value === "redirect");
  }

  return (
    <FieldGroup className="gap-5">
      <Controller
        control={control}
        name="enabled"
        render={({ field }) => (
          <Field orientation="horizontal" data-disabled={cloneMode || undefined}>
            <div className="flex-1">
              <FieldLabel htmlFor="enabled">启用站点</FieldLabel>
              <FieldDescription>
                {cloneMode ? "克隆站点固定为停用。" : "启用后会进入下一次发布配置。"}
              </FieldDescription>
            </div>
            <Switch
              id="enabled"
              checked={field.value}
              disabled={cloneMode}
              onCheckedChange={field.onChange}
            />
          </Field>
        )}
      />

      <Field>
        <FieldLabel>HTTPS 模式</FieldLabel>
        <Select value={httpsMode} onValueChange={setHTTPSMode}>
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="redirect">强制 HTTPS · HTTP 自动跳转</SelectItem>
              <SelectItem value="on">仅 HTTPS · 不创建 HTTP 跳转</SelectItem>
              <SelectItem value="off">关闭 HTTPS · 只监听 HTTP</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </Field>

      {httpsMode === "redirect" ? (
        <Controller
          control={control}
          name="httpsRedirectPort"
          render={({ field }) => (
            <Field>
              <FieldLabel htmlFor="httpsRedirectPort">HTTPS 跳转端口</FieldLabel>
              <Input
                id="httpsRedirectPort"
                type="number"
                min={1}
                max={65535}
                value={field.value}
                onChange={(event) => field.onChange(Number(event.target.value))}
              />
              <FieldDescription>
                标准端口填写 443，跳转地址不会附加端口；映射到 8443 等外部端口时填写对应端口。
              </FieldDescription>
            </Field>
          )}
        />
      ) : null}
    </FieldGroup>
  );
}

export function SiteOptions({ control }: Pick<SiteOptionsProps, "control">) {
  return (
    <FieldGroup className="gap-1">
      {transportOptions.map((option) => (
        <Controller
          key={option.name}
          control={control}
          name={option.name}
          render={({ field }) => (
            <Field orientation="horizontal" className="rounded-lg px-1 py-3">
              <div className="flex-1">
                <FieldLabel htmlFor={option.name}>{option.label}</FieldLabel>
                <FieldDescription>{option.description}</FieldDescription>
              </div>
              <Switch id={option.name} checked={field.value} onCheckedChange={field.onChange} />
            </Field>
          )}
        />
      ))}
    </FieldGroup>
  );
}

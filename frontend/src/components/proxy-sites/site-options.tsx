import { Controller, type Control, useController } from "react-hook-form";

import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import type { SiteFormValues } from "./site-form-data";

type SiteOptionsProps = { control: Control<SiteFormValues>; cloneMode: boolean };

const httpsModes = [
  { value: "redirect", label: "强制 HTTPS", description: "HTTP 自动跳转，推荐" },
  { value: "on", label: "仅 HTTPS", description: "不创建 HTTP 跳转" },
  { value: "off", label: "关闭 HTTPS", description: "只监听 HTTP" },
] as const;

const transportOptions = [
  { name: "enableWS", label: "WebSocket", description: "允许协议升级" },
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
        <RadioGroup
          value={httpsMode}
          onValueChange={setHTTPSMode}
          className="grid gap-2 md:grid-cols-3"
        >
          {httpsModes.map((mode) => (
            <label
              key={mode.value}
              className={cn(
                "flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors hover:bg-accent/50",
                httpsMode === mode.value && "border-primary bg-primary/5"
              )}
            >
              <RadioGroupItem value={mode.value} className="mt-0.5" />
              <span>
                <span className="block text-sm font-medium">{mode.label}</span>
                <span className="block text-xs text-muted-foreground">{mode.description}</span>
              </span>
            </label>
          ))}
        </RadioGroup>
      </Field>
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

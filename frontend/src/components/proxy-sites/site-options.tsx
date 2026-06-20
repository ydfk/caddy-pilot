import { Controller, type Control } from "react-hook-form";

import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@/components/ui/field";
import { Switch } from "@/components/ui/switch";
import type { SiteFormValues } from "./site-form-data";

type SiteOptionsProps = { control: Control<SiteFormValues>; cloneMode: boolean };

const options: Array<{
  name: keyof Pick<
    SiteFormValues,
    "enabled" | "enableHTTPS" | "forceHTTPS" | "enableWS" | "enableGzip" | "enableLog"
  >;
  label: string;
  description: string;
}> = [
  { name: "enabled", label: "启用站点", description: "只有启用的站点会进入下一次发布。" },
  { name: "enableHTTPS", label: "启用 HTTPS", description: "在 :443 提供站点并启用自动证书。" },
  { name: "forceHTTPS", label: "强制 HTTPS", description: "将 HTTP 请求永久跳转到 HTTPS。" },
  {
    name: "enableWS",
    label: "启用 WebSocket",
    description: "Caddy reverse_proxy 默认支持协议升级。",
  },
  { name: "enableGzip", label: "启用 gzip / zstd", description: "对可压缩响应启用内容编码。" },
  { name: "enableLog", label: "启用访问日志", description: "保留站点日志开关供运行配置使用。" },
];

export function SiteOptions({ control, cloneMode }: SiteOptionsProps) {
  return (
    <FieldSet>
      <FieldLegend>运行选项</FieldLegend>
      <FieldGroup className="gap-3">
        {options.map((option) => (
          <Controller
            key={option.name}
            control={control}
            name={option.name}
            render={({ field }) => {
              const disabled = cloneMode && option.name === "enabled";
              return (
                <Field orientation="horizontal" data-disabled={disabled || undefined}>
                  <div className="flex-1">
                    <FieldLabel htmlFor={option.name}>{option.label}</FieldLabel>
                    <FieldDescription>{option.description}</FieldDescription>
                  </div>
                  <Switch
                    id={option.name}
                    checked={field.value}
                    disabled={disabled}
                    onCheckedChange={field.onChange}
                  />
                </Field>
              );
            }}
          />
        ))}
      </FieldGroup>
    </FieldSet>
  );
}

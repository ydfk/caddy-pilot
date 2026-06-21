import { Controller, type Control, useController } from "react-hook-form";

import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSet,
} from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { SiteFormValues } from "./site-form-data";

type SiteOptionsProps = { control: Control<SiteFormValues>; cloneMode: boolean };

const options = [
  { name: "enableWS", label: "WebSocket", on: "启用协议升级", off: "关闭协议升级" },
  { name: "enableGzip", label: "响应压缩", on: "gzip + zstd", off: "不压缩" },
  { name: "enableLog", label: "访问日志", on: "记录访问日志", off: "不记录" },
] as const;

export function SiteOptions({ control, cloneMode }: SiteOptionsProps) {
  const https = useController({ control, name: "enableHTTPS" });
  const forceHTTPS = useController({ control, name: "forceHTTPS" });
  const httpsMode = !https.field.value ? "off" : forceHTTPS.field.value ? "redirect" : "on";

  function setHTTPSMode(value: string) {
    https.field.onChange(value !== "off");
    forceHTTPS.field.onChange(value === "redirect");
  }

  return (
    <FieldSet>
      <FieldLegend>运行选项</FieldLegend>
      <FieldGroup className="gap-4">
        <Controller
          control={control}
          name="enabled"
          render={({ field }) => (
            <Field data-disabled={cloneMode || undefined}>
              <FieldLabel htmlFor="enabled">站点状态</FieldLabel>
              <Select
                value={field.value ? "enabled" : "disabled"}
                disabled={cloneMode}
                onValueChange={(value) => field.onChange(value === "enabled")}
              >
                <SelectTrigger id="enabled">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="enabled">启用并进入下次发布</SelectItem>
                    <SelectItem value="disabled">停用，不进入配置</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
              <FieldDescription>
                {cloneMode ? "克隆站点固定为停用。" : "保存后仍需发布才会生效。"}
              </FieldDescription>
            </Field>
          )}
        />

        <Field>
          <FieldLabel htmlFor="httpsMode">HTTPS 模式</FieldLabel>
          <Select value={httpsMode} onValueChange={setHTTPSMode}>
            <SelectTrigger id="httpsMode">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="redirect">HTTPS + HTTP 强制跳转</SelectItem>
                <SelectItem value="on">仅启用 HTTPS</SelectItem>
                <SelectItem value="off">关闭 HTTPS</SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
        </Field>

        {options.map((option) => (
          <Controller
            key={option.name}
            control={control}
            name={option.name}
            render={({ field }) => (
              <Field>
                <FieldLabel htmlFor={option.name}>{option.label}</FieldLabel>
                <Select
                  value={field.value ? "on" : "off"}
                  onValueChange={(value) => field.onChange(value === "on")}
                >
                  <SelectTrigger id={option.name}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectGroup>
                      <SelectItem value="on">{option.on}</SelectItem>
                      <SelectItem value="off">{option.off}</SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
              </Field>
            )}
          />
        ))}
      </FieldGroup>
    </FieldSet>
  );
}

import { Controller, type Control } from "react-hook-form";

import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import type { SiteFormValues } from "./site-form-data";

const upstreamTypes = [
  { value: "http", label: "HTTP", description: "常规明文 HTTP 服务" },
  { value: "https", label: "HTTPS", description: "使用 TLS 连接上游" },
  { value: "h2c", label: "h2c", description: "明文 HTTP/2，例如 gRPC" },
  { value: "unix", label: "Unix Socket", description: "连接本机套接字文件" },
] as const;

export function UpstreamOptions({ control }: { control: Control<SiteFormValues> }) {
  return (
    <Controller
      control={control}
      name="upstreamType"
      render={({ field }) => (
        <Field>
          <FieldLabel>上游类型</FieldLabel>
          <RadioGroup
            value={field.value}
            onValueChange={field.onChange}
            className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4"
          >
            {upstreamTypes.map((type) => (
              <label
                key={type.value}
                className={cn(
                  "flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors hover:bg-accent/50",
                  field.value === type.value && "border-primary bg-primary/5"
                )}
              >
                <RadioGroupItem value={type.value} className="mt-0.5" />
                <span>
                  <span className="block text-sm font-medium">{type.label}</span>
                  <span className="block text-xs text-muted-foreground">{type.description}</span>
                </span>
              </label>
            ))}
          </RadioGroup>
        </Field>
      )}
    />
  );
}

export function UpstreamTLSOptions({ control }: { control: Control<SiteFormValues> }) {
  return (
    <div className="grid gap-4 rounded-lg border bg-muted/20 p-4 md:grid-cols-2">
      <Controller
        control={control}
        name="upstreamTLSServerName"
        render={({ field }) => (
          <Field>
            <FieldLabel htmlFor="upstreamTLSServerName">TLS Server Name</FieldLabel>
            <Input id="upstreamTLSServerName" placeholder="backend.example.com" {...field} />
            <FieldDescription>上游地址是 IP 时，可在这里指定证书域名。</FieldDescription>
          </Field>
        )}
      />
      <Controller
        control={control}
        name="upstreamTLSInsecureSkipVerify"
        render={({ field }) => (
          <Field orientation="horizontal">
            <div className="flex-1">
              <FieldLabel htmlFor="upstreamTLSInsecureSkipVerify">跳过证书校验</FieldLabel>
              <FieldDescription>仅用于可信内网的自签名证书。</FieldDescription>
            </div>
            <Switch
              id="upstreamTLSInsecureSkipVerify"
              checked={field.value}
              onCheckedChange={field.onChange}
            />
          </Field>
        )}
      />
    </div>
  );
}

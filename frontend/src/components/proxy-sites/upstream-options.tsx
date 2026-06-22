import { Controller, type Control } from "react-hook-form";

import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import type { SiteFormValues } from "./site-form-data";

export function UpstreamOptions({ control }: { control: Control<SiteFormValues> }) {
  return (
    <Controller
      control={control}
      name="upstreamType"
      render={({ field }) => (
        <Field>
          <FieldLabel>上游类型</FieldLabel>
          <Select value={field.value} onValueChange={field.onChange}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent><SelectGroup>
              <SelectItem value="http">HTTP · 常规明文服务</SelectItem>
              <SelectItem value="https">HTTPS · TLS 上游</SelectItem>
              <SelectItem value="h2c">h2c · 明文 HTTP/2 / gRPC</SelectItem>
              <SelectItem value="unix">Unix Socket · 套接字文件</SelectItem>
            </SelectGroup></SelectContent>
          </Select>
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

import { z } from "zod";

import type { ProxySite, ProxySitePayload } from "@/api/proxy-sites";

const jsonObject = z.string().refine(isJSONObject, "请输入有效的 JSON 对象");
const optionalJSON = z
  .string()
  .refine((value) => !value.trim() || isJSON(value), "请输入有效 JSON");

export const siteFormSchema = z.object({
  domains: z.string().refine((value) => splitLines(value).length > 0, "至少填写一个域名"),
  upstreams: z.string().refine((value) => splitLines(value).length > 0, "至少填写一个上游"),
  enabled: z.boolean(),
  enableHTTPS: z.boolean(),
  forceHTTPS: z.boolean(),
  enableWS: z.boolean(),
  enableGzip: z.boolean(),
  enableLog: z.boolean(),
  requestHeaders: jsonObject,
  responseHeaders: jsonObject,
  allowedIPs: z.string(),
  basicAuthEnabled: z.boolean(),
  basicAuthUsers: jsonObject,
  advancedJSON: optionalJSON,
});

export type SiteFormValues = z.infer<typeof siteFormSchema>;

export const defaultSiteValues: SiteFormValues = {
  domains: "",
  upstreams: "",
  enabled: false,
  enableHTTPS: true,
  forceHTTPS: true,
  enableWS: true,
  enableGzip: true,
  enableLog: false,
  requestHeaders: "{}",
  responseHeaders: "{}",
  allowedIPs: "",
  basicAuthEnabled: false,
  basicAuthUsers: "{}",
  advancedJSON: "",
};

export function formValuesFromSite(site: ProxySite, clone: boolean): SiteFormValues {
  return {
    domains: site.domains.join("\n"),
    upstreams: site.upstreams.join("\n"),
    enabled: clone ? false : site.enabled,
    enableHTTPS: site.enable_https,
    forceHTTPS: site.force_https,
    enableWS: site.enable_ws,
    enableGzip: site.enable_gzip,
    enableLog: site.enable_log,
    requestHeaders: JSON.stringify(site.request_headers, null, 2),
    responseHeaders: JSON.stringify(site.response_headers, null, 2),
    allowedIPs: site.allowed_ips.join("\n"),
    basicAuthEnabled: site.basic_auth_enabled,
    basicAuthUsers: JSON.stringify(site.basic_auth_users, null, 2),
    advancedJSON: site.advanced_json,
  };
}

export function payloadFromForm(values: SiteFormValues, forceDisabled = false): ProxySitePayload {
  return {
    name: "",
    description: "",
    domains: splitLines(values.domains),
    upstreams: splitLines(values.upstreams),
    enabled: forceDisabled ? false : values.enabled,
    enable_https: values.enableHTTPS,
    force_https: values.forceHTTPS,
    enable_ws: values.enableWS,
    enable_gzip: values.enableGzip,
    enable_log: values.enableLog,
    request_headers: JSON.parse(values.requestHeaders) as Record<string, string>,
    response_headers: JSON.parse(values.responseHeaders) as Record<string, string>,
    allowed_ips: splitLines(values.allowedIPs),
    basic_auth_enabled: values.basicAuthEnabled,
    basic_auth_users: JSON.parse(values.basicAuthUsers) as Record<string, string>,
    advanced_json: values.advancedJSON.trim(),
  };
}

export function draftPreview(values: SiteFormValues) {
  return {
    match: [{ host: splitLines(values.domains) }],
    handle: [
      {
        handler: "reverse_proxy",
        upstreams: splitLines(values.upstreams).map((dial) => ({ dial })),
      },
    ],
    enabled: values.enabled,
    note: "草稿片段；完整配置会在发布前由后端生成并注入管理入口。",
  };
}

function splitLines(value: string) {
  return value
    .split(/\r?\n/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function isJSONObject(value: string) {
  try {
    const parsed: unknown = JSON.parse(value);
    return typeof parsed === "object" && parsed !== null && !Array.isArray(parsed);
  } catch {
    return false;
  }
}

function isJSON(value: string) {
  try {
    JSON.parse(value);
    return true;
  } catch {
    return false;
  }
}

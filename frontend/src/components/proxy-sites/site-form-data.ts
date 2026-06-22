import { z } from "zod";

import type { ProxySite, ProxySitePayload } from "@/api/proxy-sites";

const jsonObject = z.string().refine(isJSONObject, "请输入有效的 JSON 对象");
const optionalJSON = z
  .string()
  .refine((value) => !value.trim() || isJSON(value), "请输入有效 JSON");

export const siteFormSchema = z.object({
  domains: z.array(z.string()).refine((value) => compactItems(value).length > 0, "至少填写一个域名"),
  upstreams: z.array(z.string()).refine((value) => compactItems(value).length > 0, "至少填写一个上游"),
  upstreamType: z.enum(["http", "https", "h2c", "unix"]),
  upstreamTLSServerName: z.string().max(253),
  upstreamTLSInsecureSkipVerify: z.boolean(),
  enabled: z.boolean(),
  enableHTTPS: z.boolean(),
  forceHTTPS: z.boolean(),
  certificateType: z.enum(["single", "wildcard"]),
  certificateDomain: z.string().max(253),
  acmeChallengeType: z.enum(["http", "dns"]),
  dnsProvider: z.literal("alidns"),
  dnsProviderID: z.string(),
  certificateProfileID: z.string(),
  enableWS: z.boolean(),
  enableGzip: z.boolean(),
  enableLog: z.boolean(),
  requestHeaders: jsonObject,
  responseHeaders: jsonObject,
  allowedIPs: z.string(),
  basicAuthEnabled: z.boolean(),
  basicAuthCredentialIDs: z.array(z.string()),
  advancedJSON: optionalJSON,
}).superRefine((values, context) => {
  if (values.certificateType === "wildcard" && !values.certificateProfileID) {
    context.addIssue({
      code: "custom",
      path: ["certificateProfileID"],
      message: "请选择通配符证书配置",
    });
  }
  if (values.certificateType === "single" && values.acmeChallengeType === "dns" && !values.dnsProviderID) {
    context.addIssue({ code: "custom", path: ["dnsProviderID"], message: "请选择 DNS Provider" });
  }
});

export type SiteFormValues = z.infer<typeof siteFormSchema>;

export const defaultSiteValues: SiteFormValues = {
  domains: [""],
  upstreams: [""],
  upstreamType: "http",
  upstreamTLSServerName: "",
  upstreamTLSInsecureSkipVerify: false,
  enabled: false,
  enableHTTPS: true,
  forceHTTPS: true,
  certificateType: "single",
  certificateDomain: "",
  acmeChallengeType: "http",
  dnsProvider: "alidns",
  dnsProviderID: "",
  certificateProfileID: "",
  enableWS: true,
  enableGzip: true,
  enableLog: false,
  requestHeaders: "{}",
  responseHeaders: "{}",
  allowedIPs: "",
  basicAuthEnabled: false,
  basicAuthCredentialIDs: [],
  advancedJSON: "",
};

export function formValuesFromSite(site: ProxySite, clone: boolean): SiteFormValues {
  return {
    domains: site.domains.length ? site.domains : [""],
    upstreams: site.upstreams.length ? site.upstreams : [""],
    upstreamType: site.upstream_type || "http",
    upstreamTLSServerName: site.upstream_tls_server_name,
    upstreamTLSInsecureSkipVerify: site.upstream_tls_insecure_skip_verify,
    enabled: clone ? false : site.enabled,
    enableHTTPS: site.enable_https,
    forceHTTPS: site.force_https,
    certificateType: site.certificate_type || "single",
    certificateDomain: site.certificate_domain,
    acmeChallengeType: site.acme_challenge_type || "http",
    dnsProvider: "alidns",
    dnsProviderID: site.dns_provider_id ?? "",
    certificateProfileID: site.certificate_profile_id ?? "",
    enableWS: site.enable_ws,
    enableGzip: site.enable_gzip,
    enableLog: site.enable_log,
    requestHeaders: JSON.stringify(site.request_headers, null, 2),
    responseHeaders: JSON.stringify(site.response_headers, null, 2),
    allowedIPs: site.allowed_ips.join("\n"),
    basicAuthEnabled: site.basic_auth_enabled,
    basicAuthCredentialIDs: site.basic_auth_credential_ids ?? [],
    advancedJSON: site.advanced_json,
  };
}

export function payloadFromForm(values: SiteFormValues, forceDisabled = false): ProxySitePayload {
  return {
    name: "",
    description: "",
    domains: compactItems(values.domains),
    upstreams: compactItems(values.upstreams),
    upstream_type: values.upstreamType,
    upstream_tls_server_name: values.upstreamTLSServerName.trim(),
    upstream_tls_insecure_skip_verify: values.upstreamTLSInsecureSkipVerify,
    enabled: forceDisabled ? false : values.enabled,
    enable_https: values.enableHTTPS,
    force_https: values.forceHTTPS,
    certificate_type: values.certificateType,
    certificate_domain: values.certificateDomain.trim(),
    acme_challenge_type:
      values.certificateType === "wildcard" ? "dns" : values.acmeChallengeType,
    dns_provider: values.dnsProvider,
    dns_provider_id:
      values.certificateType === "single" && values.acmeChallengeType === "dns"
        ? values.dnsProviderID || undefined
        : undefined,
    certificate_profile_id:
      values.certificateType === "wildcard" ? values.certificateProfileID || undefined : undefined,
    enable_ws: values.enableWS,
    enable_gzip: values.enableGzip,
    enable_log: values.enableLog,
    request_headers: JSON.parse(values.requestHeaders) as Record<string, string>,
    response_headers: JSON.parse(values.responseHeaders) as Record<string, string>,
    allowed_ips: splitTextLines(values.allowedIPs),
    basic_auth_enabled: values.basicAuthEnabled,
    basic_auth_users: {},
    basic_auth_credential_ids: values.basicAuthCredentialIDs,
    advanced_json: values.advancedJSON.trim(),
  };
}

export function draftPreview(values: SiteFormValues) {
  const transport = draftTransport(values);
  return {
    match: [{ host: compactItems(values.domains) }],
    handle: [
      {
        handler: "reverse_proxy",
        upstreams: compactItems(values.upstreams).map((dial) => ({ dial })),
        ...(transport ? { transport } : {}),
      },
    ],
    enabled: values.enabled,
    note: "草稿片段；完整配置会在发布前由后端生成并注入管理入口。",
  };
}

function draftTransport(values: SiteFormValues) {
  if (values.upstreamType === "https") {
    return {
      protocol: "http",
      tls: {
        ...(values.upstreamTLSServerName
          ? { server_name: values.upstreamTLSServerName }
          : {}),
        ...(values.upstreamTLSInsecureSkipVerify ? { insecure_skip_verify: true } : {}),
      },
    };
  }
  if (values.upstreamType === "h2c") return { protocol: "http", versions: ["h2c"] };
  return null;
}

function compactItems(value: string[]) {
  return value.map((item) => item.trim()).filter(Boolean);
}

function splitTextLines(value: string) {
  return value.split(/\r?\n/).map((item) => item.trim()).filter(Boolean);
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

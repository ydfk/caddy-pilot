import { z } from "zod";

import type { ProxySite, ProxySitePayload } from "@/api/proxy-sites";
import type { CertificateProfile } from "@/api/certificates";

const jsonObject = z.string().refine(isJSONObject, "请输入有效的 JSON 对象");
const optionalJSON = z
  .string()
  .refine((value) => !value.trim() || isJSON(value), "请输入有效 JSON");

export const siteFormSchema = z
  .object({
    description: z.string().max(2000),
    siteType: z.enum(["proxy", "static", "spa"]),
    configMode: z.enum(["visual", "custom"]),
    customFormat: z.enum(["json", "caddyfile"]),
    customConfig: z.string(),
    domains: z.array(z.string()),
    upstreams: z.array(z.string()),
    rootPath: z.string().max(1024),
    apiPath: z.string().max(256),
    enableSecurityHeaders: z.boolean(),
    enableAssetCache: z.boolean(),
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
  })
  .superRefine((values, context) => {
    if (values.configMode === "custom") {
      if (!values.customConfig.trim()) {
        context.addIssue({ code: "custom", path: ["customConfig"], message: "请输入自定义配置" });
      } else if (values.customFormat === "json" && !isJSONObject(values.customConfig)) {
        context.addIssue({
          code: "custom",
          path: ["customConfig"],
          message: "请输入有效的 Caddy JSON 路由对象",
        });
      }
      return;
    }
    if (compactItems(values.domains).length === 0) {
      context.addIssue({ code: "custom", path: ["domains"], message: "至少填写一个域名" });
    }
    if (values.siteType !== "static" && compactItems(values.upstreams).length === 0) {
      context.addIssue({ code: "custom", path: ["upstreams"], message: "至少填写一个上游" });
    }
    if (values.siteType !== "proxy" && !values.rootPath.trim()) {
      context.addIssue({ code: "custom", path: ["rootPath"], message: "请填写静态文件根目录" });
    }
    if (values.siteType === "spa" && !values.apiPath.trim().startsWith("/")) {
      context.addIssue({ code: "custom", path: ["apiPath"], message: "API 路径必须以 / 开头" });
    }
    if (values.certificateType === "wildcard" && !values.certificateProfileID) {
      context.addIssue({
        code: "custom",
        path: ["certificateProfileID"],
        message: "请选择通配符证书配置",
      });
    }
    if (
      values.certificateType === "single" &&
      values.acmeChallengeType === "dns" &&
      !values.dnsProviderID
    ) {
      context.addIssue({ code: "custom", path: ["dnsProviderID"], message: "请选择 DNS Provider" });
    }
  });

export type SiteFormValues = z.infer<typeof siteFormSchema>;

export const defaultSiteValues: SiteFormValues = {
  description: "",
  siteType: "proxy",
  configMode: "visual",
  customFormat: "caddyfile",
  customConfig: "",
  domains: [""],
  upstreams: [""],
  rootPath: "",
  apiPath: "/api/*",
  enableSecurityHeaders: true,
  enableAssetCache: true,
  upstreamType: "http",
  upstreamTLSServerName: "",
  upstreamTLSInsecureSkipVerify: false,
  enabled: true,
  enableHTTPS: true,
  forceHTTPS: true,
  certificateType: "single",
  certificateDomain: "",
  acmeChallengeType: "http",
  dnsProvider: "alidns",
  dnsProviderID: "",
  certificateProfileID: "",
  enableWS: false,
  enableGzip: true,
  enableLog: true,
  requestHeaders: "{}",
  responseHeaders: "{}",
  allowedIPs: "",
  basicAuthEnabled: false,
  basicAuthCredentialIDs: [],
  advancedJSON: "",
};

export function formValuesFromSite(site: ProxySite, clone: boolean): SiteFormValues {
  return {
    description: site.description,
    siteType: site.site_type || "proxy",
    configMode: site.config_mode || "visual",
    customFormat: site.custom_format === "json" ? "json" : "caddyfile",
    customConfig: site.custom_config || "",
    domains: site.domains.length ? site.domains : [""],
    upstreams: site.upstreams.length ? site.upstreams : [""],
    rootPath: site.root_path,
    apiPath: site.api_path || "/api/*",
    enableSecurityHeaders: site.enable_security_headers,
    enableAssetCache: site.enable_asset_cache,
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
    enableWS: false,
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
    description: values.description.trim(),
    site_type: values.siteType,
    config_mode: values.configMode,
    custom_format: values.configMode === "custom" ? values.customFormat : "",
    custom_config: values.configMode === "custom" ? values.customConfig.trim() : "",
    domains: compactItems(values.domains),
    upstreams: compactItems(values.upstreams),
    root_path: values.rootPath.trim(),
    api_path: values.apiPath.trim() || "/api/*",
    enable_security_headers: values.enableSecurityHeaders,
    enable_asset_cache: values.enableAssetCache,
    upstream_type: values.upstreamType,
    upstream_tls_server_name: values.upstreamTLSServerName.trim(),
    upstream_tls_insecure_skip_verify: values.upstreamTLSInsecureSkipVerify,
    enabled: forceDisabled ? false : values.enabled,
    enable_https: values.enableHTTPS,
    force_https: values.forceHTTPS,
    certificate_type: values.certificateType,
    certificate_domain: values.certificateDomain.trim(),
    acme_challenge_type: values.certificateType === "wildcard" ? "dns" : values.acmeChallengeType,
    dns_provider: values.dnsProvider,
    dns_provider_id:
      values.certificateType === "single" && values.acmeChallengeType === "dns"
        ? values.dnsProviderID || undefined
        : undefined,
    certificate_profile_id:
      values.certificateType === "wildcard" ? values.certificateProfileID || undefined : undefined,
    enable_ws: false,
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

export function matchingWildcardProfile(domains: string[], profiles: CertificateProfile[]) {
  const normalizedDomains = compactItems(domains).map((domain) => domain.toLowerCase());
  if (normalizedDomains.length === 0) return undefined;
  return profiles
    .filter((profile) => profile.enabled && profile.certificate_type === "wildcard")
    .map((profile) => ({
      profile,
      score: wildcardProfileScore(normalizedDomains, profile.subjects),
    }))
    .filter((item) => item.score > 0)
    .sort((left, right) => right.score - left.score)[0]?.profile;
}

function wildcardProfileScore(domains: string[], subjects: string[]) {
  const wildcards = subjects
    .map((subject) => subject.toLowerCase())
    .filter((subject) => subject.startsWith("*."));
  if (
    !domains.every((domain) => wildcards.some((wildcard) => wildcardCoversDomain(wildcard, domain)))
  ) {
    return 0;
  }
  return Math.max(...wildcards.map((wildcard) => wildcard.length));
}

function wildcardCoversDomain(wildcard: string, domain: string) {
  const suffix = wildcard.slice(1);
  if (!domain.endsWith(suffix)) return false;
  const label = domain.slice(0, -suffix.length);
  return label.length > 0 && !label.includes(".");
}

function compactItems(value: string[]) {
  return value.map((item) => item.trim()).filter(Boolean);
}

function splitTextLines(value: string) {
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

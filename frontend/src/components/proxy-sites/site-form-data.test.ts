import { describe, expect, test } from "vitest";

import type { ProxySite } from "@/api/proxy-sites";
import type { CertificateProfile } from "@/api/certificates";
import {
  defaultSiteValues,
  formValuesFromSite,
  matchingWildcardProfile,
  payloadFromForm,
  siteFormSchema,
} from "./site-form-data";

describe("代理站点表单转换", () => {
  test("新增站点默认启用并记录访问日志", () => {
    expect(defaultSiteValues.enabled).toBe(true);
    expect(defaultSiteValues.enableLog).toBe(true);
  });

  test("克隆表单始终默认停用", () => {
    const values = formValuesFromSite(sampleSite(), true);
    expect(values.enabled).toBe(false);
  });

  test("重复项和 JSON 字段转换为 API 载荷", () => {
    const payload = payloadFromForm({
      ...defaultSiteValues,
      domains: ["example.com", "www.example.com", ""],
      upstreams: ["127.0.0.1:3000", "127.0.0.1:3001"],
      requestHeaders: '{"X-Request":"value"}',
      responseHeaders: '{"X-Response":"value"}',
      allowedIPs: "127.0.0.1\n10.0.0.0/8",
      enabled: true,
      description: "内网服务",
    });

    expect(payload.name).toBe("");
    expect(payload.domains).toEqual(["example.com", "www.example.com"]);
    expect(payload.upstreams).toHaveLength(2);
    expect(payload.request_headers).toEqual({ "X-Request": "value" });
    expect(payload.allowed_ips).toEqual(["127.0.0.1", "10.0.0.0/8"]);
    expect(payload.dns_provider).toBe("alidns");
    expect(payload.description).toBe("内网服务");
  });

  test("域名优先匹配最具体的通配符证书", () => {
    const broad = sampleCertificate("broad", ["*.example.com"]);
    const specific = sampleCertificate("specific", ["*.internal.example.com"]);
    expect(matchingWildcardProfile(["app.internal.example.com"], [broad, specific])?.id).toBe(
      "specific"
    );
    expect(matchingWildcardProfile(["deep.app.example.com"], [broad])).toBeUndefined();
  });

  test("静态站点不要求上游，SPA 要求目录和 API 上游", () => {
    const staticResult = siteFormSchema.safeParse({
      ...defaultSiteValues,
      siteType: "static",
      domains: ["static.example.com"],
      upstreams: [],
      rootPath: "/var/www/example",
    });
    expect(staticResult.success).toBe(true);

    const spaResult = siteFormSchema.safeParse({
      ...defaultSiteValues,
      siteType: "spa",
      domains: ["app.example.com"],
      rootPath: "/var/www/app/dist",
      apiPath: "/api/*",
      upstreams: ["127.0.0.1:3000"],
    });
    expect(spaResult.success).toBe(true);
  });
});

function sampleCertificate(id: string, subjects: string[]): CertificateProfile {
  return {
    id,
    name: id,
    certificate_type: "wildcard",
    subjects,
    challenge_type: "dns",
    enabled: true,
    created_at: "2026-06-20T00:00:00Z",
    updated_at: "2026-06-20T00:00:00Z",
    issued_certificates: [],
    issuance_state: "unused",
    issuance_message: "",
    usage_count: 0,
  };
}

function sampleSite(): ProxySite {
  return {
    id: "site-id",
    name: "示例站点",
    description: "",
    site_type: "proxy",
    domains: ["example.com"],
    upstreams: ["127.0.0.1:3000"],
    root_path: "",
    api_path: "/api/*",
    enable_security_headers: false,
    enable_asset_cache: false,
    upstream_type: "http",
    upstream_tls_server_name: "",
    upstream_tls_insecure_skip_verify: false,
    enable_https: true,
    force_https: true,
    certificate_type: "single",
    certificate_domain: "",
    acme_challenge_type: "http",
    dns_provider: "",
    enable_gzip: true,
    enable_log: false,
    enable_ws: true,
    request_headers: {},
    response_headers: {},
    basic_auth_enabled: false,
    basic_auth_users: {},
    basic_auth_credential_ids: [],
    allowed_ips: [],
    advanced_json: "",
    enabled: true,
    created_at: "2026-06-20T00:00:00Z",
    updated_at: "2026-06-20T00:00:00Z",
  };
}

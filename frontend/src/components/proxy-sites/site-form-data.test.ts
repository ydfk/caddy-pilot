import { describe, expect, test } from "vitest";

import type { ProxySite } from "@/api/proxy-sites";
import { defaultSiteValues, formValuesFromSite, payloadFromForm } from "./site-form-data";

describe("代理站点表单转换", () => {
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
    });

    expect(payload.name).toBe("");
    expect(payload.domains).toEqual(["example.com", "www.example.com"]);
    expect(payload.upstreams).toHaveLength(2);
    expect(payload.request_headers).toEqual({ "X-Request": "value" });
    expect(payload.allowed_ips).toEqual(["127.0.0.1", "10.0.0.0/8"]);
  });
});

function sampleSite(): ProxySite {
  return {
    id: "site-id",
    name: "示例站点",
    description: "",
    domains: ["example.com"],
    upstreams: ["127.0.0.1:3000"],
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

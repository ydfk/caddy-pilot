import { apiRequest } from "./client";

export type ProxySitePayload = {
  name: string;
  description: string;
  domains: string[];
  upstreams: string[];
  upstream_type: "http" | "https" | "h2c" | "unix";
  upstream_tls_server_name: string;
  upstream_tls_insecure_skip_verify: boolean;
  enable_https: boolean;
  force_https: boolean;
  https_redirect_port: number;
  certificate_type: "single" | "wildcard";
  certificate_domain: string;
  acme_challenge_type: "http" | "dns";
  dns_provider: "" | "alidns";
  dns_provider_id?: string;
  certificate_profile_id?: string;
  enable_gzip: boolean;
  enable_log: boolean;
  enable_ws: boolean;
  request_headers: Record<string, string>;
  response_headers: Record<string, string>;
  basic_auth_enabled: boolean;
  basic_auth_users: Record<string, string>;
  basic_auth_credential_ids: string[];
  allowed_ips: string[];
  advanced_json: string;
  enabled: boolean;
};

export type ProxySite = ProxySitePayload & { id: string; created_at: string; updated_at: string };

export type ProxySitePreview = { caddy_json: unknown; caddyfile: string };
export type NginxImportResult = { sites: ProxySite[]; warnings: string[] };

export const listProxySites = () => apiRequest<ProxySite[]>("/api/proxy-sites");
export const getProxySite = (id: string) => apiRequest<ProxySite>(`/api/proxy-sites/${id}`);
export const createProxySite = (payload: ProxySitePayload) =>
  apiRequest<ProxySite>("/api/proxy-sites", { method: "POST", body: JSON.stringify(payload) });
export const updateProxySite = (id: string, payload: ProxySitePayload) =>
  apiRequest<ProxySite>(`/api/proxy-sites/${id}`, { method: "PUT", body: JSON.stringify(payload) });
export const deleteProxySite = (id: string) =>
  apiRequest<void>(`/api/proxy-sites/${id}`, { method: "DELETE" });
export const cloneProxySite = (
  id: string,
  payload: Partial<Pick<ProxySitePayload, "name" | "domains" | "upstreams">> = {}
) =>
  apiRequest<ProxySite>(`/api/proxy-sites/${id}/clone`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
export const setProxySiteEnabled = (id: string, enabled: boolean) =>
  apiRequest<ProxySite>(`/api/proxy-sites/${id}/${enabled ? "enable" : "disable"}`, {
    method: "POST",
  });
export const previewProxySite = (id: string) =>
  apiRequest<ProxySitePreview>(`/api/proxy-sites/${id}/preview`, { method: "POST" });
export const previewProxySiteDraft = (payload: ProxySitePayload) =>
  apiRequest<ProxySitePreview>("/api/proxy-sites/preview", {
    method: "POST",
    body: JSON.stringify(payload),
  });
export const importNginxConfig = (config: string) =>
  apiRequest<NginxImportResult>("/api/proxy-sites/import/nginx", {
    method: "POST",
    body: JSON.stringify({ config }),
  });

import { apiRequest } from "./client";

export type CaddyStatus = { online: boolean; admin_api: string; error_message?: string };
export type CaddyVersion = {
  current_version: string;
  latest_version?: string;
  update_available: boolean;
  release_url?: string;
  update_command?: string;
  update_strategy: "rebuild-container";
  error_message?: string;
};
export type CaddyJSONResponse = { caddy_json: unknown };

export const getCaddyStatus = () => apiRequest<CaddyStatus>("/api/caddy/status");
export const getCaddyVersion = () => apiRequest<CaddyVersion>("/api/caddy/version");
export const previewCaddyConfig = () =>
  apiRequest<CaddyJSONResponse>("/api/caddy/preview", { method: "POST" });
export const validateCaddyConfig = () =>
  apiRequest<{ valid: boolean }>("/api/caddy/validate", { method: "POST" });
export const publishCaddyConfig = (reason = "手动发布") =>
  apiRequest<{ id: number; version: number; status: string; reason: string }>(
    "/api/caddy/publish",
    { method: "POST", body: JSON.stringify({ reason }) }
  );
export const getCurrentCaddyConfig = () =>
  apiRequest<CaddyJSONResponse>("/api/caddy/current-config");

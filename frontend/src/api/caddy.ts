import { apiRequest } from "./client";

export type CaddyStatus = { online: boolean; admin_api: string; error_message?: string };
export type CaddyJSONResponse = { caddy_json: unknown };

export const getCaddyStatus = () => apiRequest<CaddyStatus>("/api/caddy/status");
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

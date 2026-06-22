import { apiRequest } from "./client";

export type CaddyStatus = { online: boolean; error_message?: string };
export type CaddyVersion = {
  current_version: string;
  latest_version?: string;
  update_available: boolean;
  binary_path?: string;
  version_check_url: string;
  download_url: string;
  update_url?: string;
  release_url?: string;
  update_strategy: "managed";
  error_message?: string;
};
export type CaddySettings = {
  version_check_url: string;
  download_url: string;
  checksum_url: string;
};
export type CaddyJSONResponse = { caddy_json: unknown };
export type CaddyChangeStatus = { dirty: boolean; latest_version?: number };
export type CaddyUpdateTask = {
  id?: string;
  kind?: "download" | "upload";
  status:
    | "idle"
    | "queued"
    | "downloading"
    | "verifying"
    | "installing"
    | "restarting"
    | "succeeded"
    | "failed";
  target_version?: string;
  progress: number;
  downloaded_bytes?: number;
  total_bytes?: number;
  error_message?: string;
};

export const getCaddyStatus = () => apiRequest<CaddyStatus>("/api/caddy/status");
export const getCaddyVersion = () => apiRequest<CaddyVersion>("/api/caddy/version");
export const getCaddySettings = () => apiRequest<CaddySettings>("/api/caddy/settings");
export const saveCaddySettings = (settings: CaddySettings) =>
  apiRequest<CaddySettings>("/api/caddy/settings", {
    method: "PUT",
    body: JSON.stringify(settings),
  });
export const updateManagedCaddy = (version?: string) =>
  apiRequest<{ accepted: boolean; task_id: string; status: string; target_version: string }>(
    "/api/caddy/update",
    {
      method: "POST",
      body: JSON.stringify({ version: version ?? "" }),
    }
  );
export const getCaddyUpdateTask = () =>
  apiRequest<CaddyUpdateTask>("/api/caddy/update-tasks/current");
export const uploadManagedCaddy = (file: File) => {
  const body = new FormData();
  body.set("file", file);
  return apiRequest<{ accepted: boolean; task_id: string; status: string }>("/api/caddy/upload", {
    method: "POST",
    body,
  });
};
export const getCaddyChangeStatus = () => apiRequest<CaddyChangeStatus>("/api/caddy/change-status");
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

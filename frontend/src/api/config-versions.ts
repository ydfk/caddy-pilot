import { apiRequest } from "./client";

export type ConfigVersionSummary = {
  id: number;
  version: number;
  reason: string;
  status: "draft" | "published" | "failed" | "rollback";
  published_at: string | null;
  created_at: string;
};
export type ConfigVersion = ConfigVersionSummary & {
  business_config: unknown;
  caddy_json: unknown;
  caddyfile?: string;
  error_message: string;
};

export const listConfigVersions = () => apiRequest<ConfigVersionSummary[]>("/api/config-versions");
export const getConfigVersion = (id: string | number) =>
  apiRequest<ConfigVersion>(`/api/config-versions/${id}`);
export const rollbackConfigVersion = (id: string | number) =>
  apiRequest<ConfigVersion>(`/api/config-versions/${id}/rollback`, { method: "POST" });

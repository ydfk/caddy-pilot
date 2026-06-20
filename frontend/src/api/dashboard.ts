import { apiRequest } from "./client";

export type DashboardSummary = {
  site_count: number;
  enabled_site_count: number;
  disabled_site_count: number;
  https_site_count: number;
  last_publish_time: string | null;
  caddy_online: boolean;
  caddy_admin_api: string;
};

export const getDashboardSummary = () => apiRequest<DashboardSummary>("/api/dashboard/summary");

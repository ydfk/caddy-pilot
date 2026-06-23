import { apiRequest } from "./client";

export type DashboardSummary = {
  site_count: number;
  enabled_site_count: number;
  disabled_site_count: number;
  https_site_count: number;
  last_publish_time: string | null;
  caddy_online: boolean;
  request_count_24h: number;
  error_count_24h: number;
  traffic_bytes_24h: number;
  top_sites_24h: TopSite[];
};

export type TopSite = {
  domain: string;
  request_count: number;
  error_count: number;
  bytes: number;
};

export const getDashboardSummary = () => apiRequest<DashboardSummary>("/api/dashboard/summary");

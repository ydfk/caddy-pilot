import { apiRequest } from "./client";

export type LogSource = "system" | "caddy" | "dns";
export type LogEntry = {
  id: string;
  timestamp?: string;
  level?: string;
  message: string;
  fields?: Record<string, unknown>;
};
export type LogResponse = { entries: LogEntry[]; next_cursor: number };

export const listLogs = (source: LogSource, cursor = 0, limit = 200, providerID = "") => {
  const provider = providerID ? `&provider_id=${encodeURIComponent(providerID)}` : "";
  return apiRequest<LogResponse>(
    `/api/logs?source=${source}&cursor=${cursor}&limit=${limit}${provider}`
  );
};

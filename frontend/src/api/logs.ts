import { apiRequest } from "./client";

export type LogSource = "system" | "caddy";
export type LogEntry = {
  id: string;
  timestamp?: string;
  level?: string;
  message: string;
};
export type LogResponse = { entries: LogEntry[]; next_cursor: number };

export const listLogs = (source: LogSource, cursor = 0, limit = 200) =>
  apiRequest<LogResponse>(`/api/logs?source=${source}&cursor=${cursor}&limit=${limit}`);

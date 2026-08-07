import { apiRequest } from "./client";

export type SystemInfo = { version: string; http_port: number; https_port: number };

export const getSystemInfo = () => apiRequest<SystemInfo>("/api/system/info");

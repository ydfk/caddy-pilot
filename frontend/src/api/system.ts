import { apiRequest } from "./client";

export type SystemInfo = { version: string };

export const getSystemInfo = () => apiRequest<SystemInfo>("/api/system/info");

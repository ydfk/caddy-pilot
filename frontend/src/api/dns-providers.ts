import { apiRequest } from "./client";

export type DNSProvider = {
  id: string;
  name: string;
  provider_type: "alidns";
  access_key_id_hint: string;
  region_id: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
};

export type DNSProviderPayload = {
  name: string;
  provider_type: "alidns";
  access_key_id: string;
  access_key_secret: string;
  region_id: string;
  enabled: boolean;
};

export const listDNSProviders = () => apiRequest<DNSProvider[]>("/api/dns-providers");
export const createDNSProvider = (payload: DNSProviderPayload) =>
  apiRequest<DNSProvider>("/api/dns-providers", { method: "POST", body: JSON.stringify(payload) });
export const updateDNSProvider = (id: string, payload: DNSProviderPayload) =>
  apiRequest<DNSProvider>(`/api/dns-providers/${id}`, { method: "PUT", body: JSON.stringify(payload) });
export const deleteDNSProvider = (id: string) =>
  apiRequest<void>(`/api/dns-providers/${id}`, { method: "DELETE" });

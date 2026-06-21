import { apiRequest } from "./client";

export type BasicAuthCredential = {
  id: string;
  name: string;
  username: string;
  created_at: string;
  updated_at: string;
};

export type BasicAuthCredentialPayload = {
  name: string;
  username: string;
  password: string;
};

export const listBasicAuthCredentials = () =>
  apiRequest<BasicAuthCredential[]>("/api/basic-auth-credentials");
export const createBasicAuthCredential = (payload: BasicAuthCredentialPayload) =>
  apiRequest<BasicAuthCredential>("/api/basic-auth-credentials", {
    method: "POST",
    body: JSON.stringify(payload),
  });
export const updateBasicAuthCredential = (id: string, payload: BasicAuthCredentialPayload) =>
  apiRequest<BasicAuthCredential>(`/api/basic-auth-credentials/${id}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
export const deleteBasicAuthCredential = (id: string) =>
  apiRequest<void>(`/api/basic-auth-credentials/${id}`, { method: "DELETE" });

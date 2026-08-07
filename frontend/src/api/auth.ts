import { apiRequest } from "./client";

export type Credentials = { username: string; password: string };
export type User = { id: string; username: string; createdAt: string; updatedAt: string };
export type SetupStatus = { initialized: boolean };
export type PasskeyStatus = { available: boolean; configured: boolean; error_message?: string };
export type Passkey = {
  id: string;
  name: string;
  created_at: string;
  last_used_at?: string;
};

export type PasskeyCreationOptionsJSON = Omit<
  PublicKeyCredentialCreationOptions,
  "challenge" | "user" | "excludeCredentials"
> & {
  challenge: string;
  user: Omit<PublicKeyCredentialUserEntity, "id"> & { id: string };
  excludeCredentials?: Array<Omit<PublicKeyCredentialDescriptor, "id"> & { id: string }>;
};

export type PasskeyRequestOptionsJSON = Omit<
  PublicKeyCredentialRequestOptions,
  "challenge" | "allowCredentials"
> & {
  challenge: string;
  allowCredentials?: Array<Omit<PublicKeyCredentialDescriptor, "id"> & { id: string }>;
};

export const getSetupStatus = () => apiRequest<SetupStatus>("/api/auth/setup-status");
export const getPasskeyStatus = () => apiRequest<PasskeyStatus>("/api/auth/passkeys/status");

export const login = (credentials: Credentials) =>
  apiRequest<{ token: string }>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(credentials),
  });
export const register = (credentials: Credentials) =>
  apiRequest<User>("/api/auth/register", { method: "POST", body: JSON.stringify(credentials) });
export const getProfile = () => apiRequest<User>("/api/auth/profile");

export const beginPasskeyLogin = () =>
  apiRequest<{
    session_id: string;
    options: { publicKey: PasskeyRequestOptionsJSON; mediation?: CredentialMediationRequirement };
  }>("/api/auth/passkeys/login/options", { method: "POST" });

export const finishPasskeyLogin = (sessionID: string, credential: unknown) =>
  apiRequest<{ token: string }>("/api/auth/passkeys/login/verify", {
    method: "POST",
    body: JSON.stringify({ session_id: sessionID, credential }),
  });

export const listPasskeys = () => apiRequest<{ items: Passkey[] }>("/api/auth/passkeys");

export const beginPasskeyRegistration = (name: string) =>
  apiRequest<{
    session_id: string;
    options: { publicKey: PasskeyCreationOptionsJSON; mediation?: CredentialMediationRequirement };
  }>("/api/auth/passkeys/register/options", {
    method: "POST",
    body: JSON.stringify({ name }),
  });

export const finishPasskeyRegistration = (sessionID: string, credential: unknown) =>
  apiRequest<Passkey>("/api/auth/passkeys/register/verify", {
    method: "POST",
    body: JSON.stringify({ session_id: sessionID, credential }),
  });

export const renamePasskey = (id: string, name: string) =>
  apiRequest<Passkey>(`/api/auth/passkeys/${id}`, {
    method: "PATCH",
    body: JSON.stringify({ name }),
  });

export const deletePasskey = (id: string) =>
  apiRequest<void>(`/api/auth/passkeys/${id}`, { method: "DELETE" });

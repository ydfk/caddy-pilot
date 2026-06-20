import { apiRequest } from "./client";

export type Credentials = { username: string; password: string };
export type User = { id: string; username: string; createdAt: string; updatedAt: string };

export const login = (credentials: Credentials) =>
  apiRequest<{ token: string }>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify(credentials),
  });
export const register = (credentials: Credentials) =>
  apiRequest<User>("/api/auth/register", { method: "POST", body: JSON.stringify(credentials) });
export const getProfile = () => apiRequest<User>("/api/auth/profile");

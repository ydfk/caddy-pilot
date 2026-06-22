import { useAuthStore } from "@/store/auth-store";

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, "");

export class APIError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = "APIError";
  }
}

export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = useAuthStore.getState().token;
  const headers = new Headers(options.headers);
  headers.set("Accept", "application/json");
  if (options.body && !(options.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (token) headers.set("Authorization", `Bearer ${token}`);

  const response = await fetch(`${API_BASE_URL}${path}`, { ...options, headers });
  if (response.status === 401) {
    useAuthStore.getState().clear();
    if (window.location.pathname !== "/login") window.location.assign("/login");
  }
  if (!response.ok) throw new APIError(response.status, await errorMessage(response));
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

async function errorMessage(response: Response) {
  const fallback = `请求失败（${response.status}）`;
  const contentType = response.headers.get("content-type") ?? "";
  if (!contentType.includes("json")) {
    const message = (await response.text()).trim();
    return message || fallback;
  }
  try {
    const body = (await response.json()) as {
      detail?: string;
      title?: string;
      errors?: Array<{ message?: string; location?: string }>;
    };
    const validation = body.errors
      ?.map((item) => [item.location, item.message].filter(Boolean).join(": "))
      .filter(Boolean)
      .join("；");
    return body.detail || validation || body.title || fallback;
  } catch {
    return fallback;
  }
}

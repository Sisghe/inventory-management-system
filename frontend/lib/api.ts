const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

// ===== Types =====
export type MeResponse = { user_id: number; username: string };
export type LoginResponse = { access_token: string };

export type UserDTO = {
  id?: number;
  username: string;
  // nel FE la password la userai solo per create/update
  password?: string;
  nome?: string;
  cognome?: string;
  data_nascita?: string; // ISO date string (es. "1990-01-01")
};

type RequestOptions = {
  method?: string;
  body?: unknown;
  auth?: boolean;
};

// ===== Token helpers =====
export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("access_token");
}

export function setToken(token: string) {
  // LocalStorage (per fetch client-side)
  localStorage.setItem("access_token", token);

  // Cookie (per middleware Next.js che NON vede localStorage)
  // Nota: non possiamo impostare HttpOnly da JS (si può solo lato server).
  document.cookie = `access_token=${encodeURIComponent(token)}; path=/; samesite=lax`;
}

export function clearToken() {
  localStorage.removeItem("access_token");
  document.cookie = "access_token=; Max-Age=0; path=/; samesite=lax";
}

// ===== Internal helpers =====
function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function getErrorMessage(data: unknown, fallback: string) {
  if (isObject(data)) {
    const err = data["error"];
    const msg = data["message"];
    if (typeof err === "string") return err;
    if (typeof msg === "string") return msg;
  }
  return fallback;
}

// ===== HTTP client =====
async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { method = "GET", body, auth = true } = opts;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (auth) {
    const token = getToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  const isJson = res.headers.get("content-type")?.includes("application/json");
  const data: unknown = isJson ? await res.json() : null;

  if (!res.ok) {
    throw new Error(getErrorMessage(data, `HTTP ${res.status}`));
  }

  return data as T;
}

// ===== API methods =====
export const api = {
  login: (username: string, password: string) =>
    request<LoginResponse>("/auth/login", {
      method: "POST",
      body: { username, password },
      auth: false,
    }),

  me: () => request<MeResponse>("/api/me"),

  users: {
    list: () => request<UserDTO[]>("/api/users"),
    create: (payload: UserDTO) => request<UserDTO>("/api/users", { method: "POST", body: payload }),
    update: (id: number, payload: UserDTO) =>
      request<UserDTO>(`/api/users/${id}`, { method: "PUT", body: payload }),
    delete: (id: number) => request<void>(`/api/users/${id}`, { method: "DELETE" }),
  },
};

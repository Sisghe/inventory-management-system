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

export type ProductTypeDTO = {
  id: number;
  tipo: string;
};

export type ProductDTO = {
  id?: number;
  nome_oggetto: string;
  descrizione?: string | null;
  data_inserimento?: string; // ISO string dal backend
  tipo_prodotto_id?: number | null;
};

type RequestOptions = {
  method?: string;
  body?: unknown;
  auth?: boolean; // lasciato per compatibilità, ma con cookie HttpOnly non serve più
};

// ===== Token helpers (DEPRECATED) =====
export function getToken(): string | null {
  return null;
}
export function setToken() {
  // no-op
}
export function clearToken() {
  // no-op
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
  const { method = "GET", body } = opts;

  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    credentials: "include",
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const contentType = res.headers.get("content-type") || "";
  const isJson = contentType.includes("application/json");

  let data: unknown = null;
  if (isJson) data = await res.json().catch(() => null);
  else data = await res.text().catch(() => null);

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

  logout: () =>
    request<void>("/auth/logout", {
      method: "POST",
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

  productTypes: {
    list: () => request<ProductTypeDTO[]>("/api/product-types"),
  },

  products: {
    list: () => request<ProductDTO[]>("/api/products"),
    create: (payload: ProductDTO) => request<ProductDTO>("/api/products", { method: "POST", body: payload }),
    update: (id: number, payload: ProductDTO) =>
      request<ProductDTO>(`/api/products/${id}`, { method: "PUT", body: payload }),
    delete: (id: number) => request<void>(`/api/products/${id}`, { method: "DELETE" }),
  },
};

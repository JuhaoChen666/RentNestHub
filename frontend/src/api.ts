import type {
  House,
  HouseUpdateInput,
  HouseFilters,
  InquiryMessage,
  ListingResult,
  RecommendationResult,
  AuthSession,
  HouseReview,
  User,
} from "./types";

const authTokenKey = "rentnesthub.auth-token";

export function saveAuthToken(token: string) {
  sessionStorage.setItem(authTokenKey, token);
}

export function clearAuthToken() {
  sessionStorage.removeItem(authTokenKey);
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  const token = sessionStorage.getItem(authTokenKey);
  if (token) headers.set("Authorization", `Bearer ${token}`);
  const response = await fetch(url, { ...init, headers });
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as {
      error?: string;
    } | null;
    throw new Error(body?.error ?? `请求失败 (${response.status})`);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return response.json() as Promise<T>;
}

export async function login(input: {
  identifier: string;
  password: string;
}): Promise<AuthSession> {
  return request<AuthSession>("/api/v1/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export async function register(input: {
  username: string;
  displayName: string;
  email: string;
  password: string;
}): Promise<AuthSession> {
  return request<AuthSession>("/api/v1/auth/register", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function currentUser(): Promise<User> {
  return request<User>("/api/v1/auth/me");
}

export function updateProfile(email: string): Promise<User> {
  return request<User>("/api/v1/auth/me", {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
}

export function requestPasswordReset(email: string): Promise<void> {
  return request("/api/v1/auth/password-reset/request", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
}

export function confirmPasswordReset(input: {
  email: string;
  code: string;
  newPassword: string;
}): Promise<void> {
  return request("/api/v1/auth/password-reset/confirm", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export async function listHouses(
  filters: HouseFilters,
  offset = 0,
): Promise<ListingResult> {
  const params = new URLSearchParams();
  Object.entries(filters).forEach(([key, value]) => {
    if (value) params.set(key, value);
  });
  params.set("limit", "24");
  params.set("offset", String(offset));
  return request<ListingResult>(`/api/v1/houses?${params}`);
}

export async function recommend(input: {
  need: string;
  city: string;
  district: string;
  maxRent: number;
  bedrooms: number;
}): Promise<RecommendationResult> {
  return request<RecommendationResult>(
    "/api/v1/recommendations",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tenantId: 2, limit: 6, ...input }),
    },
  );
}

export function favoriteHouse(houseId: number): Promise<void> {
  return request("/api/v1/favorites", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ tenantId: 2, houseId }),
  });
}

export function unfavoriteHouse(houseId: number): Promise<void> {
  return request(`/api/v1/favorites/2/${houseId}`, {
    method: "DELETE",
  });
}

export async function listFavoriteHouses(): Promise<House[]> {
  const data = await request<{ items: House[] }>("/api/v1/favorites/2");
  return data.items;
}

export function sendMessage(houseId: number, content: string, recipientId?: number): Promise<void> {
  return request("/api/v1/messages", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ houseId, content, recipientId }),
  });
}

export async function listInquiryMessages(): Promise<InquiryMessage[]> {
  const data = await request<{ items: InquiryMessage[] }>("/api/v1/messages");
  return data.items;
}

export function publishHouse(form: FormData): Promise<House> {
  return request("/api/v1/houses", {
    method: "POST",
    body: form,
  });
}

export async function listOwnedHouses(): Promise<House[]> {
  const data = await request<{ items: House[] }>("/api/v1/houses/mine");
  return data.items;
}

export function updateOwnedHouse(
  houseId: number,
  input: HouseUpdateInput,
): Promise<House> {
  return request<House>(`/api/v1/houses/${houseId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  });
}

export function deleteOwnedHouse(houseId: number): Promise<void> {
  return request(`/api/v1/houses/${houseId}`, { method: "DELETE" });
}

export async function listPendingHouseReviews(): Promise<HouseReview[]> {
  const data = await request<{ items: HouseReview[] }>("/api/v1/admin/houses/pending");
  return data.items;
}

export function reviewHouse(houseId: number, approved: boolean): Promise<void> {
  return request(`/api/v1/admin/houses/${houseId}/review`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ approved }),
  });
}

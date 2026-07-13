import type {
  House,
  HouseFilters,
  ListingResult,
  RecommendationResult,
} from "./types";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init);
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

export function sendMessage(houseId: number, content: string): Promise<void> {
  return request("/api/v1/messages", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ senderId: 2, houseId, content }),
  });
}

export function publishHouse(form: FormData): Promise<House> {
  return request("/api/v1/houses", {
    method: "POST",
    body: form,
  });
}

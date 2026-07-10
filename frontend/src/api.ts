import type { House, HouseFilters, Recommendation } from "./types";

interface ListResponse {
  items: House[];
}

interface RecommendationResponse {
  items: Recommendation[];
  mode: string;
}

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

export async function listHouses(filters: HouseFilters): Promise<House[]> {
  const params = new URLSearchParams();
  Object.entries(filters).forEach(([key, value]) => {
    if (value) params.set(key, value);
  });
  const data = await request<ListResponse>(`/api/v1/houses?${params}`);
  return data.items;
}

export async function recommend(input: {
  need: string;
  city: string;
  district: string;
  maxRent: number;
  bedrooms: number;
}): Promise<Recommendation[]> {
  const data = await request<RecommendationResponse>(
    "/api/v1/recommendations",
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ tenantId: 2, limit: 6, ...input }),
    },
  );
  return data.items;
}

export function favoriteHouse(houseId: number): Promise<void> {
  return request("/api/v1/favorites", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ tenantId: 2, houseId }),
  });
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

export interface House {
  id: number;
  landlordId: number;
  title: string;
  description: string;
  city: string;
  district: string;
  address: string;
  monthlyRent: number;
  bedrooms: number;
  bathrooms: number;
  areaSqm: number;
  amenities: string[];
  imageUrls: string[];
  status: string;
  createdAt: string;
}

export type HouseDetailsUpdate = Pick<
  House,
  | "title"
  | "description"
  | "city"
  | "district"
  | "address"
  | "monthlyRent"
  | "bedrooms"
  | "bathrooms"
  | "areaSqm"
  | "amenities"
>;

export type HouseUpdateInput = HouseDetailsUpdate | { status: "rented" };

export interface HouseReview {
  house: House;
  publisher: User;
}

export interface HouseFilters {
  city: string;
  district: string;
  keyword: string;
  maxRent: string;
  bedrooms: string;
}

export interface ListingMeta {
  limit: number;
  offset: number;
  count: number;
  hasMore: boolean;
  sort: string;
}

export interface ListingResult {
  items: House[];
  meta: ListingMeta;
}

export interface Recommendation {
  house: House;
  score: number;
  reason: string;
}

export interface RecommendationResult {
	items: Recommendation[];
	mode: string;
}

export interface InquiryMessage {
	id: number;
	houseId: number;
	houseTitle: string;
	sender: User;
	recipient: User;
	content: string;
	createdAt: string;
}

export interface User {
	id: number;
	role: "admin" | "landlord" | "tenant";
	username: string;
	displayName: string;
	email: string;
	createdAt: string;
}

export interface AuthSession {
	token: string;
	user: User;
}

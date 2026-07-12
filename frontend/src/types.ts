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

export interface HouseFilters {
  city: string;
  district: string;
  keyword: string;
  maxRent: string;
  bedrooms: string;
}

export interface Recommendation {
  house: House;
  score: number;
  reason: string;
}

package domain

import (
	"context"
	"time"
)

type House struct {
	ID          int64     `json:"id"`
	LandlordID  int64     `json:"landlordId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	City        string    `json:"city"`
	District    string    `json:"district"`
	Address     string    `json:"address"`
	MonthlyRent int       `json:"monthlyRent"`
	Bedrooms    int       `json:"bedrooms"`
	Bathrooms   int       `json:"bathrooms"`
	AreaSqm     float64   `json:"areaSqm"`
	Amenities   []string  `json:"amenities"`
	ImageURLs   []string  `json:"imageUrls"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

type HouseFilter struct {
	City       string
	District   string
	Keyword    string
	MinRent    int
	MaxRent    int
	Bedrooms   int
	Limit      int
	Offset     int
	Sort       string
	OnlyActive bool
}

type RecommendationRequest struct {
	TenantID int64  `json:"tenantId"`
	Need     string `json:"need"`
	City     string `json:"city"`
	District string `json:"district"`
	MaxRent  int    `json:"maxRent"`
	Bedrooms int    `json:"bedrooms"`
	Limit    int    `json:"limit"`
}

type Recommendation struct {
	House  House   `json:"house"`
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type Favorite struct {
	TenantID int64 `json:"tenantId"`
	HouseID  int64 `json:"houseId"`
}

type Message struct {
	ID         int64     `json:"id"`
	HouseID    int64     `json:"houseId"`
	HouseTitle string    `json:"houseTitle"`
	SenderID   int64     `json:"senderId"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
}

type HouseRepository interface {
	ListHouses(context.Context, HouseFilter) ([]House, error)
	GetHouse(context.Context, int64) (House, error)
	CreateHouse(context.Context, *House) error
	AddFavorite(context.Context, Favorite) error
	RemoveFavorite(context.Context, Favorite) error
	ListFavoriteHouses(context.Context, int64) ([]House, error)
	CreateMessage(context.Context, *Message) error
	ListMessages(context.Context, int64) ([]Message, error)
	Close() error
}

type Recommender interface {
	Recommend(context.Context, RecommendationRequest) ([]Recommendation, error)
	Mode() string
}

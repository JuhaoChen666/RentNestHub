package service

import (
	"context"
	"testing"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

type fakeRepository struct {
	houses []domain.House
}

type fakeProvider struct {
	called bool
	limit  int
	houses []domain.House
}

func (provider *fakeProvider) Recommend(
	_ context.Context,
	houses []domain.House,
	_ domain.RecommendationRequest,
	limit int,
) ([]domain.Recommendation, error) {
	provider.called = true
	provider.limit = limit
	provider.houses = houses
	return []domain.Recommendation{{House: houses[0], Score: 88, Reason: "provider result"}}, nil
}

func (*fakeProvider) Mode() string {
	return "fake-provider"
}

func (repository fakeRepository) ListHouses(context.Context, domain.HouseFilter) ([]domain.House, error) {
	return repository.houses, nil
}

func (fakeRepository) GetHouse(context.Context, int64) (domain.House, error) {
	return domain.House{}, nil
}

func (fakeRepository) CreateHouse(context.Context, *domain.House) error {
	return nil
}

func (fakeRepository) ListOwnedHouses(context.Context, int64) ([]domain.House, error) {
	return nil, nil
}

func (fakeRepository) UpdateHouseRent(context.Context, int64, int) error {
	return nil
}

func (fakeRepository) UpdateHouseStatus(context.Context, int64, string) error {
	return nil
}

func (fakeRepository) DeleteHouse(context.Context, int64) error {
	return nil
}

func (fakeRepository) ListPendingHouseReviews(context.Context) ([]domain.HouseReview, error) {
	return nil, nil
}

func (fakeRepository) ReviewHouse(context.Context, int64, bool) error {
	return nil
}

func (fakeRepository) AddFavorite(context.Context, domain.Favorite) error {
	return nil
}

func (fakeRepository) RemoveFavorite(context.Context, domain.Favorite) error {
	return nil
}

func (fakeRepository) ListFavoriteHouses(context.Context, int64) ([]domain.House, error) {
	return nil, nil
}

func (fakeRepository) CreateMessage(context.Context, *domain.Message) error {
	return nil
}

func (fakeRepository) ListMessages(context.Context, int64) ([]domain.Message, error) {
	return nil, nil
}

func (fakeRepository) ConversationExists(context.Context, int64, int64, int64) (bool, error) {
	return false, nil
}

func (fakeRepository) Close() error {
	return nil
}

func TestRecommendPrioritizesMatchingBudgetAndDistrict(t *testing.T) {
	repository := fakeRepository{houses: []domain.House{
		{ID: 1, District: "浦东新区", MonthlyRent: 7200, Bedrooms: 2},
		{ID: 2, District: "徐汇区", MonthlyRent: 5200, Bedrooms: 2},
	}}
	recommender := NewRecommender(repository)

	result, err := recommender.Recommend(context.Background(), domain.RecommendationRequest{
		District: "徐汇区",
		MaxRent:  6000,
		Bedrooms: 2,
	})
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if len(result) != 2 || result[0].House.ID != 2 {
		t.Fatalf("expected matching house first, got %#v", result)
	}
}

func TestRecommendUsesInjectedProvider(t *testing.T) {
	repository := fakeRepository{houses: []domain.House{{ID: 7}}}
	provider := &fakeProvider{}
	recommender := NewRecommenderWithProvider(repository, provider)

	result, err := recommender.Recommend(context.Background(), domain.RecommendationRequest{
		Limit: 3,
	})
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if !provider.called || provider.limit != 3 || len(provider.houses) != 1 {
		t.Fatalf("provider was not called correctly: %#v", provider)
	}
	if len(result) != 1 || result[0].House.ID != 7 {
		t.Fatalf("unexpected provider result: %#v", result)
	}
	if recommender.Mode() != "fake-provider" {
		t.Fatalf("unexpected recommender mode %q", recommender.Mode())
	}
}

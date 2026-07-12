package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

type RecommenderService struct {
	repository domain.HouseRepository
}

func NewRecommender(repository domain.HouseRepository) *RecommenderService {
	return &RecommenderService{repository: repository}
}

func (service *RecommenderService) Recommend(
	ctx context.Context,
	request domain.RecommendationRequest,
) ([]domain.Recommendation, error) {
	limit := request.Limit
	if limit <= 0 || limit > 20 {
		limit = 6
	}

	houses, err := service.repository.ListHouses(ctx, domain.HouseFilter{
		City:       request.City,
		District:   request.District,
		MaxRent:    request.MaxRent,
		Bedrooms:   request.Bedrooms,
		Limit:      100,
		OnlyActive: true,
	})
	if err != nil {
		return nil, err
	}

	need := strings.ToLower(strings.TrimSpace(request.Need))
	recommendations := make([]domain.Recommendation, 0, len(houses))
	for _, house := range houses {
		score, reasons := scoreHouse(house, request, need)
		recommendations = append(recommendations, domain.Recommendation{
			House:  house,
			Score:  score,
			Reason: strings.Join(reasons, "，"),
		})
	}

	sort.SliceStable(recommendations, func(i, j int) bool {
		if recommendations[i].Score == recommendations[j].Score {
			return recommendations[i].House.MonthlyRent < recommendations[j].House.MonthlyRent
		}
		return recommendations[i].Score > recommendations[j].Score
	})

	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}
	return recommendations, nil
}

func scoreHouse(
	house domain.House,
	request domain.RecommendationRequest,
	need string,
) (float64, []string) {
	score := 45.0
	reasons := make([]string, 0, 4)

	if request.District != "" && strings.EqualFold(house.District, request.District) {
		score += 20
		reasons = append(reasons, "区域完全匹配")
	}
	if request.MaxRent > 0 {
		switch {
		case house.MonthlyRent <= int(float64(request.MaxRent)*0.85):
			score += 18
			reasons = append(reasons, "租金低于预算且留有余量")
		case house.MonthlyRent <= request.MaxRent:
			score += 12
			reasons = append(reasons, "租金符合预算")
		}
	}
	if request.Bedrooms > 0 && house.Bedrooms >= request.Bedrooms {
		score += 12
		reasons = append(reasons, fmt.Sprintf("提供%d间卧室", house.Bedrooms))
	}

	searchable := strings.ToLower(strings.Join([]string{
		house.Title,
		house.Description,
		house.District,
		strings.Join(house.Amenities, " "),
	}, " "))
	for _, token := range strings.Fields(need) {
		if len([]rune(token)) >= 2 && strings.Contains(searchable, token) {
			score += 4
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "综合房源条件与需求排序")
	}
	if score > 99 {
		score = 99
	}
	return score, reasons
}

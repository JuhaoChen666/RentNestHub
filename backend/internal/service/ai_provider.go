package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

type AIProviderConfig struct {
	URL             string
	APIKey          string
	Model           string
	Thinking        string
	ReasoningEffort string
}

func NewRecommendationProvider(config AIProviderConfig) RecommendationProvider {
	if strings.TrimSpace(config.URL) == "" ||
		strings.TrimSpace(config.APIKey) == "" ||
		strings.TrimSpace(config.Model) == "" {
		return LocalRecommendationProvider{}
	}
	return HTTPRecommendationProvider{
		config:   config,
		client:   &http.Client{Timeout: 35 * time.Second},
		fallback: LocalRecommendationProvider{},
	}
}

type HTTPRecommendationProvider struct {
	config   AIProviderConfig
	client   *http.Client
	fallback LocalRecommendationProvider
}

func (HTTPRecommendationProvider) Mode() string {
	return "ai-http"
}

func (provider HTTPRecommendationProvider) Recommend(
	ctx context.Context,
	houses []domain.House,
	request domain.RecommendationRequest,
	limit int,
) ([]domain.Recommendation, error) {
	recommendations, err := provider.recommend(ctx, houses, request, limit)
	if err != nil || len(recommendations) == 0 {
		return provider.fallback.Recommend(ctx, houses, request, limit)
	}
	return recommendations, nil
}

func (provider HTTPRecommendationProvider) recommend(
	ctx context.Context,
	houses []domain.House,
	request domain.RecommendationRequest,
	limit int,
) ([]domain.Recommendation, error) {
	payload, err := json.Marshal(map[string]any{
		"request": request,
		"houses":  houses,
		"limit":   limit,
	})
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(chatCompletionRequest{
		Model: provider.config.Model,
		Messages: []chatMessage{
			{Role: "system", Content: aiRecommendationSystemPrompt},
			{Role: "user", Content: string(payload)},
		},
		Thinking:        thinkingConfig{Type: provider.thinkingMode()},
		ReasoningEffort: provider.reasoningEffort(),
		Stream:          false,
	})
	if err != nil {
		return nil, err
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		provider.config.URL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+provider.config.APIKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := provider.client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("ai provider returned %s", response.Status)
	}

	var completion chatCompletionResponse
	if err := json.NewDecoder(response.Body).Decode(&completion); err != nil {
		return nil, err
	}
	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("ai provider returned no choices")
	}
	return recommendationsFromAIContent(completion.Choices[0].Message.Content, houses, limit)
}

func (provider HTTPRecommendationProvider) thinkingMode() string {
	if provider.config.Thinking == "enabled" || provider.config.Thinking == "disabled" {
		return provider.config.Thinking
	}
	return "disabled"
}

func (provider HTTPRecommendationProvider) reasoningEffort() string {
	switch provider.config.ReasoningEffort {
	case "low", "medium", "high":
		return provider.config.ReasoningEffort
	default:
		return "low"
	}
}

func recommendationsFromAIContent(
	content string,
	houses []domain.House,
	limit int,
) ([]domain.Recommendation, error) {
	var result aiRecommendationResponse
	if err := json.Unmarshal([]byte(extractJSONObject(content)), &result); err != nil {
		return nil, err
	}

	byID := make(map[int64]domain.House, len(houses))
	for _, house := range houses {
		byID[house.ID] = house
	}

	recommendations := make([]domain.Recommendation, 0, len(result.Recommendations))
	for _, item := range result.Recommendations {
		house, ok := byID[item.ID]
		if !ok {
			continue
		}
		recommendations = append(recommendations, domain.Recommendation{
			House:  house,
			Score:  item.Score,
			Reason: strings.TrimSpace(item.Reason),
		})
	}
	sort.SliceStable(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score
	})
	if len(recommendations) > limit {
		recommendations = recommendations[:limit]
	}
	return recommendations, nil
}

func extractJSONObject(content string) string {
	content = strings.TrimSpace(content)
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end < start {
		return content
	}
	return content[start : end+1]
}

const aiRecommendationSystemPrompt = `Rank rental houses for the tenant request.
Return only JSON in this shape:
{"recommendations":[{"id":1,"score":92,"reason":"short Chinese reason"}]}
Use only ids from the provided houses.`

type chatCompletionRequest struct {
	Model           string         `json:"model"`
	Messages        []chatMessage  `json:"messages"`
	Thinking        thinkingConfig `json:"thinking"`
	ReasoningEffort string         `json:"reasoning_effort"`
	Stream          bool           `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type thinkingConfig struct {
	Type string `json:"type"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type aiRecommendationResponse struct {
	Recommendations []struct {
		ID     int64   `json:"id"`
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	} `json:"recommendations"`
}

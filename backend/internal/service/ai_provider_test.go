package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

func TestHTTPRecommendationProviderUsesAIResponse(t *testing.T) {
	provider := HTTPRecommendationProvider{
		config: AIProviderConfig{
			URL:    "https://ai.example.test/chat",
			APIKey: "test-key",
			Model:  "test-model",
		},
		client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if request.URL.String() != "https://ai.example.test/chat" {
				t.Fatalf("unexpected url %s", request.URL)
			}
			if request.Header.Get("Authorization") != "Bearer test-key" {
				t.Fatalf("missing auth header")
			}
			return completionResponse(t, `{"recommendations":[{"id":2,"score":96,"reason":"更符合预算"}]}`), nil
		})},
		fallback: LocalRecommendationProvider{},
	}
	result, err := provider.Recommend(context.Background(), []domain.House{
		{ID: 1, MonthlyRent: 7000},
		{ID: 2, MonthlyRent: 5200},
	}, domain.RecommendationRequest{}, 3)
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if len(result) != 1 || result[0].House.ID != 2 || result[0].Reason != "更符合预算" {
		t.Fatalf("unexpected recommendations: %#v", result)
	}
}

func TestHTTPRecommendationProviderFallsBackOnProviderError(t *testing.T) {
	provider := HTTPRecommendationProvider{
		config: AIProviderConfig{
			URL:    "https://ai.example.test/chat",
			APIKey: "test-key",
			Model:  "test-model",
		},
		client: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Status:     "502 Bad Gateway",
				Body:       io.NopCloser(strings.NewReader("upstream failed")),
				Header:     make(http.Header),
			}, nil
		})},
		fallback: LocalRecommendationProvider{},
	}
	result, err := provider.Recommend(context.Background(), []domain.House{
		{ID: 1, District: "浦东新区", MonthlyRent: 7200, Bedrooms: 2},
		{ID: 2, District: "徐汇区", MonthlyRent: 5200, Bedrooms: 2},
	}, domain.RecommendationRequest{
		District: "徐汇区",
		MaxRent:  6000,
		Bedrooms: 2,
	}, 2)
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if len(result) != 2 || result[0].House.ID != 2 {
		t.Fatalf("expected fallback recommendations, got %#v", result)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func completionResponse(t *testing.T, content string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(chatCompletionResponse{
		Choices: []struct {
			Message chatMessage `json:"message"`
		}{
			{Message: chatMessage{Role: "assistant", Content: content}},
		},
	}); err != nil {
		t.Fatalf("write completion: %v", err)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(&body),
		Header:     make(http.Header),
	}
}

package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

type Dependencies struct {
	Repository    domain.HouseRepository
	Recommender   domain.Recommender
	UploadDir     string
	PublicBaseURL string
	Logger        *slog.Logger
}

type API struct {
	repository    domain.HouseRepository
	recommender   domain.Recommender
	uploadDir     string
	publicBaseURL string
	logger        *slog.Logger
}

func New(dependencies Dependencies) http.Handler {
	api := &API{
		repository:    dependencies.Repository,
		recommender:   dependencies.Recommender,
		uploadDir:     dependencies.UploadDir,
		publicBaseURL: dependencies.PublicBaseURL,
		logger:        dependencies.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", api.health)
	mux.HandleFunc("GET /api/v1/houses", api.listHouses)
	mux.HandleFunc("GET /api/v1/houses/{id}", api.getHouse)
	mux.HandleFunc("POST /api/v1/houses", api.createHouse)
	mux.HandleFunc("POST /api/v1/recommendations", api.recommend)
	mux.HandleFunc("POST /api/v1/favorites", api.addFavorite)
	mux.HandleFunc("GET /api/v1/favorites/{tenantId}", api.listFavorites)
	mux.HandleFunc("DELETE /api/v1/favorites/{tenantId}/{houseId}", api.removeFavorite)
	mux.HandleFunc("POST /api/v1/messages", api.createMessage)
	mux.HandleFunc("GET /api/v1/messages/{senderId}", api.listMessages)

	return api.recoverPanic(api.logRequests(api.cors(mux)))
}

func (api *API) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

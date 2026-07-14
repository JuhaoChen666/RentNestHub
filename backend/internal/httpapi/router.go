package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
)

type Dependencies struct {
	Repository    domain.HouseRepository
	Accounts      domain.AccountRepository
	ResetCodes    domain.PasswordResetStore
	Mailer        domain.PasswordResetMailer
	Recommender   domain.Recommender
	UploadDir     string
	PublicBaseURL string
	Logger        *slog.Logger
}

type API struct {
	repository    domain.HouseRepository
	accounts      domain.AccountRepository
	resetCodes    domain.PasswordResetStore
	mailer        domain.PasswordResetMailer
	recommender   domain.Recommender
	sessions      *sessionStore
	uploadDir     string
	publicBaseURL string
	logger        *slog.Logger
}

func New(dependencies Dependencies) http.Handler {
	api := &API{
		repository:    dependencies.Repository,
		accounts:      dependencies.Accounts,
		resetCodes:    dependencies.ResetCodes,
		mailer:        dependencies.Mailer,
		recommender:   dependencies.Recommender,
		sessions:      newSessionStore(),
		uploadDir:     dependencies.UploadDir,
		publicBaseURL: dependencies.PublicBaseURL,
		logger:        dependencies.Logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", api.health)
	mux.HandleFunc("POST /api/v1/auth/register", api.register)
	mux.HandleFunc("POST /api/v1/auth/login", api.login)
	mux.HandleFunc("GET /api/v1/auth/me", api.me)
	mux.HandleFunc("PATCH /api/v1/auth/me", api.updateProfile)
	mux.HandleFunc("POST /api/v1/auth/password-reset/request", api.requestPasswordReset)
	mux.HandleFunc("POST /api/v1/auth/password-reset/confirm", api.resetPassword)
	mux.HandleFunc("GET /api/v1/houses", api.listHouses)
	mux.HandleFunc("GET /api/v1/houses/{id}", api.getHouse)
	mux.HandleFunc("POST /api/v1/houses", api.createHouse)
	mux.HandleFunc("GET /api/v1/admin/houses/pending", api.listPendingHouseReviews)
	mux.HandleFunc("PATCH /api/v1/admin/houses/{id}/review", api.reviewHouse)
	mux.HandleFunc("POST /api/v1/recommendations", api.recommend)
	mux.HandleFunc("POST /api/v1/favorites", api.addFavorite)
	mux.HandleFunc("GET /api/v1/favorites/{tenantId}", api.listFavorites)
	mux.HandleFunc("DELETE /api/v1/favorites/{tenantId}/{houseId}", api.removeFavorite)
	mux.HandleFunc("POST /api/v1/messages", api.createMessage)
	mux.HandleFunc("GET /api/v1/messages", api.listMessages)

	return api.recoverPanic(api.logRequests(api.cors(mux)))
}

func (api *API) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

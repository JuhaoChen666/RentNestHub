package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/JuhaoChen666/RentNestHub/backend/internal/domain"
	"github.com/JuhaoChen666/RentNestHub/backend/internal/repository/mysqlrepo"
)

const sessionLifetime = 24 * time.Hour

type authRequest struct {
	Identifier  string `json:"identifier"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

func (api *API) requestPasswordReset(writer http.ResponseWriter, request *http.Request) {
	var input authRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid email")
		return
	}
	user, err := api.accounts.FindUserByIdentifier(request.Context(), email)
	if errors.Is(err, mysqlrepo.ErrNotFound) {
		writer.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	code, err := resetCode()
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	if err := api.resetCodes.Save(request.Context(), email, code, 10*time.Minute); err != nil {
		api.internalError(writer, request, err)
		return
	}
	if err := api.mailer.SendPasswordReset(request.Context(), user.Email, user.DisplayName, code); err != nil {
		api.internalError(writer, request, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (api *API) resetPassword(writer http.ResponseWriter, request *http.Request) {
	var input authRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	code := strings.TrimSpace(input.Code)
	if _, err := mail.ParseAddress(email); err != nil || len(code) != 6 ||
		len([]rune(input.NewPassword)) < 6 || len([]rune(input.NewPassword)) > 128 {
		writeError(writer, http.StatusBadRequest, "invalid password reset details")
		return
	}
	valid, err := api.resetCodes.VerifyAndConsume(request.Context(), email, code)
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	if !valid {
		writeError(writer, http.StatusBadRequest, "invalid or expired verification code")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	if err := api.accounts.UpdateUserPassword(request.Context(), email, string(hash)); err != nil {
		api.internalError(writer, request, err)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func resetCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

type authResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]session
}

type session struct {
	user      domain.User
	expiresAt time.Time
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]session)}
}

func (store *sessionStore) issue(user domain.User) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	store.mu.Lock()
	store.sessions[token] = session{user: user, expiresAt: time.Now().Add(sessionLifetime)}
	store.mu.Unlock()
	return token, nil
}

func (store *sessionStore) user(token string) (domain.User, bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	session, ok := store.sessions[token]
	if !ok || time.Now().After(session.expiresAt) {
		delete(store.sessions, token)
		return domain.User{}, false
	}
	return session.user, true
}

func (store *sessionStore) update(token string, user domain.User) {
	store.mu.Lock()
	defer store.mu.Unlock()
	session, ok := store.sessions[token]
	if !ok || time.Now().After(session.expiresAt) {
		return
	}
	session.user = user
	store.sessions[token] = session
}

func (api *API) register(writer http.ResponseWriter, request *http.Request) {
	var input authRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	username := strings.TrimSpace(input.Username)
	displayName := strings.TrimSpace(input.DisplayName)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if len([]rune(username)) < 3 || len([]rune(username)) > 80 ||
		len([]rune(displayName)) < 1 || len([]rune(displayName)) > 80 ||
		len([]rune(input.Password)) < 6 || len([]rune(input.Password)) > 128 {
		writeError(writer, http.StatusBadRequest, "invalid account details")
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid email")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	user := domain.User{
		Role:         "tenant",
		Username:     username,
		DisplayName:  displayName,
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := api.accounts.CreateUser(request.Context(), &user); err != nil {
		writeError(writer, http.StatusConflict, "username or email already exists")
		return
	}
	api.writeSession(writer, user, http.StatusCreated)
}

func (api *API) login(writer http.ResponseWriter, request *http.Request) {
	var input authRequest
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	identifier := strings.TrimSpace(input.Identifier)
	if identifier == "" || input.Password == "" {
		writeError(writer, http.StatusBadRequest, "identifier and password are required")
		return
	}
	user, err := api.accounts.FindUserByIdentifier(request.Context(), identifier)
	if errors.Is(err, mysqlrepo.ErrNotFound) || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		writeError(writer, http.StatusUnauthorized, "invalid username, email, or password")
		return
	}
	if err != nil {
		api.internalError(writer, request, err)
		return
	}
	api.writeSession(writer, user, http.StatusOK)
}

func (api *API) me(writer http.ResponseWriter, request *http.Request) {
	user, ok := api.currentUser(request.Context(), request)
	if !ok {
		writeError(writer, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(writer, http.StatusOK, user)
}

func (api *API) updateProfile(writer http.ResponseWriter, request *http.Request) {
	user, ok := api.currentUser(request.Context(), request)
	if !ok {
		writeError(writer, http.StatusUnauthorized, "authentication required")
		return
	}
	var input struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid email")
		return
	}
	if email == user.Email {
		writeJSON(writer, http.StatusOK, user)
		return
	}
	if existing, err := api.accounts.FindUserByIdentifier(request.Context(), email); err == nil && existing.ID != user.ID {
		writeError(writer, http.StatusConflict, "email already exists")
		return
	} else if err != nil && !errors.Is(err, mysqlrepo.ErrNotFound) {
		api.internalError(writer, request, err)
		return
	}
	if err := api.accounts.UpdateUserEmail(request.Context(), user.ID, email); err != nil {
		api.internalError(writer, request, err)
		return
	}
	user.Email = email
	api.sessions.update(api.bearerToken(request), user)
	writeJSON(writer, http.StatusOK, user)
}

func (api *API) writeSession(writer http.ResponseWriter, user domain.User, status int) {
	user.PasswordHash = ""
	token, err := api.sessions.issue(user)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "could not create session")
		return
	}
	writeJSON(writer, status, authResponse{Token: token, User: user})
}

func (api *API) currentUser(_ context.Context, request *http.Request) (domain.User, bool) {
	token := api.bearerToken(request)
	if token == "" {
		return domain.User{}, false
	}
	return api.sessions.user(token)
}

func (api *API) bearerToken(request *http.Request) string {
	const prefix = "Bearer "
	header := request.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

package httpapi

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (writer *statusWriter) WriteHeader(status int) {
	writer.status = status
	writer.ResponseWriter.WriteHeader(status)
}

func (api *API) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		tracked := &statusWriter{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(tracked, request)
		api.logger.Info(
			"http request",
			"method", request.Method,
			"path", request.URL.Path,
			"status", tracked.status,
			"duration", time.Since(started),
		)
	})
}

func (api *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func (api *API) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				api.logger.Error("panic", "value", recovered, "stack", string(debug.Stack()))
				writeError(writer, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}

func (api *API) internalError(
	writer http.ResponseWriter,
	request *http.Request,
	err error,
) {
	api.logger.Error("request failed", "path", request.URL.Path, "error", err)
	writeError(writer, http.StatusInternalServerError, "internal server error")
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, message string) {
	writeJSON(writer, status, map[string]string{"error": message})
}

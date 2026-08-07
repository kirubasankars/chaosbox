package api

import (
	"log/slog"
	"net/http"

	"chaosbox/internal/httpx"
)

// LogErrorHandler emits a deliberate error-level slog event and returns 500.
// Used to exercise Observe / OTEL error-log pipelines in feature tests.
func LogErrorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		slog.Error("log.error_emitted", "reason", "feature_test")
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"status": "error",
			"msg":    "log.error_emitted",
		})
	}
}

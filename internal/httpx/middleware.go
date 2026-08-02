package httpx

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs every request as a structured slog event, with the
// level escalating to warn/error based on the response status code.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		durationMS := float64(time.Since(start).Microseconds()) / 1000.0
		level := slog.LevelInfo
		switch {
		case rec.status >= 500:
			level = slog.LevelError
		case rec.status >= 400:
			level = slog.LevelWarn
		}
		slog.Log(r.Context(), level, "http.request",
			"http.method", r.Method,
			"http.target", r.URL.RequestURI(),
			"http.status_code", rec.status,
			"http.remote_addr", r.RemoteAddr,
			"http.user_agent", r.UserAgent(),
			"duration_ms", durationMS,
		)
	})
}

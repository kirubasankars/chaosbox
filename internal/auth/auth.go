// Package auth provides a small token-based authentication middleware. When a
// shared token is configured, every request (except an explicit public
// allowlist) must present it via an Authorization: Bearer header, an
// X-Auth-Token header, or a token query parameter. Failed authentication is
// logged and answered with 401.
package auth

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"chaosbox/internal/httpx"
)

// HeaderName is a convenience header clients and peers may use to present the
// token in addition to the standard "Authorization: Bearer <token>" form.
const HeaderName = "X-Auth-Token"

// QueryParam lets browsers reach otherwise-protected pages (e.g. /ui) by
// appending ?token=<token>, since navigation cannot set request headers.
const QueryParam = "token"

// Middleware wraps next so requests must carry the configured token. When
// token is empty, authentication is disabled and next is returned unwrapped.
// Paths listed in public bypass the check (e.g. /metrics for scraping).
func Middleware(token string, next http.Handler, public ...string) http.Handler {
	if token == "" {
		return next
	}
	skip := make(map[string]bool, len(public))
	for _, p := range public {
		skip[p] = true
	}
	tokenBytes := []byte(token)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skip[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		presented, provided := TokenFromRequest(r)
		if !provided || subtle.ConstantTimeCompare([]byte(presented), tokenBytes) != 1 {
			reason := "missing_token"
			if provided {
				reason = "invalid_token"
			}
			slog.Warn("auth.failed",
				"http.method", r.Method,
				"http.target", r.URL.Path,
				"http.remote_addr", r.RemoteAddr,
				"auth.reason", reason,
			)
			httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// TokenFromRequest extracts a presented token from (in order) the
// Authorization: Bearer header, the X-Auth-Token header, or the token query
// parameter. The bool reports whether any token was supplied at all.
func TokenFromRequest(r *http.Request) (string, bool) {
	if h := r.Header.Get("Authorization"); h != "" {
		if bearer, ok := cutBearer(h); ok {
			return bearer, true
		}
	}
	if h := r.Header.Get(HeaderName); h != "" {
		return h, true
	}
	if q := r.URL.Query().Get(QueryParam); q != "" {
		return q, true
	}
	return "", false
}

func cutBearer(header string) (string, bool) {
	const prefix = "bearer "
	if len(header) >= len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):]), true
	}
	return "", false
}

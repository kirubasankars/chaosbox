package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"chaosbox/internal/auth"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestMiddleware_DisabledWhenTokenEmpty(t *testing.T) {
	h := auth.Middleware("", okHandler())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/count", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 when auth disabled", rec.Code)
	}
}

func TestMiddleware_AcceptsValidToken(t *testing.T) {
	const token = "s3cr3t"
	cases := []struct {
		name  string
		apply func(r *http.Request)
	}{
		{"bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+token) }},
		{"bearer_case_insensitive", func(r *http.Request) { r.Header.Set("Authorization", "bearer "+token) }},
		{"x_auth_token", func(r *http.Request) { r.Header.Set(auth.HeaderName, token) }},
		{"query_param", func(r *http.Request) {
			q := r.URL.Query()
			q.Set(auth.QueryParam, token)
			r.URL.RawQuery = q.Encode()
		}},
	}
	h := auth.Middleware(token, okHandler())
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/count", nil)
			tc.apply(req)
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200 for valid token via %s", rec.Code, tc.name)
			}
		})
	}
}

func TestMiddleware_RejectsMissingAndInvalidToken(t *testing.T) {
	h := auth.Middleware("s3cr3t", okHandler())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/count", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token: status = %d, want 401", rec.Code)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/count", nil)
	req.Header.Set("Authorization", "Bearer wrong")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token: status = %d, want 401", rec.Code)
	}
}

func TestMiddleware_PublicPathsBypassAuth(t *testing.T) {
	h := auth.Middleware("s3cr3t", okHandler(), "/metrics")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("public path /metrics: status = %d, want 200 without token", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/count", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("protected path /count: status = %d, want 401 without token", rec.Code)
	}
}

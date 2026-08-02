package counter

import (
	"net/http"

	"chaosbox/internal/httpx"
	"chaosbox/internal/metrics"
)

type countResponse struct {
	Count int64 `json:"count"`
}

func IncrHandler(c Counter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		n, err := c.Incr(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		metrics.ChaosboxCount.Set(float64(n))
		httpx.WriteJSON(w, http.StatusOK, countResponse{Count: n})
	}
}

func DecrHandler(c Counter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		n, err := c.Decr(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		metrics.ChaosboxCount.Set(float64(n))
		httpx.WriteJSON(w, http.StatusOK, countResponse{Count: n})
	}
}

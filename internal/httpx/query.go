package httpx

import (
	"net/http"
	"strconv"
	"time"
)

func QueryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func QueryFloat(r *http.Request, key string, def float64) float64 {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return n
}

// QueryPeriod parses the "period" query param (seconds) into a duration,
// falling back to defaultSec when absent or non-positive.
func QueryPeriod(r *http.Request, defaultSec int) time.Duration {
	sec := QueryInt(r, "period", defaultSec)
	if sec <= 0 {
		sec = defaultSec
	}
	return time.Duration(sec) * time.Second
}

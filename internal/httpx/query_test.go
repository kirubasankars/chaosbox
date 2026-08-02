package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chaosbox/internal/httpx"
)

func TestQueryPeriod_Seconds(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		defSec int
		want   time.Duration
	}{
		{name: "explicit seconds", query: "period=5", defSec: 10, want: 5 * time.Second},
		{name: "larger seconds", query: "period=30", defSec: 10, want: 30 * time.Second},
		{name: "one second", query: "period=1", defSec: 10, want: 1 * time.Second},
		{name: "missing uses default seconds", query: "", defSec: 10, want: 10 * time.Second},
		{name: "zero falls back to default seconds", query: "period=0", defSec: 10, want: 10 * time.Second},
		{name: "negative falls back to default seconds", query: "period=-3", defSec: 7, want: 7 * time.Second},
		{name: "non-numeric falls back to default seconds", query: "period=abc", defSec: 10, want: 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)
			got := httpx.QueryPeriod(req, tt.defSec)
			if got != tt.want {
				t.Fatalf("QueryPeriod(%q, %d) = %v, want %v", tt.query, tt.defSec, got, tt.want)
			}
			// period is always a whole number of seconds (never ms/min units).
			if got%time.Second != 0 {
				t.Fatalf("QueryPeriod(%q) = %v, not an integer number of seconds", tt.query, got)
			}
			if int(got/time.Second) != int(tt.want/time.Second) {
				t.Fatalf("QueryPeriod seconds = %d, want %d", got/time.Second, tt.want/time.Second)
			}
		})
	}
}

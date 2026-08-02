// Package ui serves a single-page control console for browsing status and
// triggering chaosbox APIs (counter, load simulators) from a browser.
package ui

import (
	_ "embed"
	"net/http"

	"chaosbox/internal/httpx"
	"chaosbox/internal/membership"
)

//go:embed index.html
var indexHTML []byte

// Handler serves the control console HTML page.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(indexHTML)
	}
}

// NodesHandler returns the membership snapshot as JSON for the console.
func NodesHandler(m *membership.Membership) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, m.Snapshot())
	}
}

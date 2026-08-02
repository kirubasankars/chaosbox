// Package api holds small, stateless HTTP handlers (health, file cat) that
// don't warrant their own domain package.
package api

import (
	"net/http"

	"chaosbox/internal/httpx"
)

// HealthHandler serves the / endpoint. loadStatus reports which load
// simulators (cpu/memory/disk) are currently running on this node.
func HealthHandler(version string, loadStatus func() map[string]bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.WritePrettyJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"version": version,
			"load":    loadStatus(),
		})
	}
}

package loadsim

import (
	"net/http"
	"time"

	"chaosbox/internal/httpx"
)

// RegisterHandlers wires the /load/{cpu,mem,disk,all}/{start,stop} routes to
// c. Each route is passed through wrap before registration; wrap is nil-safe
// and defaults to a no-op, but callers typically pass a peer fanout wrapper
// (e.g. Membership.Fanout) so that triggering a load action on this node
// also triggers it on known peers.
func RegisterHandlers(mux *http.ServeMux, c *Controller, wrap func(http.HandlerFunc) http.HandlerFunc) {
	if wrap == nil {
		wrap = func(h http.HandlerFunc) http.HandlerFunc { return h }
	}

	mux.HandleFunc("/load/cpu/start", wrap(func(w http.ResponseWriter, r *http.Request) {
		minPct := httpx.QueryFloat(r, "min_pct", DefaultCPUMinPct)
		maxPct := httpx.QueryFloat(r, "max_pct", DefaultCPUMaxPct)
		period := httpx.QueryPeriod(r, DefaultPeriodSec)
		c.StartCPU(minPct, maxPct, period)
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"status":  "started",
			"kind":    "cpu",
			"min_pct": minPct,
			"max_pct": maxPct,
			"period":  int(period / time.Second),
		})
	}))
	mux.HandleFunc("/load/cpu/stop", wrap(func(w http.ResponseWriter, r *http.Request) {
		c.StopCPU()
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "stopped", "kind": "cpu"})
	}))

	mux.HandleFunc("/load/mem/start", wrap(func(w http.ResponseWriter, r *http.Request) {
		mb := httpx.QueryInt(r, "mb", 0)
		period := httpx.QueryPeriod(r, DefaultPeriodSec)
		peakMB, err := c.StartMem(mb, period)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"status": "started",
			"kind":   "mem",
			"mb":     peakMB,
			"period": int(period / time.Second),
		})
	}))
	mux.HandleFunc("/load/mem/stop", wrap(func(w http.ResponseWriter, r *http.Request) {
		c.StopMem()
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "stopped", "kind": "mem"})
	}))

	mux.HandleFunc("/load/disk/start", wrap(func(w http.ResponseWriter, r *http.Request) {
		mb := httpx.QueryInt(r, "mb", DefaultDiskMB)
		period := httpx.QueryPeriod(r, DefaultPeriodSec)
		path, mb, err := c.StartDisk(mb, period)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"status": "started",
			"kind":   "disk",
			"data":   c.dataDir,
			"file":   path,
			"mb":     mb,
			"period": int(period / time.Second),
		})
	}))
	mux.HandleFunc("/load/disk/stop", wrap(func(w http.ResponseWriter, r *http.Request) {
		c.StopDisk()
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "stopped", "kind": "disk"})
	}))

	mux.HandleFunc("/load/all/start", wrap(func(w http.ResponseWriter, r *http.Request) {
		minPct := httpx.QueryFloat(r, "min_pct", DefaultCPUMinPct)
		maxPct := httpx.QueryFloat(r, "max_pct", DefaultCPUMaxPct)
		var mb, diskMB int
		if r.URL.Query().Get("mb") == "" {
			diskMB = DefaultDiskMB
		} else {
			mb = httpx.QueryInt(r, "mb", 0)
			diskMB = mb
		}
		period := httpx.QueryPeriod(r, DefaultPeriodSec)
		memMB, diskPath, err := c.StartAll(minPct, maxPct, mb, diskMB, period)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"status":  "started",
			"kind":    "all",
			"min_pct": minPct,
			"max_pct": maxPct,
			"mem_mb":  memMB,
			"disk_mb": diskMB,
			"data":    c.dataDir,
			"file":    diskPath,
			"period":  int(period / time.Second),
		})
	}))
	mux.HandleFunc("/load/all/stop", wrap(func(w http.ResponseWriter, r *http.Request) {
		c.StopAll()
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "stopped", "kind": "all"})
	}))
}

// Command chaosbox runs a small HTTP service used for infra demos and testing:
// a health/version endpoint, a file cat endpoint, a counter (in-memory or
// Redis), CPU/memory/disk load simulators, peer membership tracking, and a
// Prometheus /metrics endpoint. See README.md for the full API list.
package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"chaosbox/internal/api"
	"chaosbox/internal/config"
	"chaosbox/internal/counter"
	"chaosbox/internal/docs"
	"chaosbox/internal/httpx"
	"chaosbox/internal/loadsim"
	"chaosbox/internal/logging"
	"chaosbox/internal/membership"
	"chaosbox/internal/ui"
)

func main() {
	opts, err := parseCLI(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if err := applyCLIDefaults(&opts); err != nil {
		log.Fatal(err)
	}

	logOut, closeLog, err := setupLogOutput(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer closeLog()

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	selfAddr := config.SelfFromListen(cfg.Listen)
	logging.Init(logOut, cfg.Version, selfAddr)

	var ctr counter.Counter
	backend := "memory"
	if opts.RedisDSN != "" {
		rc, err := counter.NewRedis(opts.RedisDSN)
		if err != nil {
			logging.Fatal("redis.connect_failed", "error", err.Error())
		}
		ctr = rc
		backend = "redis"
	} else {
		ctr = counter.NewMemory()
	}
	slog.Info("counter.backend", "backend", backend)

	lc := loadsim.NewController(opts.DataDir)

	mem, err := membership.New(selfAddr, cfg.Version, cfg.Peers, cfg.PeerCheckSec, cfg.PeerCACert)
	if err != nil {
		logging.Fatal("membership.init_failed", "error", err.Error())
	}
	mem.SetSelfLoadStatus(lc.Status)
	mem.Start()
	slog.Info("membership.start",
		"peer.count", len(cfg.Peers),
		"peer.check_sec", int(mem.CheckInterval().Seconds()),
	)

	if opts.StartupDelay > 0 {
		slog.Info("startup.delay", "startup_delay_sec", opts.StartupDelay)
		time.Sleep(time.Duration(opts.StartupDelay) * time.Second)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", api.HealthHandler(cfg.Version, lc.Status))
	mux.HandleFunc("/_cat/file", api.CatFileHandler(opts.FilePath))
	mux.HandleFunc("/_cat/nodes", membership.NodesHandler(mem))
	mux.HandleFunc("/count", counter.GetHandler(ctr))
	mux.HandleFunc("/count/incr", counter.IncrHandler(ctr))
	mux.HandleFunc("/count/decr", counter.DecrHandler(ctr))
	loadsim.RegisterHandlers(mux, lc, mem.Fanout)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/docs", docs.UIHandler())
	mux.HandleFunc("/docs/openapi.yaml", docs.SpecHandler())
	mux.HandleFunc("/ui", ui.Handler())
	mux.HandleFunc("/ui/nodes", ui.NodesHandler(mem))
	mux.HandleFunc("/log/error", api.LogErrorHandler())

	handler := httpx.RequestLogger(mux)

	scheme := "http"
	if cfg.UseTLS() {
		scheme = "https"
	}
	slog.Info("server.listen",
		"server.scheme", scheme,
		"server.address", cfg.Listen,
		"server.tls", scheme == "https",
	)

	if scheme == "https" {
		if err := http.ListenAndServeTLS(cfg.Listen, cfg.TLSCert, cfg.TLSKey, handler); err != nil {
			logging.Fatal("server.stopped", "error", err.Error())
		}
		return
	}
	if err := http.ListenAndServe(cfg.Listen, handler); err != nil {
		logging.Fatal("server.stopped", "error", err.Error())
	}
}

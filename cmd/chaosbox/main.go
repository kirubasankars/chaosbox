// Command chaosbox runs a small HTTP service used for infra demos and testing:
// a health/version endpoint, a file cat endpoint, a counter (in-memory or
// Redis), CPU/memory/disk load simulators, peer membership tracking, and a
// Prometheus /metrics endpoint. See README.md for the full API list.
package main

import (
	"flag"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	configPath := flag.String("config", "", "path to config JSON (optional; built-in defaults apply when omitted)")
	startupDelay := flag.Int("startup-delay", 0, "delay in seconds before starting listeners")
	filePath := flag.String("file", "", "path to plain text file served by /_cat/file (default: <data>/file.txt)")
	dataDir := flag.String("data", "", "data folder for disk IO load (default: system temp dir)")
	logsDir := flag.String("logs", "", "logs folder for chaosbox.log (default: stdout only)")
	redisDSN := flag.String("redis", "", "Redis DSN (e.g. redis://localhost:6379/0); empty uses in-memory counter")
	flag.Parse()

	// Flag/config validation happens before the structured logger exists, so
	// stay on the stdlib logger for these early fatals.
	if *dataDir == "" {
		*dataDir = filepath.Join(os.TempDir(), "chaosbox", "data")
	}
	if *filePath == "" {
		*filePath = filepath.Join(*dataDir, "file.txt")
	}
	if *startupDelay < 0 {
		log.Fatal("-startup-delay must be >= 0")
	}

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		log.Fatalf("data dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(*filePath), 0o755); err != nil {
		log.Fatalf("file dir: %v", err)
	}
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		if err := os.WriteFile(*filePath, nil, 0o644); err != nil {
			log.Fatalf("file: %v", err)
		}
	} else if err != nil {
		log.Fatalf("file: %v", err)
	}

	var logOut io.Writer = os.Stdout
	if *logsDir != "" {
		if err := os.MkdirAll(*logsDir, 0o755); err != nil {
			log.Fatalf("logs dir: %v", err)
		}
		logFile, err := os.OpenFile(filepath.Join(*logsDir, "chaosbox.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			log.Fatalf("log file: %v", err)
		}
		defer logFile.Close()
		logOut = io.MultiWriter(os.Stdout, logFile)
	}

	cfg, err := config.Load(*configPath)
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
	if *redisDSN != "" {
		rc, err := counter.NewRedis(*redisDSN)
		if err != nil {
			logging.Fatal("redis.connect_failed", "error", err.Error())
		}
		ctr = rc
		backend = "redis"
	} else {
		ctr = counter.NewMemory()
	}
	slog.Info("counter.backend", "backend", backend)

	lc := loadsim.NewController(*dataDir)

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

	if *startupDelay > 0 {
		slog.Info("startup.delay", "startup_delay_sec", *startupDelay)
		time.Sleep(time.Duration(*startupDelay) * time.Second)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", api.HealthHandler(cfg.Version, lc.Status))
	mux.HandleFunc("/_cat/file", api.CatFileHandler(*filePath))
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

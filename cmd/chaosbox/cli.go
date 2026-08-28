package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type cliOptions struct {
	ConfigPath   string
	StartupDelay int
	FilePath     string
	DataDir      string
	LogsDir      string
	RedisDSN     string
}

func parseCLI(args []string) (cliOptions, error) {
	fs := flag.NewFlagSet("chaosbox", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to config JSON (optional; built-in defaults apply when omitted)")
	startupDelay := fs.Int("startup-delay", 0, "delay in seconds before starting listeners")
	filePath := fs.String("file", "", "path to plain text file served by /_cat/file (default: <data>/file.txt)")
	dataDir := fs.String("data", "", "data folder for disk IO load (default: system temp dir)")
	logsDir := fs.String("logs", "", "logs folder for chaosbox.log (default: stdout only)")
	redisDSN := fs.String("redis", "", "Redis DSN (e.g. redis://localhost:6379/0); empty uses in-memory counter")
	if err := fs.Parse(args); err != nil {
		return cliOptions{}, err
	}
	return cliOptions{
		ConfigPath:   *configPath,
		StartupDelay: *startupDelay,
		FilePath:     *filePath,
		DataDir:      *dataDir,
		LogsDir:      *logsDir,
		RedisDSN:     *redisDSN,
	}, nil
}

func applyCLIDefaults(opts *cliOptions) error {
	if opts.DataDir == "" {
		opts.DataDir = filepath.Join(os.TempDir(), "chaosbox", "data")
	}
	if opts.FilePath == "" {
		opts.FilePath = filepath.Join(opts.DataDir, "file.txt")
	}
	if opts.StartupDelay < 0 {
		return fmt.Errorf("-startup-delay must be >= 0")
	}

	if err := os.MkdirAll(opts.DataDir, 0o755); err != nil {
		return fmt.Errorf("data dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(opts.FilePath), 0o755); err != nil {
		return fmt.Errorf("file dir: %w", err)
	}
	if _, err := os.Stat(opts.FilePath); os.IsNotExist(err) {
		if err := os.WriteFile(opts.FilePath, nil, 0o644); err != nil {
			return fmt.Errorf("file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("file: %w", err)
	}
	return nil
}

func setupLogOutput(opts cliOptions) (io.Writer, func(), error) {
	if opts.LogsDir == "" {
		return os.Stdout, func() {}, nil
	}
	if err := os.MkdirAll(opts.LogsDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("logs dir: %w", err)
	}
	logFile, err := os.OpenFile(filepath.Join(opts.LogsDir, "chaosbox.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("log file: %w", err)
	}
	return io.MultiWriter(os.Stdout, logFile), func() { logFile.Close() }, nil
}

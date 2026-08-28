package main

import (
	"os"
	"path/filepath"
	"testing"

	"chaosbox/internal/config"
)

func TestParseCLI_OnlyStartupDelay(t *testing.T) {
	opts, err := parseCLI([]string{"-startup-delay", "2"})
	if err != nil {
		t.Fatalf("parseCLI: %v", err)
	}
	if opts.StartupDelay != 2 {
		t.Errorf("StartupDelay = %d, want 2", opts.StartupDelay)
	}
	if opts.ConfigPath != "" {
		t.Errorf("ConfigPath = %q, want empty", opts.ConfigPath)
	}
	if opts.LogsDir != "" {
		t.Errorf("LogsDir = %q, want empty", opts.LogsDir)
	}
	if opts.RedisDSN != "" {
		t.Errorf("RedisDSN = %q, want empty", opts.RedisDSN)
	}
	if opts.DataDir != "" {
		t.Errorf("DataDir = %q, want empty before defaults", opts.DataDir)
	}
	if opts.FilePath != "" {
		t.Errorf("FilePath = %q, want empty before defaults", opts.FilePath)
	}
}

func TestApplyCLIDefaults_OnlyStartupDelay(t *testing.T) {
	opts := cliOptions{StartupDelay: 2}
	if err := applyCLIDefaults(&opts); err != nil {
		t.Fatalf("applyCLIDefaults: %v", err)
	}

	wantData := filepath.Join(os.TempDir(), "chaosbox", "data")
	if opts.DataDir != wantData {
		t.Errorf("DataDir = %q, want %q", opts.DataDir, wantData)
	}
	wantFile := filepath.Join(wantData, "file.txt")
	if opts.FilePath != wantFile {
		t.Errorf("FilePath = %q, want %q", opts.FilePath, wantFile)
	}
	if _, err := os.Stat(opts.FilePath); err != nil {
		t.Errorf("file.txt should exist: %v", err)
	}

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Default()
	if cfg.Version != want.Version {
		t.Errorf("Version = %q, want %q", cfg.Version, want.Version)
	}
	if cfg.Listen != want.Listen {
		t.Errorf("Listen = %q, want %q", cfg.Listen, want.Listen)
	}
	if len(cfg.Peers) != 0 {
		t.Errorf("Peers = %v, want empty", cfg.Peers)
	}
}

func TestApplyCLIDefaults_NegativeStartupDelay(t *testing.T) {
	opts := cliOptions{StartupDelay: -1}
	if err := applyCLIDefaults(&opts); err == nil {
		t.Fatal("expected error for negative startup-delay")
	}
}

func TestParseCLI_DataPathPassed(t *testing.T) {
	dataDir := t.TempDir()
	opts, err := parseCLI([]string{"-data", dataDir})
	if err != nil {
		t.Fatalf("parseCLI: %v", err)
	}
	if opts.DataDir != dataDir {
		t.Errorf("DataDir = %q, want %q", opts.DataDir, dataDir)
	}
	if opts.FilePath != "" {
		t.Errorf("FilePath = %q, want empty before defaults", opts.FilePath)
	}
}

func TestApplyCLIDefaults_DataPathPassed(t *testing.T) {
	dataDir := t.TempDir()
	opts := cliOptions{DataDir: dataDir}
	if err := applyCLIDefaults(&opts); err != nil {
		t.Fatalf("applyCLIDefaults: %v", err)
	}
	if opts.DataDir != dataDir {
		t.Errorf("DataDir = %q, want passed path %q", opts.DataDir, dataDir)
	}
	wantFile := filepath.Join(dataDir, "file.txt")
	if opts.FilePath != wantFile {
		t.Errorf("FilePath = %q, want %q", opts.FilePath, wantFile)
	}
	if _, err := os.Stat(opts.FilePath); err != nil {
		t.Errorf("file.txt should exist under passed data dir: %v", err)
	}
}

func TestApplyCLIDefaults_DefaultDataPath(t *testing.T) {
	opts := cliOptions{}
	if err := applyCLIDefaults(&opts); err != nil {
		t.Fatalf("applyCLIDefaults: %v", err)
	}
	wantData := filepath.Join(os.TempDir(), "chaosbox", "data")
	if opts.DataDir != wantData {
		t.Errorf("DataDir = %q, want default %q", opts.DataDir, wantData)
	}
	wantFile := filepath.Join(wantData, "file.txt")
	if opts.FilePath != wantFile {
		t.Errorf("FilePath = %q, want default %q", opts.FilePath, wantFile)
	}
}

func TestParseCLI_LogsPathPassed(t *testing.T) {
	logsDir := t.TempDir()
	opts, err := parseCLI([]string{"-logs", logsDir})
	if err != nil {
		t.Fatalf("parseCLI: %v", err)
	}
	if opts.LogsDir != logsDir {
		t.Errorf("LogsDir = %q, want %q", opts.LogsDir, logsDir)
	}
}

func TestSetupLogOutput_LogsPathPassed(t *testing.T) {
	logsDir := t.TempDir()
	opts := cliOptions{LogsDir: logsDir}
	_, closeLog, err := setupLogOutput(opts)
	if err != nil {
		t.Fatalf("setupLogOutput: %v", err)
	}
	closeLog()

	logFile := filepath.Join(logsDir, "chaosbox.log")
	if _, err := os.Stat(logFile); err != nil {
		t.Errorf("chaosbox.log should exist under passed logs dir: %v", err)
	}
}

func TestSetupLogOutput_DefaultLogsPath(t *testing.T) {
	opts := cliOptions{}
	out, closeLog, err := setupLogOutput(opts)
	if err != nil {
		t.Fatalf("setupLogOutput: %v", err)
	}
	closeLog()
	if out != os.Stdout {
		t.Errorf("log output = %v, want os.Stdout when logs dir omitted", out)
	}
	if opts.LogsDir != "" {
		t.Errorf("LogsDir = %q, want empty default", opts.LogsDir)
	}
}

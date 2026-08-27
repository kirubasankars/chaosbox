package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Version != "0.1.0" {
		t.Errorf("Version = %q, want 0.1.0", cfg.Version)
	}
	if cfg.Listen != ":8080" {
		t.Errorf("Listen = %q, want :8080", cfg.Listen)
	}
	if cfg.PeerCheckSec != 5 {
		t.Errorf("PeerCheckSec = %d, want 5", cfg.PeerCheckSec)
	}
	if len(cfg.Peers) != 0 {
		t.Errorf("Peers = %v, want empty", cfg.Peers)
	}
	if cfg.TLSCert != "" || cfg.TLSKey != "" || cfg.PeerCACert != "" {
		t.Errorf("TLS/CA fields should be empty in default config")
	}
}

func TestLoadEmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	want := Default()
	if cfg.Version != want.Version || cfg.Listen != want.Listen || cfg.PeerCheckSec != want.PeerCheckSec {
		t.Errorf("Load(\"\") = %+v, want %+v", cfg, want)
	}
	if len(cfg.Peers) != 0 {
		t.Errorf("Peers = %v, want empty", cfg.Peers)
	}
}

func TestLoadMergesPartialJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	const body = `{"peers": ["peer-a:8080", "peer-b:8080"]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != "0.1.0" {
		t.Errorf("Version = %q, want default 0.1.0", cfg.Version)
	}
	if cfg.Listen != ":8080" {
		t.Errorf("Listen = %q, want default :8080", cfg.Listen)
	}
	if cfg.PeerCheckSec != 5 {
		t.Errorf("PeerCheckSec = %d, want default 5", cfg.PeerCheckSec)
	}
	if len(cfg.Peers) != 2 || cfg.Peers[0] != "peer-a:8080" {
		t.Errorf("Peers = %v, want [peer-a:8080 peer-b:8080]", cfg.Peers)
	}
}

func TestLoadExplicitEmptyPeers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	const body = `{"peers": []}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Peers == nil || len(cfg.Peers) != 0 {
		t.Errorf("Peers = %v, want empty slice", cfg.Peers)
	}
}

func TestValidateTLSMismatch(t *testing.T) {
	cfg := Default()
	cfg.TLSCert = "/path/to/cert.pem"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected TLS validation error when only cert is set")
	}
}

func TestValidateOK(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("Default().Validate() = %v", err)
	}
}

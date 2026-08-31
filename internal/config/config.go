// Package config loads and validates the JSON config file that drives
// chaosbox's listen address, TLS material, and peer membership settings.
package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Config struct {
	Version      string   `json:"version"`
	Listen       string   `json:"listen"`
	TLSCert      string   `json:"tls_cert"`
	TLSKey       string   `json:"tls_key"`
	Peers        []string `json:"peers"`
	PeerCheckSec int      `json:"peer_check_sec"`
	PeerCACert   string   `json:"peer_ca_cert"`
	// AuthToken, when non-empty, is a shared secret required on incoming
	// requests and sent on outgoing peer probes/fan-out. Empty disables auth.
	AuthToken string `json:"auth_token"`
}

// Default returns the built-in configuration used when no config file is
// provided or when JSON fields are omitted.
func Default() Config {
	return Config{
		Version:      "0.1.0",
		Listen:       ":8080",
		PeerCheckSec: 5,
	}
}

func Load(path string) (Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return merge(Default(), fileCfg), nil
}

func merge(base, file Config) Config {
	if file.Version != "" {
		base.Version = file.Version
	}
	if file.Listen != "" {
		base.Listen = file.Listen
	}
	if file.TLSCert != "" {
		base.TLSCert = file.TLSCert
	}
	if file.TLSKey != "" {
		base.TLSKey = file.TLSKey
	}
	if file.Peers != nil {
		base.Peers = file.Peers
	}
	if file.PeerCheckSec != 0 {
		base.PeerCheckSec = file.PeerCheckSec
	}
	if file.PeerCACert != "" {
		base.PeerCACert = file.PeerCACert
	}
	if file.AuthToken != "" {
		base.AuthToken = file.AuthToken
	}
	return base
}

func (c Config) Validate() error {
	if c.Listen == "" {
		return fmt.Errorf("listen is required")
	}
	if (c.TLSCert == "") != (c.TLSKey == "") {
		return fmt.Errorf("both tls_cert and tls_key are required for TLS")
	}
	return nil
}

func (c Config) UseTLS() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}

// SelfFromListen derives the address other nodes can use to reach this
// instance. ":8080", "0.0.0.0:8080", and "[::]:8080" become "127.0.0.1:8080".
func SelfFromListen(listen string) string {
	listen = strings.TrimSpace(listen)
	host, port, err := net.SplitHostPort(listen)
	if err != nil {
		return listen
	}
	switch host {
	case "", "0.0.0.0", "::", "[::]":
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}

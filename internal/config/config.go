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
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
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

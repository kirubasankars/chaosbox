// Package membership tracks cluster peers in memory: it periodically probes
// each peer's health endpoint and reports up/down status plus last-known
// version via /_cat/nodes.
package membership

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultCheckSec  = 5
	checkTimeout     = 2 * time.Second
	unreachableAtLog = 2 // consecutive fails before escalating warn -> error

	// fanoutHeader marks a request as already relayed by a peer's Fanout, so
	// the receiving node executes it locally but does not relay it again.
	// This keeps fan-out to a single hop regardless of peer topology.
	fanoutHeader = "X-Chaosbox-Fanout"
)

type NodeInfo struct {
	IP      string          `json:"ip"`
	Version string          `json:"version"`
	Status  string          `json:"status"` // "up" or "down"
	Self    bool            `json:"self"`
	Load    map[string]bool `json:"load,omitempty"` // last-known load simulator state (cpu/memory/disk)
}

type Membership struct {
	mu         sync.RWMutex
	selfIP     string
	version    string
	interval   time.Duration
	peers      map[string]*NodeInfo // key = normalized base URL
	order      []string             // stable order: self first, then peers as configured
	failCounts map[string]int       // consecutive unreachable checks per peer
	client     *http.Client
	selfLoad   func() map[string]bool // optional: reports this node's own load state
}

func normalizePeer(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return ""
	}
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return strings.TrimRight(addr, "/")
	}
	return "http://" + strings.TrimRight(addr, "/")
}

func peerDisplayIP(base string) string {
	base = strings.TrimPrefix(base, "https://")
	base = strings.TrimPrefix(base, "http://")
	return base
}

// buildClient constructs the HTTP client used to probe peers. When caCertPath
// is set, it is used as the sole trust root for verifying peer TLS certs
// (e.g. when peers are configured with https:// addresses and a private CA).
func buildClient(caCertPath string) (*http.Client, error) {
	client := &http.Client{Timeout: checkTimeout}
	if caCertPath == "" {
		return client, nil
	}

	pemData, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("read peer ca cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("parse peer ca cert: no valid certificates found in %s", caCertPath)
	}
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: pool},
	}
	return client, nil
}

// New builds a Membership tracking self plus the given peers. checkSec <= 0
// falls back to a 5 second probe interval. caCertPath, if non-empty, is used
// to validate peer TLS certs (for https peer addresses) against a private CA
// instead of the system trust store.
func New(self, version string, peers []string, checkSec int, caCertPath string) (*Membership, error) {
	if checkSec <= 0 {
		checkSec = defaultCheckSec
	}
	client, err := buildClient(caCertPath)
	if err != nil {
		return nil, err
	}

	selfBase := normalizePeer(self)
	m := &Membership{
		selfIP:     peerDisplayIP(selfBase),
		version:    version,
		interval:   time.Duration(checkSec) * time.Second,
		peers:      make(map[string]*NodeInfo),
		failCounts: make(map[string]int),
		client:     client,
	}

	if selfBase != "" {
		m.order = append(m.order, selfBase)
		m.peers[selfBase] = &NodeInfo{
			IP:      m.selfIP,
			Version: version,
			Status:  "up",
			Self:    true,
		}
	}

	seen := map[string]bool{selfBase: true}
	for _, p := range peers {
		base := normalizePeer(p)
		if base == "" || seen[base] {
			continue
		}
		seen[base] = true
		m.order = append(m.order, base)
		m.peers[base] = &NodeInfo{
			IP:     peerDisplayIP(base),
			Status: "down",
		}
	}
	return m, nil
}

// CheckInterval returns the resolved probe interval (after defaulting).
func (m *Membership) CheckInterval() time.Duration {
	return m.interval
}

// SetSelfLoadStatus registers a callback used to report this node's own
// load simulator state (cpu/memory/disk) in Snapshot, so /_cat/nodes can
// show it without membership depending on the loadsim package directly.
func (m *Membership) SetSelfLoadStatus(fn func() map[string]bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.selfLoad = fn
}

func (m *Membership) Start() {
	go func() {
		m.checkPeers()
		t := time.NewTicker(m.interval)
		defer t.Stop()
		for range t.C {
			m.checkPeers()
		}
	}()
}

func (m *Membership) checkPeers() {
	m.mu.RLock()
	bases := make([]string, 0, len(m.order))
	for _, base := range m.order {
		if info := m.peers[base]; info != nil && !info.Self {
			bases = append(bases, base)
		}
	}
	m.mu.RUnlock()

	type result struct {
		base    string
		up      bool
		version string
		load    map[string]bool
	}
	results := make([]result, len(bases))
	var wg sync.WaitGroup
	for i, base := range bases {
		wg.Add(1)
		go func(i int, base string) {
			defer wg.Done()
			up, version, load := m.probe(base)
			results[i] = result{base: base, up: up, version: version, load: load}
		}(i, base)
	}
	wg.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, r := range results {
		info := m.peers[r.base]
		if info == nil || info.Self {
			continue
		}
		if r.up {
			if info.Status == "down" || m.failCounts[r.base] > 0 {
				slog.Info("peer.reachable",
					"peer.ip", info.IP,
					"peer.version", r.version,
					"peer.fail_count", m.failCounts[r.base],
				)
			}
			info.Status = "up"
			m.failCounts[r.base] = 0
			if r.version != "" {
				info.Version = r.version
			}
			if r.load != nil {
				info.Load = r.load
			}
			continue
		}

		info.Status = "down"
		m.failCounts[r.base]++
		n := m.failCounts[r.base]
		attrs := []any{
			"peer.ip", info.IP,
			"peer.fail_count", n,
			"peer.status", "down",
		}
		if n > unreachableAtLog {
			slog.Error("peer.unreachable", attrs...)
		} else {
			slog.Warn("peer.unreachable", attrs...)
		}
	}

	for _, info := range m.peers {
		if info.Self {
			info.Status = "up"
			info.Version = m.version
		}
	}
}

func (m *Membership) probe(base string) (up bool, version string, load map[string]bool) {
	resp, err := m.client.Get(base + "/")
	if err != nil {
		return false, "", nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, "", nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return true, "", nil
	}
	var payload struct {
		Status  string          `json:"status"`
		Version string          `json:"version"`
		Load    map[string]bool `json:"load"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return true, "", nil
	}
	return true, payload.Version, payload.Load
}

// Fanout wraps a handler so that, after it runs locally, the same request
// (method, path, and query) is replicated best-effort to every known peer.
// Used so that calling an action (e.g. a /load/* start or stop) on one node
// triggers the same action on its peers. Requests already relayed by a
// peer's Fanout are executed locally only, so fan-out never cascades beyond
// one hop.
func (m *Membership) Fanout(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
		if r.Header.Get(fanoutHeader) != "" {
			return
		}
		method, path, rawQuery := r.Method, r.URL.Path, r.URL.RawQuery
		go m.broadcast(method, path, rawQuery)
	}
}

func (m *Membership) broadcast(method, path, rawQuery string) {
	m.mu.RLock()
	bases := make([]string, 0, len(m.order))
	for _, base := range m.order {
		if info := m.peers[base]; info != nil && !info.Self {
			bases = append(bases, base)
		}
	}
	m.mu.RUnlock()

	for _, base := range bases {
		go m.relay(base, method, path, rawQuery)
	}
}

func (m *Membership) relay(base, method, path, rawQuery string) {
	target := base + path
	if rawQuery != "" {
		target += "?" + rawQuery
	}

	req, err := http.NewRequest(method, target, nil)
	if err != nil {
		slog.Warn("peer.fanout_failed", "peer.ip", peerDisplayIP(base), "path", path, "error", err.Error())
		return
	}
	req.Header.Set(fanoutHeader, "1")

	resp, err := m.client.Do(req)
	if err != nil {
		slog.Warn("peer.fanout_failed", "peer.ip", peerDisplayIP(base), "path", path, "error", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Warn("peer.fanout_failed", "peer.ip", peerDisplayIP(base), "path", path, "status_code", resp.StatusCode)
		return
	}
	slog.Info("peer.fanout_ok", "peer.ip", peerDisplayIP(base), "path", path)
}

func (m *Membership) Snapshot() []NodeInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]NodeInfo, 0, len(m.order))
	for _, base := range m.order {
		info := m.peers[base]
		if info == nil {
			continue
		}
		n := *info
		if n.Self && m.selfLoad != nil {
			n.Load = m.selfLoad()
		}
		out = append(out, n)
	}
	return out
}

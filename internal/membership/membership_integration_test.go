package membership_test

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"chaosbox/internal/membership"
)

// pollUntil polls cond every interval until it returns true or timeout
// elapses, failing the test with msg on timeout.
func pollUntil(t *testing.T, timeout, interval time.Duration, msg string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for: %s", msg)
		}
		time.Sleep(interval)
	}
}

// fakeNodeServer starts an httptest server that answers GET / like a real
// chaosbox node's health handler would, reporting the given version and load
// state.
func fakeNodeServer(t *testing.T, version string, load map[string]bool) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"version": version,
			"load":    load,
		})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func trimScheme(base string) string {
	base = strings.TrimPrefix(base, "https://")
	return strings.TrimPrefix(base, "http://")
}

func TestMembership_ProbesPeers(t *testing.T) {
	peerA := fakeNodeServer(t, "1.2.3", map[string]bool{"cpu": true, "memory": false, "disk": false})
	peerB := fakeNodeServer(t, "1.2.3", map[string]bool{"cpu": false, "memory": false, "disk": false})

	m, err := membership.New("self:0", "test", []string{peerA.URL, peerB.URL}, 1, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.Start()

	pollUntil(t, 3*time.Second, 50*time.Millisecond, "peers reported up with load", func() bool {
		nodes := m.Snapshot()
		if len(nodes) != 3 { // self + 2 peers
			return false
		}
		upCount := 0
		for _, n := range nodes {
			if n.Self {
				continue
			}
			if n.Status == "up" && n.Version == "1.2.3" {
				upCount++
			}
		}
		return upCount == 2
	})

	nodes := m.Snapshot()
	byIP := map[string]membership.NodeInfo{}
	for _, n := range nodes {
		byIP[n.IP] = n
	}
	a, ok := byIP[trimScheme(peerA.URL)]
	if !ok {
		t.Fatalf("peer A not found in snapshot: %+v", nodes)
	}
	if !a.Load["cpu"] {
		t.Errorf("peer A load.cpu = %v, want true", a.Load["cpu"])
	}
	b, ok := byIP[trimScheme(peerB.URL)]
	if !ok {
		t.Fatalf("peer B not found in snapshot: %+v", nodes)
	}
	if b.Load["cpu"] {
		t.Errorf("peer B load.cpu = %v, want false", b.Load["cpu"])
	}
}

// TestMembership_ReusesPeerConnections verifies that repeated probes to the
// same peer reuse a pooled keep-alive connection instead of dialing a new TCP
// connection each interval. The peer server counts newly-accepted connections
// via ConnState; after several sequential probes only one connection should
// have been opened.
func TestMembership_ReusesPeerConnections(t *testing.T) {
	var newConns, requests atomic.Int64

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	srv := httptest.NewUnstartedServer(mux)
	srv.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			newConns.Add(1)
		}
	}
	srv.Start()
	defer srv.Close()

	m, err := membership.New("self:0", "test", []string{srv.URL}, 1, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.Start()

	// Wait until the peer has been probed several times (across intervals).
	pollUntil(t, 5*time.Second, 20*time.Millisecond, "peer probed multiple times", func() bool {
		return requests.Load() >= 3
	})

	if got := newConns.Load(); got != 1 {
		t.Fatalf("opened %d connections for %d probes; want a single reused connection", got, requests.Load())
	}
}

func TestMembership_UnreachablePeerMarkedDown(t *testing.T) {
	up := fakeNodeServer(t, "1.0.0", nil)

	down := httptest.NewServer(http.NotFoundHandler())
	down.Close() // closed immediately: connections to it are refused

	m, err := membership.New("self:0", "test", []string{up.URL, down.URL}, 1, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.Start()

	pollUntil(t, 3*time.Second, 50*time.Millisecond, "up/down peers settle", func() bool {
		statuses := map[string]string{}
		for _, n := range m.Snapshot() {
			if !n.Self {
				statuses[n.IP] = n.Status
			}
		}
		return statuses[trimScheme(up.URL)] == "up" && statuses[trimScheme(down.URL)] == "down"
	})
}

func TestMembership_SelfLoadStatus(t *testing.T) {
	m, err := membership.New("self:0", "test", nil, 5, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.SetSelfLoadStatus(func() map[string]bool {
		return map[string]bool{"cpu": true, "memory": false, "disk": true}
	})

	nodes := m.Snapshot()
	if len(nodes) != 1 || !nodes[0].Self {
		t.Fatalf("expected a single self node, got %+v", nodes)
	}
	if !nodes[0].Load["cpu"] || nodes[0].Load["memory"] || !nodes[0].Load["disk"] {
		t.Errorf("self load = %+v, want cpu=true memory=false disk=true", nodes[0].Load)
	}
}

// TestMembership_Fanout_SingleHop wires two Memberships together as mutual
// peers (as main.go does across real chaosbox nodes) and verifies that calling
// a Fanout-wrapped handler on node A also invokes it on node B exactly
// once, with no further relay back to A (which would otherwise cascade
// forever in a mutual-peer topology).
func TestMembership_Fanout_SingleHop(t *testing.T) {
	var countA, countB atomic.Int64

	muxA := http.NewServeMux()
	muxB := http.NewServeMux()

	// srv references are needed before constructing each Membership's
	// peer list, so start plain httptest servers first, then attach the
	// real handlers (backed by the Membership.Fanout of each side) once
	// both URLs are known.
	srvA := httptest.NewServer(muxA)
	defer srvA.Close()
	srvB := httptest.NewServer(muxB)
	defer srvB.Close()

	memA, err := membership.New(srvA.URL, "test", []string{srvB.URL}, 5, "")
	if err != nil {
		t.Fatalf("New A: %v", err)
	}
	memB, err := membership.New(srvB.URL, "test", []string{srvA.URL}, 5, "")
	if err != nil {
		t.Fatalf("New B: %v", err)
	}

	muxA.HandleFunc("/probe", memA.Fanout(func(w http.ResponseWriter, r *http.Request) {
		countA.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	muxB.HandleFunc("/probe", memB.Fanout(func(w http.ResponseWriter, r *http.Request) {
		countB.Add(1)
		w.WriteHeader(http.StatusOK)
	}))

	resp, err := http.Get(srvA.URL + "/probe")
	if err != nil {
		t.Fatalf("GET /probe on A: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	pollUntil(t, 2*time.Second, 20*time.Millisecond, "fanout reaches B once", func() bool {
		return countA.Load() == 1 && countB.Load() == 1
	})

	// Give any errant extra relay a chance to land, then assert it didn't.
	time.Sleep(200 * time.Millisecond)
	if a, b := countA.Load(), countB.Load(); a != 1 || b != 1 {
		t.Fatalf("countA=%d countB=%d; want 1,1 (fan-out must not cascade beyond one hop)", a, b)
	}
}

func TestMembership_NodesHandler_ShowsLoadState(t *testing.T) {
	m, err := membership.New("self:0", "test", nil, 5, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	m.SetSelfLoadStatus(func() map[string]bool {
		return map[string]bool{"cpu": true, "memory": false, "disk": true}
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_cat/nodes", nil)
	membership.NodesHandler(m)(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "cpu,disk") {
		t.Errorf("_cat/nodes body missing expected load column %q, got:\n%s", "cpu,disk", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain prefix", ct)
	}
}

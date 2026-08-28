package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func portInUse(hostPort string) bool {
	conn, err := net.DialTimeout("tcp", hostPort, 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func buildChaosboxBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "chaosbox")
	cmd := exec.Command("go", "build", "-trimpath", "-o", bin, ".")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}

func healthReachable(ctx context.Context, url string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, err
	}
	return body.Status == "ok", nil
}

func TestMain_OnlyStartupDelay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}

	const addr = "127.0.0.1:8080"
	if portInUse(addr) {
		t.Skipf("%s already in use", addr)
	}

	bin := buildChaosboxBinary(t)
	const delaySec = 1
	cmd := exec.Command(bin, "-startup-delay", "1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start chaosbox: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	url := "http://" + addr + "/"
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if ok, _ := healthReachable(ctx, url); ok {
		t.Fatal("health endpoint reachable before startup delay elapsed")
	}

	deadline := time.Now().Add(time.Duration(delaySec+2) * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		ok, err := healthReachable(ctx, url)
		cancel()
		if err == nil && ok {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("health endpoint not reachable within %ds after startup delay", delaySec+2)
}

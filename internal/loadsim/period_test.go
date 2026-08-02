package loadsim

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"chaosbox/internal/metrics"
)

func TestPulseFraction_PeriodInSeconds(t *testing.T) {
	period := 4 * time.Second

	// Triangle wave: 0 at start/end, 1 at mid-period.
	cases := []struct {
		elapsed time.Duration
		want    float64
	}{
		{0, 0},
		{1 * time.Second, 0.5}, // quarter: rising
		{2 * time.Second, 1},   // half: peak
		{3 * time.Second, 0.5}, // three-quarter: falling
		{4 * time.Second, 0},   // full cycle wraps
		{5 * time.Second, 0.5}, // into next cycle
	}
	for _, tc := range cases {
		got := pulseFraction(tc.elapsed, period)
		if math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("pulseFraction(%v, %v) = %v, want %v", tc.elapsed, period, got, tc.want)
		}
	}
}

func TestPulseFraction_DifferentSecondPeriods(t *testing.T) {
	// At the midpoint of any period-in-seconds, fraction must be 1.
	for _, sec := range []int{1, 2, 5, 10, 60} {
		period := time.Duration(sec) * time.Second
		got := pulseFraction(period/2, period)
		if math.Abs(got-1) > 1e-9 {
			t.Errorf("pulseFraction(mid, %ds) = %v, want 1", sec, got)
		}
		got = pulseFraction(period, period)
		if math.Abs(got) > 1e-9 {
			t.Errorf("pulseFraction(end, %ds) = %v, want 0", sec, got)
		}
	}
}

func TestLoadHandlers_PeriodQueryIsSeconds(t *testing.T) {
	dir := t.TempDir()
	c := NewController(dir)
	mux := http.NewServeMux()
	RegisterHandlers(mux, c, nil)
	t.Cleanup(func() { c.StopAll() })

	paths := []string{
		"/load/cpu/start?period=3",
		"/load/mem/start?mb=1&period=3",
		"/load/disk/start?mb=1&period=3",
		"/load/all/start?mb=1&period=3",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			var body struct {
				Period int `json:"period"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v; body=%s", err, rec.Body.String())
			}
			if body.Period != 3 {
				t.Fatalf("period = %d, want 3 (seconds)", body.Period)
			}
		})
	}
}

func TestLoadHandlers_DefaultPeriodIsTenSeconds(t *testing.T) {
	dir := t.TempDir()
	c := NewController(dir)
	mux := http.NewServeMux()
	RegisterHandlers(mux, c, nil)
	t.Cleanup(func() { c.StopAll() })

	req := httptest.NewRequest(http.MethodPost, "/load/cpu/start", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Period int `json:"period"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Period != DefaultPeriodSec {
		t.Fatalf("period = %d, want default %d seconds", body.Period, DefaultPeriodSec)
	}
}

func TestStartCPU_PeriodDrivesTargetOverSeconds(t *testing.T) {
	dir := t.TempDir()
	c := NewController(dir)
	t.Cleanup(func() { c.StopCPU() })

	// 2s period, 10..100% pulse: target should rise toward peak and fall again.
	// min must be > 0 because StartCPU treats <= 0 as "use default".
	c.StartCPU(10, 100, 2*time.Second)

	deadline := time.Now().Add(3 * time.Second)
	var sawHigh, sawLow bool
	var last float64
	for time.Now().Before(deadline) {
		last = testutil.ToFloat64(metrics.LoadCPUTargetPercent)
		if last >= 80 {
			sawHigh = true
		}
		if last <= 30 {
			sawLow = true
		}
		if sawHigh && sawLow {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("CPU target did not pulse over a 2s period; last=%v sawHigh=%v sawLow=%v",
		last, sawHigh, sawLow)
}

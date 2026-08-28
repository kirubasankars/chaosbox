// Package loadsim implements pulsing CPU, memory, and disk IO load
// simulators used to generate visible, fluctuating load for infrastructure
// demos (autoscaling, monitoring, capacity).
package loadsim

import (
	"context"
	"crypto/sha256"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"chaosbox/internal/metrics"
)

const (
	diskLoadFileName = "chaosbox-load.bin"

	DefaultPeriodSec = 10
	DefaultDiskMB    = 64
	DefaultCPUMinPct = 60
	DefaultCPUMaxPct = 80

	memAvailableFrac = 0.60
	cpuTick          = 100 * time.Millisecond
)

// Controller owns the lifecycle of the three load simulators. All Start*/Stop*
// methods are safe for concurrent use.
type Controller struct {
	dataDir string

	mu sync.Mutex

	cpuCancel  context.CancelFunc
	memCancel  context.CancelFunc
	diskCancel context.CancelFunc

	memHeld atomic.Int64
}

func NewController(dataDir string) *Controller {
	c := &Controller{dataDir: dataDir}
	metrics.LoadRunning.WithLabelValues("cpu").Set(0)
	metrics.LoadRunning.WithLabelValues("mem").Set(0)
	metrics.LoadRunning.WithLabelValues("disk").Set(0)
	metrics.LoadCPUTargetPercent.Set(0)
	metrics.LoadMemHeldBytes.Set(0)
	return c
}

// pulseFraction maps elapsed time within a period to a 0..1..0 triangle wave,
// producing the up-then-down pulse shared by all three simulators.
func pulseFraction(elapsed, period time.Duration) float64 {
	if period <= 0 {
		return 0
	}
	p := float64(elapsed%period) / float64(period)
	if p < 0.5 {
		return p * 2
	}
	return 2 - p*2
}

func (c *Controller) StartCPU(minPct, maxPct float64, period time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopCPULocked()

	if minPct <= 0 {
		minPct = DefaultCPUMinPct
	}
	if maxPct <= 0 {
		maxPct = DefaultCPUMaxPct
	}
	if maxPct < minPct {
		maxPct = minPct
	}
	if period <= 0 {
		period = DefaultPeriodSec * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cpuCancel = cancel
	metrics.LoadRunning.WithLabelValues("cpu").Set(1)

	n := runtime.NumCPU()
	var target atomic.Uint64
	target.Store(uint64(math.Round(minPct)))
	metrics.LoadCPUTargetPercent.Set(minPct)

	go func() {
		start := time.Now()
		t := time.NewTicker(cpuTick)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				frac := pulseFraction(time.Since(start), period)
				pct := minPct + (maxPct-minPct)*frac
				target.Store(uint64(math.Round(pct)))
				metrics.LoadCPUTargetPercent.Set(pct)
			}
		}
	}()

	for i := 0; i < n; i++ {
		go func() {
			var sink uint64
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				pct := float64(target.Load())
				busy := time.Duration(float64(cpuTick) * pct / 100)
				idle := cpuTick - busy
				deadline := time.Now().Add(busy)
				for time.Now().Before(deadline) {
					select {
					case <-ctx.Done():
						return
					default:
						h := sha256.Sum256([]byte{byte(sink), byte(sink >> 8)})
						sink += uint64(h[0]) + uint64(h[1])<<8
					}
				}
				if idle > 0 {
					select {
					case <-ctx.Done():
						return
					case <-time.After(idle):
					}
				}
			}
		}()
	}
}

func (c *Controller) StopCPU() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopCPULocked()
}

func (c *Controller) stopCPULocked() {
	if c.cpuCancel != nil {
		c.cpuCancel()
		c.cpuCancel = nil
	}
	metrics.LoadRunning.WithLabelValues("cpu").Set(0)
	metrics.LoadCPUTargetPercent.Set(0)
}

func (c *Controller) StartMem(mbOverride int, period time.Duration) (int, error) {
	peakMB, err := resolveMemPeakMB(mbOverride)
	if err != nil {
		return 0, err
	}
	if period <= 0 {
		period = DefaultPeriodSec * time.Second
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopMemLocked()

	ctx, cancel := context.WithCancel(context.Background())
	c.memCancel = cancel
	metrics.LoadRunning.WithLabelValues("mem").Set(1)

	go func() {
		defer func() {
			c.memHeld.Store(0)
			metrics.LoadMemHeldBytes.Set(0)
			runtime.GC()
		}()

		var held [][]byte
		start := time.Now()
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		peakBytes := int64(peakMB) << 20
		const chunk = 1 << 20 // 1 MiB chunks

		for {
			select {
			case <-ctx.Done():
				held = nil
				return
			case <-ticker.C:
				frac := pulseFraction(time.Since(start), period)
				want := int64(float64(peakBytes) * frac)
				want = (want / chunk) * chunk

				current := int64(len(held)) * chunk
				for current < want {
					b := make([]byte, chunk)
					for i := 0; i < len(b); i += 4096 {
						b[i] = 1
					}
					held = append(held, b)
					current += chunk
				}
				for current > want && len(held) > 0 {
					held[len(held)-1] = nil
					held = held[:len(held)-1]
					current -= chunk
				}
				if current < want/2 {
					runtime.GC()
				}
				c.memHeld.Store(current)
				metrics.LoadMemHeldBytes.Set(float64(current))
			}
		}
	}()

	return peakMB, nil
}

func (c *Controller) StopMem() {
	c.mu.Lock()
	c.stopMemLocked()
	c.mu.Unlock()
	runtime.GC()
}

func (c *Controller) stopMemLocked() {
	if c.memCancel != nil {
		c.memCancel()
		c.memCancel = nil
	}
	metrics.LoadRunning.WithLabelValues("mem").Set(0)
	c.memHeld.Store(0)
	metrics.LoadMemHeldBytes.Set(0)
}

func (c *Controller) StartDisk(mb int, period time.Duration) (string, int, error) {
	if mb <= 0 {
		mb = DefaultDiskMB
	}
	if period <= 0 {
		period = DefaultPeriodSec * time.Second
	}

	// Prepare the file without holding the lock. A concurrent StartDisk doing
	// the same thing is harmless: both open the same path and the last writer
	// wins. The lock is only held for the final cancel/start swap.
	path := filepath.Join(c.dataDir, diskLoadFileName)
	wantSize := int64(mb) << 20

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return "", 0, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return "", 0, err
	}
	size := info.Size()
	if size < wantSize {
		if err := f.Truncate(wantSize); err != nil {
			_ = f.Close()
			return "", 0, err
		}
		buf := make([]byte, 1<<20)
		for i := range buf {
			buf[i] = byte(i)
		}
		off := size
		for off < wantSize {
			n := wantSize - off
			if n > int64(len(buf)) {
				n = int64(len(buf))
			}
			if _, err := f.WriteAt(buf[:n], off); err != nil {
				_ = f.Close()
				return "", 0, err
			}
			off += n
		}
		_ = f.Sync()
		size = wantSize
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopDiskLocked()

	ctx, cancel := context.WithCancel(context.Background())
	c.diskCancel = cancel
	metrics.LoadRunning.WithLabelValues("disk").Set(1)

	go func() {
		defer f.Close()

		buf := make([]byte, 1<<20)
		start := time.Now()
		var off int64

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			frac := pulseFraction(time.Since(start), period)
			if frac < 0.05 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(200 * time.Millisecond):
				}
				continue
			}

			n := int64(len(buf))
			if off+n > size {
				off = 0
			}
			wn, err := f.WriteAt(buf, off)
			if err != nil && err != io.ErrShortWrite {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			_ = f.Sync()
			rn, _ := f.ReadAt(buf, off)
			metrics.LoadDiskBytesTotal.Add(float64(wn + rn))
			off += n

			idle := time.Duration((1 - frac) * float64(50*time.Millisecond))
			if idle > 0 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(idle):
				}
			}
		}
	}()

	return path, mb, nil
}

func (c *Controller) StopDisk() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopDiskLocked()
}

func (c *Controller) stopDiskLocked() {
	if c.diskCancel != nil {
		c.diskCancel()
		c.diskCancel = nil
	}
	metrics.LoadRunning.WithLabelValues("disk").Set(0)
}

func (c *Controller) StartAll(minPct, maxPct float64, mbOverride, diskMB int, period time.Duration) (memMB int, diskPath string, err error) {
	c.StartCPU(minPct, maxPct, period)
	memMB, err = c.StartMem(mbOverride, period)
	if err != nil {
		c.StopCPU()
		return 0, "", err
	}
	diskPath, diskMB, err = c.StartDisk(diskMB, period)
	if err != nil {
		c.StopCPU()
		c.StopMem()
		return 0, "", err
	}
	return memMB, diskPath, nil
}

func (c *Controller) StopAll() {
	c.StopCPU()
	c.StopMem()
	c.StopDisk()
}

// Status reports which load simulators are currently running.
func (c *Controller) Status() map[string]bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return map[string]bool{
		"cpu":    c.cpuCancel != nil,
		"memory": c.memCancel != nil,
		"disk":   c.diskCancel != nil,
	}
}

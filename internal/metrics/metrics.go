// Package metrics defines the Prometheus collectors shared across chaosbox's
// counter and load-simulator features, exposed via GET /metrics.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ChaosboxCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chaosbox_count",
		Help: "Current value of the /count counter",
	})

	LoadRunning = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "chaosbox_load_running",
		Help: "Whether a load simulator is running (1) or stopped (0)",
	}, []string{"kind"})

	LoadCPUTargetPercent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chaosbox_load_cpu_target_percent",
		Help: "Current CPU duty-cycle target percent",
	})

	LoadMemHeldBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "chaosbox_load_mem_held_bytes",
		Help: "Bytes currently held by the memory load simulator",
	})

	LoadDiskBytesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "chaosbox_load_disk_bytes_total",
		Help: "Total bytes read and written by the disk load simulator",
	})
)

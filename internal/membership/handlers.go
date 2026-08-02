package membership

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// loadKinds fixes the display order of load simulator kinds in /_cat/nodes.
var loadKinds = []string{"cpu", "memory", "disk"}

// formatLoad renders a node's load state as a short, deterministic string,
// e.g. "cpu,disk" or "-" when nothing is running or the state is unknown.
func formatLoad(load map[string]bool) string {
	if load == nil {
		return "-"
	}
	var running []string
	for _, kind := range loadKinds {
		if load[kind] {
			running = append(running, kind)
		}
	}
	if len(running) == 0 {
		return "-"
	}
	return strings.Join(running, ",")
}

// NodesHandler renders self + peers as a plain-text, table-formatted list.
func NodesHandler(m *Membership) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodes := m.Snapshot()

		ipW, verW, stW, selfW, loadW := len("ip"), len("version"), len("status"), len("self"), len("load")
		rows := make([][5]string, 0, len(nodes))
		for _, n := range nodes {
			self := ""
			if n.Self {
				self = "*"
			}
			row := [5]string{n.IP, n.Version, n.Status, self, formatLoad(n.Load)}
			rows = append(rows, row)
			if len(row[0]) > ipW {
				ipW = len(row[0])
			}
			if len(row[1]) > verW {
				verW = len(row[1])
			}
			if len(row[2]) > stW {
				stW = len(row[2])
			}
			if len(row[3]) > selfW {
				selfW = len(row[3])
			}
			if len(row[4]) > loadW {
				loadW = len(row[4])
			}
		}

		var b strings.Builder
		fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %-*s  %-*s\n", ipW, "ip", verW, "version", stW, "status", selfW, "self", loadW, "load")
		for _, row := range rows {
			fmt.Fprintf(&b, "%-*s  %-*s  %-*s  %-*s  %-*s\n", ipW, row[0], verW, row[1], stW, row[2], selfW, row[3], loadW, row[4])
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, b.String())
	}
}

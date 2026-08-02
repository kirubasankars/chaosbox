package loadsim

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func availableMemoryBytes() (uint64, error) {
	if n, ok := cgroupAvailableMemory(); ok {
		return n, nil
	}
	return hostAvailableMemory()
}

func cgroupAvailableMemory() (uint64, bool) {
	// cgroup v2
	if limit, ok := readUintFile("/sys/fs/cgroup/memory.max"); ok {
		usage, uok := readUintFile("/sys/fs/cgroup/memory.current")
		if uok && limit > usage {
			return limit - usage, true
		}
		if limit > 0 && !uok {
			return limit, true
		}
	}
	// cgroup v1
	if limit, ok := readUintFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); ok {
		// ignore absurd host-sized "unlimited" limits
		if limit > 1<<62 {
			return 0, false
		}
		usage, uok := readUintFile("/sys/fs/cgroup/memory/memory.usage_in_bytes")
		if uok && limit > usage {
			return limit - usage, true
		}
		if limit > 0 {
			return limit, true
		}
	}
	return 0, false
}

func hostAvailableMemory() (uint64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "MemAvailable:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			break
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, err
		}
		return kb * 1024, nil
	}
	if err := sc.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("MemAvailable not found in /proc/meminfo")
}

func readUintFile(path string) (uint64, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	s := strings.TrimSpace(string(b))
	if s == "" || s == "max" {
		return 0, false
	}
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// resolveMemPeakMB caps mbOverride to memAvailableFrac of currently available
// memory, or resolves the cap itself when no override is given.
func resolveMemPeakMB(mbOverride int) (int, error) {
	avail, err := availableMemoryBytes()
	if err != nil {
		if mbOverride > 0 {
			return mbOverride, nil
		}
		return 0, fmt.Errorf("available memory: %w", err)
	}
	ceiling := int(float64(avail) * memAvailableFrac / float64(1<<20))
	if ceiling < 1 {
		ceiling = 1
	}
	if mbOverride > 0 {
		if mbOverride > ceiling {
			return ceiling, nil
		}
		return mbOverride, nil
	}
	return ceiling, nil
}

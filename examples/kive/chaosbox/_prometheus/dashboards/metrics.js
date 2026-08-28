// Copyright 2025 Kiruba Sankar Swaminathan. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

/**
 * Chaosbox PromQL helpers for Observe dashboard plugins.
 * Shared lib module (kind: lib) — not listed on the Monitoring hub.
 */

export const meta = { title: "Chaosbox metrics", kind: "lib" };

export const CHAOSBOX_JOB = "chaosbox";
export const DISK_RATE_WINDOW = "1m";

export function jobSel(job = CHAOSBOX_JOB) {
  return `kive_job="${String(job).replace(/\\/g, "\\\\").replace(/"/g, '\\"')}"`;
}

export function byHost(expr) {
  return `label_replace(${expr}, "host", "$1", "instance", "([^:]+):.*")`;
}

export function upExpr(job = CHAOSBOX_JOB) {
  return `up{${jobSel(job)}}`;
}

export function scrapeExpr(job = CHAOSBOX_JOB) {
  return `scrape_duration_seconds{${jobSel(job)}}`;
}

/** Shared Redis/memory counter — use max so multi-instance does not multiply. */
export function countExpr(job = CHAOSBOX_JOB) {
  return `max(chaosbox_count{${jobSel(job)}})`;
}

export function loadRunningExpr(job = CHAOSBOX_JOB) {
  return `chaosbox_load_running{${jobSel(job)}}`;
}

export function loadCpuTargetExpr(job = CHAOSBOX_JOB) {
  return `chaosbox_load_cpu_target_percent{${jobSel(job)}}`;
}

export function loadMemHeldExpr(job = CHAOSBOX_JOB) {
  return `chaosbox_load_mem_held_bytes{${jobSel(job)}}`;
}

export function loadDiskBytesRateExpr(job = CHAOSBOX_JOB) {
  return `rate(chaosbox_load_disk_bytes_total{${jobSel(job)}}[${DISK_RATE_WINDOW}])`;
}

export const CHAOSBOX_COLORS = {
  series: {
    light: { a: "#2563EB", b: "#7C3AED", c: "#059669", d: "#E11D48", e: "#F59E0B" },
    dark: { a: "#60A5FA", b: "#A78BFA", c: "#34D399", d: "#FB7185", e: "#FBBF24" },
  },
};

export function pickThemeColors(map) {
  const theme = document.documentElement.getAttribute("data-theme") === "dark" ? "dark" : "light";
  return map[theme] || map.light || map;
}

export function seriesPalette() {
  return Object.values(pickThemeColors(CHAOSBOX_COLORS.series));
}

// Copyright 2025 Kiruba Sankar Swaminathan. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

/**
 * Chaosbox load simulator dashboard — CPU / mem / disk load metrics.
 */

import {
  CHAOSBOX_JOB,
  byHost,
  loadCpuTargetExpr,
  loadDiskBytesRateExpr,
  loadMemHeldExpr,
  loadRunningExpr,
  seriesPalette,
} from "./metrics.js";

export const meta = { title: "Load", kind: "list" };

function first(rows) {
  if (!rows?.length) return null;
  const n = parseFloat(rows[0].value[1]);
  return isFinite(n) ? n : null;
}

function formatBytes(v) {
  if (!isFinite(v) || v < 0) return "—";
  const units = ["B", "KiB", "MiB", "GiB", "TiB"];
  let n = v;
  let i = 0;
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024;
    i += 1;
  }
  return `${n.toFixed(n >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
}

function formatBytesAxis(v) {
  if (!isFinite(v) || v <= 0) return "0";
  return formatBytes(v);
}

function formatBytesRate(v) {
  if (!isFinite(v) || v < 0) return "—";
  return `${formatBytes(v)}/s`;
}

function formatPct(v) {
  if (!isFinite(v)) return "—";
  return `${v.toFixed(v >= 10 ? 0 : 1)}%`;
}

export default async function render(ctx) {
  const { job = CHAOSBOX_JOB, prom, kit, href } = ctx;
  const { buildDashboard, formatCount, esc } = kit;
  const palette = () => seriesPalette();

  if (
    !(await ctx.requirePrometheus({
      retry: () => render(ctx),
      errorTitle: "Could not load load metrics",
    }))
  ) {
    return;
  }

  await buildDashboard(ctx, {
    page: {
      title: "Load",
      lead: `CPU, memory, and disk load simulators (<span class="mono">${esc(job)}</span>).`,
      crumbs: [{ label: "Monitoring", href: href.monitoring }, { label: "Load" }],
    },
    summary: { id: "loadSummary", loading: "Loading load…" },
    stats: [
      { id: "statRunning", label: "Kinds running" },
      { id: "statCpu", label: "CPU target" },
      { id: "statMem", label: "Mem held" },
      { id: "statDisk", label: "Disk rate" },
    ],
    timeBar: { initial: "1h" },
    sections: [
      {
        id: "running",
        title: "Simulators",
        cols: 2,
        charts: [
          {
            id: "runningGraph",
            title: "Load running by kind",
            expr: () => loadRunningExpr(job),
            labelKey: "kind",
            topN: 8,
            plotType: "line",
            includeZero: true,
            format: "count",
            colorPool: palette,
          },
          {
            id: "cpuGraph",
            title: "CPU target percent",
            expr: () => byHost(loadCpuTargetExpr(job)),
            labelKey: "host",
            topN: 8,
            plotType: "line",
            includeZero: true,
            yFormat: formatPct,
            yAxisFormat: (v) => (isFinite(v) ? `${Math.round(v)}%` : "0"),
            marginLeft: 48,
            colorPool: palette,
          },
        ],
      },
      {
        id: "resources",
        title: "Resources",
        cols: 2,
        charts: [
          {
            id: "memGraph",
            title: "Memory held",
            expr: () => byHost(loadMemHeldExpr(job)),
            labelKey: "host",
            topN: 8,
            plotType: "line",
            includeZero: true,
            yFormat: formatBytes,
            yAxisFormat: formatBytesAxis,
            marginLeft: 56,
            colorPool: palette,
          },
          {
            id: "diskGraph",
            title: "Disk bytes rate",
            expr: () => byHost(loadDiskBytesRateExpr(job)),
            labelKey: "host",
            topN: 8,
            plotType: "line",
            includeZero: true,
            yFormat: formatBytesRate,
            yAxisFormat: formatBytesAxis,
            marginLeft: 56,
            colorPool: palette,
          },
        ],
      },
    ],
    live: {
      intervalMs: 2000,
      summary: "#loadSummary",
      async refresh({ updateStats, summaryEl }) {
        const [running, cpu, mem, disk] = await Promise.all([
          prom.fetchInstant(`sum(${loadRunningExpr(job)})`),
          prom.fetchInstant(`max(${loadCpuTargetExpr(job)})`),
          prom.fetchInstant(`max(${loadMemHeldExpr(job)})`),
          prom.fetchInstant(`max(${loadDiskBytesRateExpr(job)})`),
        ]);
        const runningN = first(running) || 0;
        updateStats({
          statRunning: formatCount(runningN),
          statCpu: formatPct(first(cpu) ?? 0),
          statMem: formatBytes(first(mem) ?? 0),
          statDisk: formatBytesRate(first(disk) ?? 0),
        });
        if (summaryEl) {
          summaryEl.innerHTML = `<span class="live-dot" aria-hidden="true"></span>${formatCount(runningN)} kind(s) running · live`;
        }
      },
    },
  });
}

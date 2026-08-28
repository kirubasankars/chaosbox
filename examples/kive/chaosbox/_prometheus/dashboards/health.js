// Copyright 2025 Kiruba Sankar Swaminathan. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

/**
 * Chaosbox health — scrape availability and duration across workers.
 */

import {
  CHAOSBOX_JOB,
  byHost,
  jobSel,
  scrapeExpr,
  seriesPalette,
  upExpr,
} from "./metrics.js";

export const meta = { title: "Health", kind: "list" };

function first(rows) {
  if (!rows?.length) return null;
  const n = parseFloat(rows[0].value[1]);
  return isFinite(n) ? n : null;
}

function maxValue(rows) {
  let value = null;
  (rows || []).forEach((row) => {
    const n = parseFloat(row.value[1]);
    if (isFinite(n) && (value == null || n > value)) value = n;
  });
  return value;
}

function formatSeconds(v) {
  if (!isFinite(v)) return "—";
  if (v < 0.001) return "<1 ms";
  if (v < 1) return `${(v * 1000).toFixed(0)} ms`;
  return `${v.toFixed(v >= 10 ? 1 : 2)} s`;
}

function formatSecondsAxis(v) {
  if (!isFinite(v) || v <= 0) return "0";
  return v < 1 ? `${Math.round(v * 1000)}ms` : `${v.toFixed(1)}s`;
}

export default async function render(ctx) {
  const { job = CHAOSBOX_JOB, prom, kit, href } = ctx;
  const { buildDashboard, esc } = kit;
  const j = jobSel(job);
  const palette = () => seriesPalette();

  if (
    !(await ctx.requirePrometheus({
      retry: () => render(ctx),
      errorTitle: "Could not load health",
    }))
  ) {
    return;
  }

  await buildDashboard(ctx, {
    page: {
      title: "Health",
      lead: `Scrape availability and duration across chaosbox workers (<span class="mono">${esc(job)}</span>).`,
      crumbs: [{ label: "Monitoring", href: href.monitoring }, { label: "Health" }],
    },
    summary: { id: "healthSummary", loading: "Loading health…" },
    stats: [
      { id: "statUp", label: "Up" },
      { id: "statScrape", label: "Max scrape" },
    ],
    timeBar: { initial: "1h" },
    sections: [
      {
        id: "avail",
        title: "Availability",
        cols: 2,
        charts: [
          {
            id: "upGraph",
            title: "Scrape availability",
            expr: () => byHost(upExpr(job)),
            labelKey: "host",
            topN: 8,
            plotType: "line",
            includeZero: true,
            format: "count",
            colorPool: palette,
          },
          {
            id: "scrapeGraph",
            title: "Scrape duration",
            expr: () => byHost(scrapeExpr(job)),
            labelKey: "host",
            topN: 8,
            plotType: "line",
            includeZero: true,
            yFormat: formatSeconds,
            yAxisFormat: formatSecondsAxis,
            marginLeft: 52,
            colorPool: palette,
          },
        ],
      },
    ],
    live: {
      summary: "#healthSummary",
      async refresh({ updateStats, summaryEl }) {
        const [up, total, scrape] = await Promise.all([
          prom.fetchInstant(`count(up{${j}} == 1)`),
          prom.fetchInstant(`count(up{${j}})`),
          prom.fetchInstant(`max(${scrapeExpr(job)})`),
        ]);
        const upN = first(up) || 0;
        const totalN = first(total) || 0;
        updateStats({
          statUp: `${Math.round(upN)}/${Math.round(totalN)}`,
          statScrape: formatSeconds(maxValue(scrape)),
        });
        if (summaryEl) {
          summaryEl.innerHTML = `<span class="live-dot" aria-hidden="true"></span>${Math.round(upN)} of ${Math.round(totalN)} up · live`;
        }
      },
    },
  });
}

// Copyright 2025 Kiruba Sankar Swaminathan. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

/**
 * Chaosbox counter dashboard — live value + history of chaosbox_count.
 */

import { countExpr } from "./metrics.js";

export const meta = { title: "Counter", kind: "list" };

export default async function render(ctx) {
  const { job, prom, kit } = ctx;
  const { buildDashboard, formatCount, setLiveText } = kit;

  if (
    !(await ctx.requirePrometheus({
      retry: () => render(ctx),
      errorTitle: "Could not load counter",
    }))
  ) {
    return;
  }

  const expr = () => countExpr(job);

  await buildDashboard(ctx, {
    summary: { id: "counterSummary", loading: "Loading counter…" },
    stats: [{ id: "statCount", label: "Counter" }],
    timeBar: { initial: "1h" },
    sections: [
      {
        id: "counter",
        title: "Counter",
        cols: 1,
        charts: [
          {
            id: "countGraph",
            expr,
            seriesKey: "count",
            plotType: "line",
            format: "count",
            includeZero: true,
          },
        ],
      },
    ],
    live: {
      intervalMs: 2000,
      summary: "#counterSummary",
      async refresh({ updateStats, summaryEl }) {
        // max() so a shared Redis counter is not multiplied by instance count.
        const results = await prom.fetchInstant(expr());
        const v = results?.[0] ? parseFloat(results[0].value[1]) : NaN;
        const n = isFinite(v) ? v : null;
        updateStats({ statCount: formatCount(n ?? 0) });
        if (!summaryEl) return;
        if (n == null) {
          setLiveText(summaryEl, "No chaosbox_count series");
        } else {
          summaryEl.innerHTML = `<span class="live-dot" aria-hidden="true"></span>Counter ${formatCount(n)} · live`;
        }
      },
    },
  });
}

# Advanced: Prometheus metrics

Demo **only** metrics scrape — not counter, load, or peers in the same session.

## Prerequisites

- A `prometheus` job in your bucket (copy from
  [kive-go `examples/prometheus`](https://github.com/kive-sh/kive/tree/main/examples/prometheus))
- This `chaosbox` job with `_prometheus/scrape.yaml`

## Install

```bash
cp -R examples/kive/advanced/metrics/chaosbox workspace/jobs/chaosbox
kive build
kive deploy --jobs chaosbox
# redeploy prometheus if scrape config changed
kive deploy --jobs prometheus
```

## Demo (metrics only)

```bash
export BASE_URL=http://<worker>:<chaosbox_http_port>
curl -s "${BASE_URL}/metrics" | head
```

In Observe **Monitoring**, confirm the `chaosbox` scrape target is up.

## Phase 2

Observe dashboard plugins (`health`, `counter`, `load`) are not included yet.
Add `_prometheus/dashboards/*.js` when wiring Observe dashboard demos.

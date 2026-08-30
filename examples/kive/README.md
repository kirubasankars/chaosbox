# chaosbox on Kive

Complete Compose job for deploying chaosbox as a single-node Kive workload.
Uses the published `ghcr.io/kive-sh/chaosbox:latest` image (no `config.json`
mount), HTTP readiness, Prometheus scrape, Observe dashboards, alerts, and
Docker log labels for Observe Logs.

## Prerequisites

- Bucket with at least one worker labeled `worker` (see [First deploy](https://kive.sh/start/02-first-deploy))
- Workers: Docker Engine + Compose plugin (`docker compose version`)
- For scrape, alerts, and Observe dashboards: a `prometheus` job in the bucket (see [kive-go `examples/prometheus`](https://github.com/kive-sh/kive/tree/main/examples/prometheus))

## Install into your bucket

From your Kive bucket workspace (adjust the source path if you cloned chaosbox elsewhere):

```bash
cp -R examples/kive/chaosbox workspace/jobs/chaosbox
kive build
kive deploy --jobs chaosbox
```

If the bucket already has Prometheus, redeploy it so scrape configs and rules pick up this job:

```bash
kive deploy --jobs prometheus
```

## Verify

```bash
kive health_check --jobs chaosbox --wait --verbose
kive job status chaosbox
kive job run chaosbox --target test
```

Find the assigned port (`kive cat allocations` or the UI), then:

```bash
export BASE_URL=http://<worker-ip>:<chaosbox_http_port>

curl -sS "${BASE_URL}/"
curl -sS -X POST "${BASE_URL}/count/incr"
curl -sS -X POST "${BASE_URL}/load/cpu/start"
curl -sS -X POST "${BASE_URL}/log/error"
curl -sS "${BASE_URL}/metrics" | grep chaosbox_
```

Open the control console at `${BASE_URL}/ui` or the API browser at `/docs`.
Use `/ui?demo=counter` (or `health`, `load`, `file`) to focus the console.

### Observe

With Prometheus allocated, open **Observe → Dashboards → chaosbox**:

| Plugin | What it shows |
|--------|----------------|
| Health | Scrape `up` and duration |
| Counter | `chaosbox_count` (max across instances) |
| Load | CPU / memory / disk simulators |

Plugin source is catalog-only (`kive build`; `kive push` on a server bucket). Redeploy `prometheus` when scrape or alert files change.

Compose stamps `kive.bucket`, `kive.job`, and `kive.allocation` on the
container and uses the `json-file` log driver so Coroot / OTEL can collect
stdout. After `POST /log/error`, open **Observe → Logs** and filter job
`chaosbox` (ClickHouse `ResourceAttributes['kive.job']`).

## Files

| File | Role |
|------|------|
| `job.conf` | Selectors, memory/CPU, public HTTP port, readiness probe |
| `Makefile` | `start` / `stop` / `restart` / `status` / `logs` / `test` |
| `docker-compose.yml.tpl` | Published image, `chaosbox_http_port`, `json-file` logs, `kive.*` labels |
| `config.env.tpl` | `HTTP_PORT` for `make test` |
| `.dockerignore` | Default ignore list |
| `_prometheus/scrape.yaml` | `/metrics` on `chaosbox_http_port`, `kive_job` label |
| `_prometheus/alerts/alerts.yaml` | Target-down and load-running-long |
| `_prometheus/runbooks/` | Runbooks linked from those alerts |
| `_prometheus/dashboards/` | Observe plugins: `health`, `counter`, `load` (+ `metrics.js` lib) |

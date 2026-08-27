# chaosbox demo curriculum (Kive)

Deploy **once**, then walk through **one feature at a time**. Do not mix counter,
load, and peer demos in the same smoke-test script.

## Prerequisites

Copy and deploy the starter job:

```bash
cp -R examples/kive/chaosbox workspace/jobs/chaosbox
kive build
kive deploy --jobs chaosbox
kive health_check --jobs chaosbox --wait --verbose
```

Set `BASE_URL` for the walkthrough scripts (local or Kive):

```bash
export BASE_URL=http://<worker-ip>:<chaosbox_http_port>
# local: export BASE_URL=http://localhost:8080
```

## Walkthroughs (single job)

Run in order. Each step exercises **one** capability.

| Step | Feature | Doc | Script |
|------|---------|-----|--------|
| 1 | Health & readiness | [demos/01-health.md](demos/01-health.md) | [demos/01-health.sh](demos/01-health.sh) |
| 2 | Counter (in-memory) | [demos/02-counter.md](demos/02-counter.md) | [demos/02-counter.sh](demos/02-counter.sh) |
| 3 | Load (CPU only) | [demos/03-load-cpu.md](demos/03-load-cpu.md) | [demos/03-load-cpu.sh](demos/03-load-cpu.sh) |
| 4 | File cat | [demos/04-cat-file.md](demos/04-cat-file.md) | [demos/04-cat-file.sh](demos/04-cat-file.sh) |
| 5 | Observe error logs | [demos/05-observe-logs.md](demos/05-observe-logs.md) | [demos/05-observe-logs.sh](demos/05-observe-logs.sh) |

See [demos/README.md](demos/README.md) for the single-deploy model and UI focus URLs (`/ui?demo=counter`, etc.).

**Rule:** finish a load demo with `03-load-cpu.sh` (it stops CPU load) before moving on.

## Advanced (multi-job)

Use these when a feature needs extra infrastructure. Each advanced README demos **one** concept only.

| Stack | Feature | README |
|-------|---------|--------|
| [advanced/counter-redis/](advanced/counter-redis/) | Shared Redis counter | [advanced/counter-redis/README.md](advanced/counter-redis/README.md) |
| [advanced/peers/](advanced/peers/) | Membership & load fan-out | [advanced/peers/README.md](advanced/peers/README.md) |
| [advanced/metrics/](advanced/metrics/) | Prometheus scrape | [advanced/metrics/README.md](advanced/metrics/README.md) |

For a local all-in-one stack (Redis + two peered nodes), see the repo-root
[`docker-compose.yml`](../../docker-compose.yml).

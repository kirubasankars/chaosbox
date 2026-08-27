# chaosbox

A small HTTP service used for infra demos and testing: health/version,
file cat, a counter (in-memory or Redis), CPU/memory/disk load simulators,
peer membership tracking, and Prometheus metrics.

## Quick start

No config file or flags required — built-in defaults listen on `:8080`, log to
stdout, and use a system temp directory for disk-load scratch files.

```bash
make run
# or:
go build -o chaosbox ./cmd/chaosbox && ./chaosbox
```

```bash
docker pull ghcr.io/kive-sh/chaosbox:latest
docker run --rm -p 8080:8080 ghcr.io/kive-sh/chaosbox:latest
```

Then open [http://localhost:8080/ui](http://localhost:8080/ui) for the control
console, or [http://localhost:8080/docs](http://localhost:8080/docs) for the
API browser.

```bash
curl http://localhost:8080/
```

## Demo curriculum

Walk through **one feature at a time** — deploy once, then follow the scripts
in order:

| Step | Feature |
|------|---------|
| 1 | [Health](examples/kive/demos/01-health.md) |
| 2 | [Counter](examples/kive/demos/02-counter.md) |
| 3 | [Load (CPU)](examples/kive/demos/03-load-cpu.md) |
| 4 | [File cat](examples/kive/demos/04-cat-file.md) |
| 5 | [Observe logs](examples/kive/demos/05-observe-logs.md) |

Full index: [`examples/kive/README.md`](examples/kive/README.md). Use
`/ui?demo=counter` (or `health`, `load`, `file`) to focus the console during
live demos.

## Deploy on Kive

A starter single-node job lives under [`examples/kive/`](examples/kive/):

```bash
cp -R examples/kive/chaosbox workspace/jobs/chaosbox
kive build
kive deploy --jobs chaosbox
kive health_check --jobs chaosbox --wait --verbose
```

Then run the [demo walkthroughs](examples/kive/demos/) one at a time. Advanced
multi-job stacks (Redis counter, peers, metrics) live under
[`examples/kive/advanced/`](examples/kive/advanced/).

## Layout

```
cmd/chaosbox/            entrypoint: flag parsing and wiring
internal/
  config/            config.json loading + validation (optional file)
  logging/           structured JSON slog setup (stdout + file, OTEL-friendly)
  metrics/           Prometheus collectors
  httpx/             shared HTTP helpers (JSON writers, query parsing, request logging)
  api/                health (/) and file cat (/_cat/file) handlers
  counter/           /count/incr and /count/decr, memory or Redis backend
  loadsim/           /load/{cpu,mem,disk,all}/{start,stop} pulsing load simulators
  membership/        peer health tracking + /_cat/nodes
  docs/              OpenAPI spec + Swagger UI (/docs)
  ui/                single-page control console (/ui)
examples/kive/       Kive jobs + demo curriculum (demos/, advanced/)
```

## Build & run

### Makefile

```bash
make build   # compile to bin/chaosbox
make run     # build + run with built-in defaults
make test    # go test ./...
make vet     # go vet ./...
make fmt     # gofmt -l -w .
make tidy    # go mod tidy
make clean   # remove bin/, data/, logs/
make help    # list all targets
```

`make run` accepts optional overrides, e.g.
`make run CONFIG=other.json REDIS=redis://localhost:6379/0`.
`make docker-build` / `make docker-run` build and run the image with no
volume mounts required.

### Tests

`make test` (`go test ./...`) includes integration-style tests that exercise
real network/Redis code paths, not just unit-level logic:

- `internal/counter` — runs the Redis-backed counter against a real Redis
  wire protocol server via an in-process [miniredis](https://github.com/alicebob/miniredis)
  instance (no external Redis required); set `CHAOSBOX_TEST_REDIS_DSN` to point
  at a real Redis instance instead (e.g. in CI).
- `internal/membership` — spins up real `httptest` HTTP servers to act as
  peer nodes, verifying peer up/down/version/load probing, and verifying
  that `/load/*`-style fan-out reaches every peer exactly once with no
  cascading relay even when peers list each other mutually.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | *(built-in defaults)* | path to config JSON; omit for `:8080`, no peers |
| `-file` | `<data>/file.txt` | plain text file served by `/_cat/file` |
| `-data` | *(system temp)* | folder used by the disk load simulator |
| `-logs` | *(stdout only)* | folder for `chaosbox.log`; omit to skip file logging |
| `-redis` | *(in-memory)* | Redis DSN for a shared counter backend |
| `-startup-delay` | `0` | seconds to sleep before listening |

Add `-config` when you need peers, TLS, or a custom listen address. Add
`-redis` when multiple nodes should share the same counter. Pass `-data` and/or
`-logs` when you want persistent paths instead of temp data and stdout-only logs.

## Docker

Images are published to GitHub Container Registry on pushes to `main` and on
`v*` tags:

```bash
docker pull ghcr.io/kive-sh/chaosbox:latest
# or a release / commit tag, e.g. ghcr.io/kive-sh/chaosbox:1.0.0
```

If the package is private, authenticate first:

```bash
echo "$GITHUB_TOKEN" | docker login ghcr.io -u USERNAME --password-stdin
```

Build and run locally (no volumes or args required):

```bash
docker build -t chaosbox .
docker run --rm -p 8080:8080 chaosbox
```

To override defaults, pass flags on the command, e.g.
`docker run --rm -p 8080:8080 chaosbox -redis redis://redis:6379/0`.
Mount a `config.json` (and `peer_ca_cert` PEM, if used) when configuring
peers or TLS.

### Local full stack (Docker Compose)

For local development, `docker-compose.yml` runs Redis plus two peered nodes
(`chaosbox-a`, `chaosbox-b`) in one stack. For Kive-native equivalents split
by feature, see [`examples/kive/advanced/`](examples/kive/advanced/).

```bash
docker compose up --build -d   # or: make compose-up
```

| Service | Host port | Container port |
|---------|-----------|-----------------|
| `chaosbox-a` | 8081 | 8080 |
| `chaosbox-b` | 8082 | 8080 |
| `redis` | 6379 | 6379 |

Demo one feature at a time even on this stack — e.g. membership only:

```bash
curl http://localhost:8081/_cat/nodes
curl -X POST http://localhost:8081/load/cpu/start
curl http://localhost:8081/_cat/nodes
curl -X POST http://localhost:8081/load/cpu/stop
```

Counter-with-Redis and other advanced flows: [`examples/kive/advanced/counter-redis/`](examples/kive/advanced/counter-redis/).

`docker compose down` (or `make compose-down`) stops the stack; add
`rm-volumes=1` to `make compose-down` (or `-v` to `docker compose down`) to
also drop the `chaosbox-a`/`chaosbox-b` data and log volumes.

### config.json (optional)

When omitted, chaosbox uses built-in defaults (`version` `0.1.0`, `listen`
`:8080`, no peers). To customize, pass `-config path/to/config.json`:

```json
{
  "version": "0.1.0",
  "listen": ":8080",
  "tls_cert": "",
  "tls_key": "",
  "peers": [],
  "peer_check_sec": 5,
  "peer_ca_cert": ""
}
```

Set `tls_cert`/`tls_key` to serve HTTPS on `listen` instead of HTTP. `peers`
lists other chaosbox instances (host:port, or a full `http://`/`https://` URL);
the node's own address is derived from `listen`. When peers are addressed
with `https://` and use certs signed by a private CA, set `peer_ca_cert` to
a PEM file so their certs verify against that CA instead of the system trust
store.

## API

| Method | Path | Description |
|--------|------|--------------|
| `GET` | `/` | health + version + running load simulators (pretty JSON) |
| `GET` | `/_cat/file` | contents of `-file` |
| `GET` | `/_cat/nodes` | membership table incl. per-node load state (plain text) |
| `POST` | `/count/incr` / `/count/decr` | counter ops |
| `POST` | `/log/error` | emit error-level slog + HTTP 500 (Observe feature test) |
| `POST` | `/load/cpu/start` / `/stop` | pulsing CPU load (`min_pct`, `max_pct`, `period`) |
| `POST` | `/load/mem/start` / `/stop` | pulsing memory load (`mb`, `period`) |
| `POST` | `/load/disk/start` / `/stop` | pulsing disk IO load (`mb`, `period`) |
| `POST` | `/load/all/start` / `/stop` | all three together |
| `GET` | `/metrics` | Prometheus exposition |
| `GET` | `/docs` | Swagger UI for browsing/testing the API |
| `GET` | `/docs/openapi.yaml` | raw OpenAPI 3.0 spec |
| `GET` | `/ui` | single-page control console (status + API triggers) |
| `GET` | `/ui/nodes` | membership snapshot as JSON (used by `/ui`) |

`/` includes a `load` object reporting which simulators are currently
running on this node, e.g. `"load": {"cpu": true, "memory": false, "disk": false}`.
`/_cat/nodes` shows the same info per node as a `load` column (e.g.
`cpu,disk` or `-` when idle or unknown), self computed live and peers from
their last successful health check.

Open `/ui` for a single-page control console that polls health and peer
status and can trigger counter/load APIs. Open `/docs` for the full
Swagger UI (backed by the `swagger-ui-dist` CDN bundle, spec served from
`/docs/openapi.yaml`).

Calling any `/load/*` endpoint on a node also triggers the same request
(method + query) on that node's known peers, best-effort and in the
background. The local response reflects only the local result. Fan-out is
single-hop: a request relayed from a peer is executed locally but not
relayed again, so it's safe even when peer lists form a full mesh.

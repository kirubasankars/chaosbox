# chaosbox

A small HTTP service used for infra demos and testing: health/version,
file cat, a counter (in-memory or Redis), CPU/memory/disk load simulators,
peer membership tracking, and Prometheus metrics.

## Layout

```
cmd/chaosbox/            entrypoint: flag parsing and wiring
internal/
  config/            config.json loading + validation
  logging/           structured JSON slog setup (stdout + file, OTEL-friendly)
  metrics/           Prometheus collectors
  httpx/             shared HTTP helpers (JSON writers, query parsing, request logging)
  api/                health (/) and file cat (/_cat/file) handlers
  counter/           /count/incr and /count/decr, memory or Redis backend
  loadsim/           /load/{cpu,mem,disk,all}/{start,stop} pulsing load simulators
  membership/        peer health tracking + /_cat/nodes
  docs/              OpenAPI spec + Swagger UI (/docs)
  ui/                single-page control console (/ui)
```

## Build & run

```bash
go build -o chaosbox ./cmd/chaosbox
./chaosbox -config config.json -file data.txt -data ./data -logs ./logs
```

### Makefile

```bash
make build   # compile to bin/chaosbox
make run     # build + run with local defaults (config.json, data.txt, ./data, ./logs)
make test    # go test ./...
make vet     # go vet ./...
make fmt     # gofmt -l -w .
make tidy    # go mod tidy
make clean   # remove bin/, data/, logs/
make help    # list all targets
```

`make run` accepts overrides, e.g. `make run CONFIG=other.json REDIS=redis://localhost:6379/0`.
`make docker-build` / `make docker-run` build and run the image described below.

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

| Flag | Description |
|------|-------------|
| `-config` | path to config JSON (required) |
| `-file` | plain text file served by `/_cat/file` (required) |
| `-data` | folder used by the disk load simulator (required) |
| `-logs` | folder for `chaosbox.log` (required) |
| `-redis` | Redis DSN for the counter backend; omit for in-memory |
| `-startup-delay` | seconds to sleep before listening |

## Docker

```bash
docker build -t chaosbox .
docker run --rm -p 8080:8080 \
  -v "$PWD/config.json:/app/config.json:ro" \
  -v chaosbox-data:/app/data \
  -v chaosbox-logs:/app/logs \
  chaosbox
```

The image bakes in default flags (`-config /app/config.json -file
/app/data/file.txt -data /app/data -logs /app/logs`); override the command
to change them, e.g. add `-redis redis://redis:6379/0` for the Redis counter
backend. Mount your own `config.json` (and `peer_ca_cert` PEM, if used) into
the container to configure peers/TLS.

### Docker Compose

`docker-compose.yml` runs a small demo chaosbox cluster: a `redis` service (shared
counter backend) plus two peered nodes, `chaosbox-a` and `chaosbox-b`, each
configured with the other as its peer (`docker/chaosbox-a.config.json`,
`docker/chaosbox-b.config.json`).

```bash
docker compose up --build -d   # or: make compose-up
```

| Service | Host port | Container port |
|---------|-----------|-----------------|
| `chaosbox-a` | 8081 | 8080 |
| `chaosbox-b` | 8082 | 8080 |
| `redis` | 6379 | 6379 |

```bash
curl http://localhost:8081/_cat/nodes   # shows chaosbox-a (self) and chaosbox-b, reachable via the chaosbox-b hostname
curl -X POST http://localhost:8081/count/incr   # incr on chaosbox-a...
curl -X POST http://localhost:8082/count/incr   # ...and chaosbox-b share the same Redis-backed counter
curl -X POST http://localhost:8081/load/cpu/start   # fans out to chaosbox-b too; check /_cat/nodes on either node
```

`docker compose down` (or `make compose-down`) stops the stack; add
`rm-volumes=1` to `make compose-down` (or `-v` to `docker compose down`) to
also drop the `chaosbox-a`/`chaosbox-b` data and log volumes.

### config.json

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

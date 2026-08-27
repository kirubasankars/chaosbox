# Demo 1: Health & readiness

**Feature:** HTTP health, version, idle load state.

**Endpoints:** `GET /`

## Run

```bash
export BASE_URL=http://<worker>:<port>
./examples/kive/demos/01-health.sh
```

Or manually:

```bash
curl -s "${BASE_URL}/" | jq .
```

Expect `"status": "ok"`, a `version` string, and `"load": {"cpu": false, ...}`.

## Kive

`kive health_check --jobs chaosbox --wait` exercises the same readiness probe
defined in `job.conf`.

## UI

Open `${BASE_URL}/ui?demo=health` — version and load chips only.

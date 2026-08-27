# chaosbox (Kive job)

Single-node chaosbox workload for Kive demos. Uses the published
`ghcr.io/kive-sh/chaosbox:latest` image with no config file or command
override.

## Files

| File | Purpose |
|------|---------|
| `job.conf` | Resources, port pool key, HTTP readiness probe |
| `docker-compose.yml.tpl` | Container image, port bind, data/log volumes |
| `Makefile` | `start` / `stop` / `restart` / `status` / `logs` |
| `config.json.tpl` | Optional reference for peers/TLS when you outgrow defaults |

## Deploy

```bash
cp -R examples/kive/chaosbox workspace/jobs/chaosbox
kive build
kive deploy --jobs chaosbox
kive health_check --jobs chaosbox --wait --verbose
```

## Smoke test

```bash
curl http://<worker>:<port>/
curl -X POST http://<worker>:<port>/count/incr
open http://<worker>:<port>/ui
```

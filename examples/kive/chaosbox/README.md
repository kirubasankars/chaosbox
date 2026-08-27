# chaosbox (Kive job)

Single-node starter job. Deploy once, then follow the [demo curriculum](../README.md).

## Files

| File | Purpose |
|------|---------|
| `job.conf` | Resources, port pool key, HTTP readiness probe |
| `docker-compose.yml.tpl` | Container image and port bind |
| `Makefile` | `start` / `stop` / `restart` / `status` / `logs` |
| `config.json.tpl` | Optional reference for peers/TLS |

## Deploy

```bash
cp -R examples/kive/chaosbox workspace/jobs/chaosbox
kive build
kive deploy --jobs chaosbox
kive health_check --jobs chaosbox --wait --verbose
```

## Smoke test (health only)

```bash
curl http://<worker>:<port>/
open http://<worker>:<port>/ui?demo=health
```

Counter, load, file, and Observe demos: [../demos/](../demos/).

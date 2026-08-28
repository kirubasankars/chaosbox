# Chaosbox target down

## Alert

- `chaosbox target down` — scrape (`up{kive_job=...}`) failing for 5 minutes

## Diagnosis

```bash
docker ps --filter name=chaosbox
docker logs "$(docker ps -q --filter name=chaosbox)" --tail 80
PORT=$(kive cat kv get kive/bucket chaosbox_http_port)
curl -sS "http://127.0.0.1:${PORT}/"
curl -sS "http://127.0.0.1:${PORT}/metrics" | head
```

Confirm the job is allocated on the expected worker and that `chaosbox_http_port` matches the published container port.

## Remediation

1. Restart the job: `kive job run chaosbox --target restart`
2. Verify scrape config after deploy: redeploy prometheus so job scrapes regenerate
3. Check Observe → Dashboards → chaosbox → Health for which instances are down

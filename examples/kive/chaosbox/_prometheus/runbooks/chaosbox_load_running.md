# Chaosbox load running long

## Alert

- `chaosbox load running long` — `chaosbox_load_running{kind=...}` stayed at `1` for 30 minutes

Intentional for Kive load / alert testing. Fire the alert on purpose, then stop the simulator when finished.

## Diagnosis

```bash
PORT=$(kive cat kv get kive/bucket chaosbox_http_port)
curl -sS "http://127.0.0.1:${PORT}/"
curl -sS "http://127.0.0.1:${PORT}/_cat/nodes"
curl -sS "http://127.0.0.1:${PORT}/metrics" | grep chaosbox_load
```

Open `/ui` on the public HTTP port to see which simulators are active locally and on peers.

## Remediation

Stop the kind that is still running (or all):

```bash
PORT=$(kive cat kv get kive/bucket chaosbox_http_port)
curl -sS -X POST "http://127.0.0.1:${PORT}/load/cpu/stop"
curl -sS -X POST "http://127.0.0.1:${PORT}/load/mem/stop"
curl -sS -X POST "http://127.0.0.1:${PORT}/load/disk/stop"
# or:
curl -sS -X POST "http://127.0.0.1:${PORT}/load/all/stop"
```

Confirm in Observe → Dashboards → chaosbox → Load that running kinds return to `0`. Peer fan-out means stopping on one node also stops known peers.

# Demo 5: Observe error logs

**Feature:** emit a deliberate error-level log line for Observe / OTEL pipelines.

**Endpoints:** `POST /log/error` (returns HTTP 500)

## Run

```bash
./examples/kive/demos/05-observe-logs.sh
```

Expect status **500** and JSON `{"status":"error","msg":"log.error_emitted"}`.
On the worker, structured logs include `"msg":"log.error_emitted"`.

## Kive Observe

Point Observe Logs at the worker log source for the `chaosbox` job and filter
for `log.error_emitted` after running this demo.

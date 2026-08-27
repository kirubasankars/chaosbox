# Advanced: shared Redis counter

Demo **only** the shared counter — deploy Redis first, then chaosbox with
`-redis`. Do not enable peers or load fan-out in this stack.

## Jobs

| Job | Directory | Purpose |
|-----|-----------|---------|
| `chaosbox_redis` | [`redis/`](redis/) | Redis 7 on a bucket port |
| `chaosbox` | [`chaosbox/`](chaosbox/) | chaosbox using Redis backend |

Both jobs should allocate to the **same worker** so chaosbox can reach Redis
via the host-published port.

## Install

```bash
cp -R examples/kive/advanced/counter-redis/redis workspace/jobs/chaosbox_redis
cp -R examples/kive/advanced/counter-redis/chaosbox workspace/jobs/chaosbox
kive build
kive deploy --jobs chaosbox_redis
kive deploy --jobs chaosbox
```

## Demo (counter only)

```bash
export BASE_URL=http://<worker>:<chaosbox_http_port>

curl -s -X POST "${BASE_URL}/count/incr"
curl -s -X POST "${BASE_URL}/count/incr"
curl -s "${BASE_URL}/count"
# expect {"count":2}
```

Restart the chaosbox container and `GET /count` should still show `2` — state
lives in Redis, not the process.

## How it connects

The chaosbox Compose template uses `host.docker.internal` and the bucket
`chaosbox_redis_port` to reach the Redis job's published port on the worker
host.

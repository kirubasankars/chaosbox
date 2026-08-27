# Advanced: peer membership & fan-out

Demo **only** cluster membership and load fan-out — no Redis. Deploy both peer
jobs to the **same worker**.

## Jobs

| Job | Directory | Peer config |
|-----|-----------|-------------|
| `chaosbox_a` | [`chaosbox_a/`](chaosbox_a/) | lists `chaosbox_b:8080` |
| `chaosbox_b` | [`chaosbox_b/`](chaosbox_b/) | lists `chaosbox_a:8080` |

Peer hostnames match Compose service names. On a single worker, attach both
jobs to a shared Docker network named `chaosbox_peers` (see each job's Compose
template).

## Install

```bash
cp -R examples/kive/advanced/peers/chaosbox_a workspace/jobs/chaosbox_a
cp -R examples/kive/advanced/peers/chaosbox_b workspace/jobs/chaosbox_b
kive build
kive deploy --jobs chaosbox_a,chaosbox_b
```

## Demo (membership + one fan-out)

```bash
export BASE_A=http://<worker>:<chaosbox_a_http_port>

# 1. Membership table
curl -s "${BASE_A}/_cat/nodes"

# 2. CPU load on A fans out to B (single hop)
curl -s -X POST "${BASE_A}/load/cpu/start"
sleep 2
curl -s "${BASE_A}/_cat/nodes"   # both nodes should show cpu in load column

curl -s -X POST "${BASE_A}/load/cpu/stop"
```

Open `${BASE_A}/ui` — the **Cluster nodes** panel shows peer status.

## Local equivalent

Repo-root [`docker-compose.yml`](../../../docker-compose.yml) runs the same
topology with Redis omitted from the peer demo path (compose stack includes
Redis for counter sharing — use peer jobs here for membership-only demos).

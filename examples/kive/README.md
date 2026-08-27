# chaosbox on Kive

Starter job for deploying chaosbox as a single-node Kive workload. No
`config.json` mount is required — the image runs with built-in defaults.

## Install

From your Kive bucket workspace:

```bash
cp -R examples/kive/chaosbox workspace/jobs/chaosbox
kive build
kive deploy --jobs chaosbox
kive health_check --jobs chaosbox --wait --verbose
```

Replace `examples/kive/chaosbox` with the path to this directory if you
cloned chaosbox elsewhere.

## Try it

After deploy, find the assigned port (from `kive cat allocations` or the UI),
then:

```bash
curl http://<worker-ip>:<chaosbox_http_port>/
curl -X POST http://<worker-ip>:<chaosbox_http_port>/count/incr
curl -X POST http://<worker-ip>:<chaosbox_http_port>/load/cpu/start
```

Open the control console at `http://<worker-ip>:<chaosbox_http_port>/ui` or
browse the API at `/docs`.

## Optional config

The starter job uses image defaults (in-memory counter, no peers). To add
peers, TLS, or a custom listen address later, render
[`config.json.tpl`](chaosbox/config.json.tpl) at deploy time and mount it
into the container at `/app/config.json`, then pass `-config /app/config.json`
in the Compose `command`.

For a multi-node demo with Redis and peer fan-out, see the repo's
[`docker-compose.yml`](../../docker-compose.yml) (advanced).

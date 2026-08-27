# Demo 2: Counter (in-memory)

**Feature:** increment/decrement a local counter. No Redis, no peers.

**Endpoints:** `GET /count`, `POST /count/incr`, `POST /count/decr`

## Run

```bash
./examples/kive/demos/02-counter.sh
```

## UI

Open `${BASE_URL}/ui?demo=counter` and use **+ incr** / **− decr**.

## Notes

- Do not call `/load/*` in this demo.
- For a shared counter across nodes, see [../advanced/counter-redis/](../advanced/counter-redis/).

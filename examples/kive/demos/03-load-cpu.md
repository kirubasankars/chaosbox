# Demo 3: Load (CPU only)

**Feature:** pulsing CPU load simulator. Memory, disk, and `/load/all` are out of scope for this step.

**Endpoints:** `POST /load/cpu/start`, `POST /load/cpu/stop`, `GET /` (poll `load.cpu`)

## Run

```bash
./examples/kive/demos/03-load-cpu.sh
```

The script starts CPU load, waits until `load.cpu` is true, then stops and confirms idle.

## UI

Open `${BASE_URL}/ui?demo=load` — use the **CPU** card only.

## Variations (same job, separate sessions)

- Memory: `POST /load/mem/start` then `/load/mem/stop`
- Disk: `POST /load/disk/start` then `/load/disk/stop`

Run one simulator at a time. For peer fan-out, see [../advanced/peers/](../advanced/peers/).

# Single-deploy demo walkthroughs

These scripts run against **one** deployed `chaosbox` job. Install the job once
(see [../README.md](../README.md)), then run each demo in order.

## Usage

```bash
export BASE_URL=http://localhost:8080   # or your Kive worker URL
./examples/kive/demos/01-health.sh
./examples/kive/demos/02-counter.sh
# ...
```

Scripts exit non-zero on failure. They use only `curl` and standard shell tools.

## UI focus mode

During live demos, open a focused console so only one panel is visible:

| Demo step | URL |
|-----------|-----|
| Health | `/ui?demo=health` |
| Counter | `/ui?demo=counter` |
| Load | `/ui?demo=load` |
| File cat | `/ui?demo=file` |

Omit `?demo=` to show the full console.

## What not to do

- Do not run counter and load curls in the same “hello world” script.
- Do not start `/load/all` in the CPU demo — use CPU only.
- Stop load simulators before switching to another load type.

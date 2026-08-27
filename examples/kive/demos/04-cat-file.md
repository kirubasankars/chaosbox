# Demo 4: File cat

**Feature:** serve plain text from the `-file` path via `/_cat/file`.

**Endpoints:** `GET /_cat/file`

## Run

```bash
./examples/kive/demos/04-cat-file.sh
curl -s "${BASE_URL}/_cat/file"
```

## UI

Open `${BASE_URL}/ui?demo=file`.

## Notes

The starter job uses the image default (`<data>/file.txt`, often empty). Pass
`-file` on the container command to serve custom content.

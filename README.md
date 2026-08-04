# Orthograph

Pencil-first collaborative sketch surface for technical diagrams — shared
between a human on iPad and coding agents on the Mac.

```
  Agent ──MCP──►  orthograph (daemon)  ◄──pigeon──►  iPad Orthograph.app
                       │
                       └── vector document + op log
                           (optional PNG render)
```

- **Vector-first observability** — agents read a stable scene graph
  (`orthograph_scene`); pixels on request (`orthograph_render`).
- **Same-surface collab** — human Pencil ink and agent primitives share
  one document; optional ghost/propose mode for agent edits.
- **Standalone product** — not a Jevons feature. Jevons (and any MCP
  client) integrates via tools; stopping either product leaves the other
  working. See [docs/design.md](docs/design.md) §15.

## Status

Scaffold. Phase 0: daemon + MCP scene create/read on Mac. iPad + pigeon
and shape recognition follow. Design: [docs/design.md](docs/design.md).

## Install (development)

```bash
git clone https://github.com/marcelocantos/orthograph.git
cd orthograph
make bullseye          # lint + vet + test + build → bin/orthograph
./bin/orthograph --help
```

Homebrew formula and `brew services` arrive with the first release.

## CLI (planned shape)

| Command | Purpose |
|---------|---------|
| `orthograph daemon` | Run the always-on Mac daemon (MCP + pigeon) |
| `orthograph pair` | Print QR / instance id for iPad pairing |
| `orthograph status` | Paired clients, active document |
| `orthograph open` | Create or select a document |
| `orthograph export` | SVG / PNG / PDF / JSON to a host path |

Binary supports `--version`, `--help`, and `--help-agent`.

## MCP (planned)

Server name: **`orthograph`**. Core tools (see design for full set):

| Tool | Role |
|------|------|
| `orthograph_status` | Paired?, active doc, presence |
| `orthograph_scene` | Vector scene (primary observe) |
| `orthograph_render` | PNG/WebP snapshot |
| `orthograph_create` / `update` / `delete` | Mutate objects |
| `orthograph_push_image` | Host file or clipboard → canvas |

## Data

Runtime state lives under `~/.orthograph/` (pairings, documents, media).
Never commit that directory.

## License

Apache License 2.0. Copyright 2026 Marcelo Cantos.

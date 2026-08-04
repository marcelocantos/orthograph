# Orthograph — agent guide

Short reference for agents operating this product once the daemon is
running. Product architecture: [docs/design.md](docs/design.md).
Repo conventions: [AGENTS.md](AGENTS.md).

## Install / run (development)

```bash
cd ~/work/github.com/marcelocantos/orthograph
make bullseye
./bin/orthograph --help-agent
# later:
# ./bin/orthograph daemon
# claude mcp add --scope user --transport http orthograph http://localhost:<port>/mcp
```

## Observe first

1. `orthograph_status` — is a pad paired? which doc is active?
2. `orthograph_scene` — vector graph (preferred over pixels).
3. `orthograph_render` — only when layout, handwriting, or images matter.
4. Prefer `orthograph_summary` (when available) for large documents.

## Collaborate

- Create scaffolding with `orthograph_create` (rects, connectors, text).
- Leave human ink alone mid-stroke; don't thrash shared layers.
- Respect propose/ghost mode when enabled — wait for human accept.
- Never wipe a document without an explicit confirm flag and human-visible toast.

## Push media

- `orthograph_push_image` with a host path, or clipboard when supported.
- Spyder screenshots are first-class underlays for UI markup.

## Boundary

Orthograph owns the canvas. Jevons (or any overseer) only directs *when*
to open or annotate a diagram. Do not store scene graphs in other
products' state directories.

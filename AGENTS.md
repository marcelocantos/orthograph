# Orthograph — agent instructions

Canonical agent-facing instructions for this repo. Tool adapters
(`CLAUDE.md`, etc.) import or mirror this file.

## What this is

**Orthograph** is a standalone collaborative sketch product:

- Mac daemon + CLI (`orthograph`)
- HTTP MCP server (`orthograph_*` tools)
- iPad app (Pencil) over [pigeon](https://github.com/marcelocantos/pigeon)

It is **not** a Jevons subsystem. Jevons may integrate as a consumer;
document truth lives in orthograph (`~/.orthograph/`). Full design:
[docs/design.md](docs/design.md).

## Layout

| Path | Role |
|------|------|
| `main.go` | CLI entry (`--version`, `--help`, `--help-agent`, subcommands) |
| `internal/doc` | Document model, op log, persistence |
| `internal/mcp` | MCP tool surface |
| `internal/render` | Vector → PNG/SVG export |
| `internal/protocol` | Wire protocol (ops, media, presence) for pigeon |
| `ios/` | iPad app (later) |
| `docs/design.md` | Product architecture |

## Build & gates

```bash
make bullseye   # lint + vet + test + build
```

`make bullseye` is the durable green signal. Do not pass `-j` to make.

## Delivery

- Default branch: `master`
- **Build plane:** commit freely locally
- **Ship plane:** push/PR only when explicitly asked (`/push`, “ship”, …)
- Squash-only merges via PR when shipping

## Conventions

- Apache-2.0; SPDX headers on source when adding files that need them
- No new TOML config — JSON, YAML, or plain text
- Standalone binaries: `--version`, `--help`, `--help-agent`
- Language rules: read `~/.claude/go.md` before substantial Go;
  `~/.claude/mobile-development.md` before iOS/device work
- Prefer spyder for device deploy/smoke on the lab iPad

## Targets

Bullseye owns followable work (`bullseye.yaml`). Use bullseye MCP tools;
do not hand-edit the file unless repairing unloadable state.

# Orthograph — Design

**Status:** adopted into repo (names locked); design still evolving  
**Date:** 2026-08-05  
**Working name:** *Orthograph* (daemon + CLI `orthograph`, MCP server `orthograph`, app *Orthograph*)  
**Primary device:** iPad + Apple Pencil  
**Host:** Mac daemon (brew service), agent-visible via MCP  

---

## 1. Problem

Agents can read code, logs, and screenshots, but they cannot **share a drawing surface** with a human in real time. Technical discussion still jumps to ASCII, PlantUML, Excalidraw links, or “describe what you mean.” That loses spatial intent, forces re-transcription, and makes collaborative diagramming awkward.

**Orthograph** is a Pencil-first iPad sketch surface whose document is:

1. **Live on the Mac** (source of truth in a local daemon).
2. **Observable by agents** as vector scene data (preferred) and pixels.
3. **Writable by agents** on the same surface (or a new diagram they open with you).
4. **Carried to the iPad** over **pigeon** (E2E encrypted session; same pairing shape as Jevons).

---

## 2. Goals and non-goals

### Goals

| Goal | Notes |
|------|--------|
| Pencil-native ink | Pressure, tilt, predicted touches, palm rejection, hover if available |
| Technical-diagram bias | Shape recognition, grid/object snap, connectors, layers — not a full Procreate clone |
| Agent-first observability | Stable IDs, vector scene graph, cheap summaries, optional raster |
| Same-surface collab | Human + agent(s) concurrent; presence and authorship visible |
| Push media to pad | Files, clipboard images, spyder screenshots as canvas objects |
| Offline-capable pad | Local buffer while pigeon drops; resync on reconnect |
| Reuse stack | pigeon (transport), brew launchd daemon pattern (spyder/blurter/jevons), HTTP MCP |

### Non-goals (v1)

- Full illustration suite (brushes, layer blend modes, CMYK, print pipeline).
- Multi-human real-time rooms across the internet (one Mac + one or few paired pads).
- Perfect CAD / parametric constraint solver (snapping + recognition is enough).
- Replacing rustuml/PlantUML for *text-defined* diagrams (complement, not replace).
- iPhone as primary surface (phone can view later; design for iPad).

---

## 3. Architecture

```
                    ┌─────────────────────────────────────┐
  Agent tools       │  orthographd  (Mac daemon)              │
  ──────────MCP────►│  • document store (SQLite + files)  │
                    │  • op log / CRDT merge               │
                    │  • render service (vector → PNG)    │
                    │  • pigeon backend (register/listen) │
                    │  • HTTP MCP :port/mcp               │
                    └───┬───────────────▲─────────────────┘
                        │ streams       │ streams
                        │ + datagrams   │
                    ┌───▼───────────────┴─────────────────┐
                    │  Orthograph.app  (iPad, SwiftUI)     │
                    │  • Pencil input + local ink preview │
                    │  • scene apply + presence           │
                    │  • QR pairing (pigeon)              │
                    └─────────────────────────────────────┘
```

### Why the Mac owns the document

- Agents already run on the Mac; MCP is local and fast.
- One writer-of-record for persistence, export, and multi-client fan-out.
- iPad stays a high-quality **viewport + input device**, not a second database to reconcile with agent tool results.
- Matches Jevons: device is a peer over pigeon; truth lives in the daemon.

### Transport (pigeon)

Reuse pigeon as-is:

| Channel | Use |
|---------|-----|
| Reliable named stream `control` | Pairing status, open/close doc, tool modes, undo, config |
| Reliable named stream `ops` | Document operations (JSON or CBOR frames) — ordered |
| Reliable named stream `media` | Chunked image/PDF payloads (or file-id + side channel) |
| Unreliable datagram `ink` | Live freehand samples for low-latency remote cursors / in-progress strokes |
| Unreliable datagram `presence` | Viewport, pointer, “agent thinking” badge |

Pairing: QR from `orthographd pair` (instance ID + relay), scan on iPad, 6-digit confirm — same ceremony as other pigeon apps.

LAN upgrade when available (pigeon path switch) keeps stroke latency low at home.

### Process layout on Mac

```
orthographd
  ├── MCP HTTP  (e.g. :13720/mcp)     — agents
  ├── optional local WS/HTTP preview  — Mac browser debug view
  ├── pigeon Register/Accept          — iPad
  └── on-disk:
        ~/.orthograph/
          config.yaml                 # only if needed; prefer JSON/YAML per conventions
          pairings/
          docs/<doc-id>/
            meta.json
            scene.sqlite              # or scene.jsonl + snapshots
            media/<asset-id>
            exports/
```

Single binary + brew service, spyder-style. MCP and pigeon share process memory (no second hop for agent→pad).

---

## 4. Document model (vector-first)

### Scene graph

```
Document
  id, title, created, revised, page_size? (infinite canvas default)
  layers[]
    id, name, visible, locked, z
    objects[]
      id, type, author, created_ms, props...
  assets[]     # images, PDF page snaps
  guides[]     # user + auto alignment guides
  viewport     # last human viewport (hint only)
```

### Object types (v1 set)

| Type | Fields (sketch) | Role |
|------|-----------------|------|
| `stroke` | points[{x,y,p,tilt?,t}], style, recognized_as? | Freehand / ink |
| `path` | SVG-like path / poly | Cleaned vector after recognition |
| `rect` / `ellipse` / `polygon` | geom + style | Recognized / agent-placed |
| `line` / `arrow` | endpoints, heads | |
| `connector` | from_id+port, to_id+port, route | Orthogonal / curved |
| `text` | string, box, font, align | Labels |
| `image` | asset_id, rect, opacity | Pushed screenshots etc. |
| `group` | child_ids | |
| `note` | text, pin | Sticky / agent annotation (distinct chrome) |
| `frame` | rect, label | Named region for agent focus (“API boundary”) |

Every object has:

- stable `id` (ULID)
- `author`: `{kind: human|agent, name, session?}`
- `z` / layer membership
- optional `tags[]` and `semantic` (free-form label the agent or human attaches: `"db"`, `"T14.2"`)

### Operations (append-only log)

Prefer an **op log** with idempotent ops and LWW (or CRDT) merge — not full free-form mutations.

```
Op {
  id, ts, author,
  kind: create|update|delete|reorder|recognize|set_meta|begin_stroke|append_stroke|end_stroke
  target?, patch?
}
```

- Human Pencil: `begin_stroke` → many `append_stroke` (or datagram samples coalesced into ops) → `end_stroke` → optional `recognize`.
- Agent: usually `create`/`update` of primitives (no pressure series unless simulating ink).
- Undo = inverse op in log (shared undo stack visible to both).

**Why not pure CRDT libraries first:** scene graphs with connectors and recognition are awkward in generic CRDTs. Start with single-daemon authority + op log; if multi-writer races hurt, layer Yjs/Automerge-style counters later. One Mac + one pad is a weak concurrency case — keep it simple.

### Snapshots

Periodic compact snapshots for fast load; ops since snapshot for catch-up. Export formats: **JSON scene**, **SVG**, **PNG**, **PDF**.

---

## 5. Drawing intelligence (pad + daemon)

### Shape recognition

On stroke end (configurable: auto / button / double-tap Pencil):

1. Resample + simplify stroke.
2. Classify: line, polyline, arrow, rect, rounded-rect, ellipse, triangle, diamond, freehand-keep.
3. Fit geometry; **replace** or **overlay** (user preference). Keep original stroke in history for undo.
4. Optional: arrow heads from stroke velocity / hook shape.
5. Optional: “ink to connector” when stroke starts near object A port and ends near B.

Run recognition on **iPad for snappy UX**, re-validate on daemon for agent-visible canonical geom (or run only on daemon if latency is fine over LAN). Recommendation: **pad proposes, daemon commits** (one authority).

### Snapping

| Mode | Behaviour |
|------|-----------|
| Grid | Configurable spacing; shift temporarily disables |
| Object | Edges, centres, corners, equal-gap |
| Angle | 0/15/30/45/… for lines |
| Port | Connectors snap to object connection points (N/E/S/W + custom) |
| Align | Live guides when matching other objects (Keynote-style) |

Agent-created objects respect the same snap settings unless `snap: false`.

### Technical-diagram helpers

- Orthogonal connector routing with simple obstacle avoidance.
- Text inside / beside shapes with auto-size.
- Duplicate / distribute / align toolbar (and MCP equivalents).
- Measure tool: distance and angle readouts (screen only; also `measure` MCP).
- Symbol library: box, cylinder, actor, cloud, phone, server — expandable JSON packs.
- Templates: blank, flowchart, sequence-ish lanes, C4-ish containers, UI wireframe grid.

### Pencil specifics

| Feature | Use |
|---------|-----|
| Force / altitude / azimuth | Stroke width and optional calligraphy; store in samples |
| Predicted touches | Local ink only (do not commit predictions to op log) |
| Hover (Pencil Pro) | Cursor + snap previews without marks |
| Double-tap | Toggle eraser ↔ pen, or recognize last stroke (setting) |
| Barrel roll / squeeze | Optional tool palette (if hardware present) |
| Palm rejection | UIKit Pencil-only drawing where possible |
| Pencil-only mode | Ignore finger draws; finger = pan/zoom |

Finger: pan, pinch zoom, two-finger undo (optional). Pencil: draw.

---

## 6. Agent surface (MCP)

MCP server name: **`orthograph`**. Tools stay few and composable — avoid a 40-tool kitchen sink.

### Observation

| Tool | Returns | When |
|------|---------|------|
| `orthograph_status` | paired?, active doc, client viewport, authors online | Session start |
| `orthograph_list_docs` | recent documents | |
| `orthograph_open` | doc summary + revision | open or create |
| `orthograph_scene` | vector scene (filterable: layer, bbox, types, tags) | **primary observe** |
| `orthograph_summary` | compact textual/topology summary for tokens | planning |
| `orthograph_render` | PNG/WebP (full, bbox, or “viewport”); optional scale | visual check / OCR-ish |
| `orthograph_get_asset` | image bytes / path | deep look at embedded image |
| `orthograph_history` | recent ops | “what changed” |
| `orthograph_wait` | block until op or timeout | collab turn-taking |

`orthograph_scene` is the default. Use `orthograph_render` when layout aesthetics, handwriting, or screenshots matter. Prefer summaries for large docs.

**Scene JSON** should be agent-friendly: absolute coordinates, units in points, no proprietary binary. Include bbox on every object.

### Mutation

| Tool | Effect |
|------|--------|
| `orthograph_create` | one or many objects (batch) |
| `orthograph_update` | patch by id |
| `orthograph_delete` | by ids |
| `orthograph_stroke` | optional freehand (points) |
| `orthograph_recognize` | force recognition on stroke ids |
| `orthograph_group` / `ungroup` | |
| `orthograph_layer` | create/rename/reorder/visibility |
| `orthograph_undo` / `redo` | shared stack |
| `orthograph_focus` | pan iPad viewport to bbox / object (polite request) |
| `orthograph_presence` | set agent cursor label / colour / “drawing…” |

### Media & extras

| Tool | Effect |
|------|--------|
| `orthograph_push_image` | path or clipboard → image object on canvas |
| `orthograph_push_pdf_page` | optional later |
| `orthograph_export` | svg/png/pdf/json to host path |
| `orthograph_import_svg` | place as group of paths |
| `orthograph_set_guide` / grid | |

Clipboard on Mac: daemon reads `pbpaste` / NSPasteboard via small helper when agent says “clipboard”.

### Events (optional MCP resources / notifications)

If the host supports it: resource `orthograph://active` and notifications on `doc_changed` so agents can subscribe instead of polling. Fallback: `orthograph_wait` / `orthograph_history`.

### Authorship & trust

- All agent ops tagged `author.kind=agent`.
- Distinct default stroke/fill colour for agent (e.g. indigo) vs human (black/ink).
- Optional **propose mode**: agent ops land as translucent “ghost” objects until human taps Accept / Reject (important for “don’t mess up my diagram”).
- Mode is a document setting + MCP flag.

---

## 7. Collaboration model

### Turns without rigid locks

- Both can write anytime; ops ordered by daemon receive time.
- In-progress human stroke is exclusive on that stroke id (agent shouldn’t edit mid-stroke).
- Presence: human viewport rect + pencil hover; agent “gaze” bbox when calling `orthograph_focus` or while editing.
- Conflict: last update-wins per object field; deletes win over updates if delete is later.

### Starting a new diagram together

1. Agent: `orthograph_open` with `create: true`, title.
2. Daemon opens doc; iPad switches (banner: “Agent opened *Auth flow*”).
3. Agent draws scaffolding; human refines with Pencil.
4. Or human sketches first; agent `orthograph_scene` + annotates.

### Multi-agent

Multiple MCP clients can attach to one `orthographd`. Presence list shows all. Optional soft “lease” on a layer (`layer.locked_by`) to reduce collisions.

---

## 8. iPad UX (pad)

### Layout

- Infinite canvas, dark and light technical paper textures.
- Floating compact toolbar (pen, highlighter, eraser, shapes, text, connector, image, select).
- Minimap (toggle).
- Status chip: connected / doc title / agent present.
- Layer drawer; object inspector on selection.
- Snap and grid toggles always one tap away.

### Modes

| Mode | Purpose |
|------|---------|
| Ink | Freehand; recognition on lift optional |
| Shape | Tap-drag primitives |
| Connector | Port-to-port |
| Text | Tap to place; Scribble support |
| Select | Move/scale/rotate; multi-select lasso |
| Pan | Fist / finger only |
| Laser / point | Ephemeral pointer for “look here” (not in scene) |

### Agent UX on pad

- Toast when agent creates/deletes.
- Ghost proposals with Accept/Reject bar.
- “Follow agent” optional auto-pan.
- Agent objects subtly badged (corner pip) so authorship stays clear.

---

## 9. Integration with the rest of the fleet

| System | Integration |
|--------|-------------|
| **pigeon** | Transport only; no drawing-domain logic in pigeon |
| **spyder** | `orthograph_push_image` after `spyder` screenshot — annotate device UI |
| **blurter** | Optional: notify when agent needs human accept on ghosts |
| **mnemo** | Index exported diagrams / session links to doc ids |
| **bullseye** | Tag objects with target ids (`semantic: T14.2`); open diagram from target context later |
| **rustuml / PlantUML** | Export rough structure → text diagram; or import rendered PNG as underlay |
| **vellum** | Push PDF page images into canvas for markup |
| **Excalidraw MCP** (if present) | Interop via SVG import/export, not dual live sync |
| **Jevons** | **Consumer, not owner** — see §15. Overseer/workers use `orthograph_*` MCP; optional cockpit status chip and deep links. Not required to boot orthographd |

---

## 10. Implementation plan (phased)

### Phase 0 — Skeleton (1 vertical slice)

- `orthographd`: open empty doc, MCP `status` / `scene` / `create` rect, local JSON persistence.
- Mac debug viewer (simple web canvas) before iPad.
- Prove agent can draw a box and re-read it.

### Phase 1 — iPad + pigeon

- Pairing QR, session, op stream.
- Pencil freehand + pan/zoom.
- Live ink datagrams; commit on stroke end.
- Agent create visible on pad; human stroke visible in `orthograph_scene`.

### Phase 2 — Intelligence

- Recognition, grid/object snap, connectors, layers, undo.
- `orthograph_render`, `orthograph_push_image` (file + clipboard).
- Export SVG/PNG.

### Phase 3 — Collab polish

- Ghost/propose mode, presence, multi-doc UI, templates, symbol packs.
- `orthograph_summary` topology, frames/tags, focus requests.
- Propose-mode default for untrusted automation.

### Phase 4 — Depth (as needed)

- OCR handwriting → text objects.
- Constraint-ish align distribute.
- PDF underlay multi-page.
- Symbol packs (infra / UML / UI).
- Stroke-to-PlantUML heuristic export.
- iPhone companion viewer.

---

## 11. Tech choices (recommended)

| Layer | Choice | Why |
|-------|--------|-----|
| Daemon | Go | Matches spyder/blurter/jevons/pigeon ecosystem |
| MCP | HTTP like spyder | Always-on brew service |
| iOS | SwiftUI + PencilKit *or* custom Metal/CoreGraphics canvas | PencilKit is fast to ship but fights custom object model — **prefer custom canvas** for vector scene control; use PencilKit only if you accept bridging cost |
| Wire codec | JSON ops v1; CBOR later if ink volume hurts | Debuggable |
| Render on Mac | resvg / tiny-skia / cairo / pure Go vector | Deterministic `orthograph_render` for agents without iPad |
| Persistence | SQLite (ops + snapshot blob) + media files | Crash-safe, queryable |

**Canvas recommendation:** custom `CALayer`/`Metal` stroke renderer + hit testing. PencilKit is excellent for pure ink apps but poor as a collaborative structured-object editor with agent-authored primitives. Study Freeform / Concepts / Linea mental models; implement the subset we need.

---

## 12. Oracle / verification posture

(Per oracle-first doctrine — design so “done” is not self-attested.)

| Property | Oracle |
|----------|--------|
| Round-trip scene | Agent create → scene JSON equals golden (geometry tolerance) |
| Pad sync | Integration: op on MCP appears in iOS model fixture / UI test |
| Recognition | Fixed stroke fixtures → expected primitive class + geom error &lt; ε |
| Render stability | SVG/PNG hash or perceptual hash against goldens |
| Pairing | pigeon ceremony e2e already exists; orthographd smoke on top |
| No silent drop | Op log length monotonic; reconnect replay tested |

Headline demos for humans: “agent draws architecture skeleton while I label with Pencil”; “push spyder screenshot, agent circles the bug.”

---

## 13. Extra product ideas (beyond the ask)

1. **Semantic frames** — named regions agents can address (`frame:auth`) without pixel coords.
2. **Ghost propose / human accept** — trust boundary for aggressive agents.
3. **Underlay mode** — screenshot or PDF page locked under ink (design review, paper markup).
4. **Laser + voice note** — ephemeral point + short audio pin (stored as note object).
5. **Diff view** — highlight ops since revision N (agent: “what did we add?”).
6. **Topology summary** — auto graph of boxes + connectors for LLM (token-cheap).
7. **Code deep links** — object metadata `path:line` or bullseye id; tap opens on Mac later.
8. **Stroke replay** — scrub creation time for teaching / review.
9. **Handwriting OCR** (Apple Vision) → editable `text` objects.
10. **Agent brush styles** — dashed construction lines vs solid committed geometry.
11. **Read-only spectator** on Mac browser for when the pad is across the room.
12. **Template “interview”** — agent opens sequence-lane template and says “you draw the client; I’ll draw the server.”
13. **Pressure-as-emphasis** only for human ink; agent uses explicit `weight` field (don’t fake pressure).
14. **Safe clear** — agent cannot wipe doc without `confirm` flag + human toast.
15. **Session binding** — doc id written into agent transcript / mnemo so “the diagram we drew” is recallable.

---

## 14. Naming

| Candidate | Notes |
|-----------|--------|
| **Orthograph** | Clear, boring, good |
| **Slate** | Short binary `slated` |
| **Plumb** | Technical-drawing vibe |
| **Chalk** | Collaborative classroom vibe |
| **Figment** | Agent + imagination (maybe cute) |

Recommendation: **Orthograph** / `orthographd` / MCP `orthograph` until a better name sticks.

---

## 15. Product boundary: standalone with Jevons integration (not a Jevons feature)

### Decision

**Orthograph is a standalone product** (`orthographd` + iPad app + MCP `orthograph`), with **first-class Jevons integration** so it feels native in the cockpit. It is **not** a feature folded into `jevonsd` or the Jevons iOS shell.

This is the same pattern as mnemo (memory), spyder (devices), blurter (notify), pigeon (transport), claudia (harness): Jevons **uses and surfaces** them; it does not **own** their domain.

### Why not “just a Jevons feature”

| Pressure | If inside jevonsd / Jevon.app | If standalone |
|----------|-------------------------------|---------------|
| Charter | Drawing is a *capability*, not CEO arbitration. Charter already excludes “Jevons is memory / executor / oracle” — absorbing canvas logic repeats the mistake mnemo extraction fixed. | Matches constitution: Jevons directs; workers + tools act. |
| Lifecycle | Sketch releases couple to Jevons releases, cost clamp, thread model, UI freezes. | Independent version, brew formula, crash domain. |
| Consumers | Only usable when Jevons is the session host. | Any agent (Grok TUI, Claude, Cursor, one-shot workers) with MCP. |
| iOS surface | Jevons app is chat-first WKWebView; Pencil-grade canvas fights that shell. | Dedicated drawing app; Jevons remains chat/CEO. |
| Coupling | Op log, media, render, recognition all bloat jevonsd. | Thin integration surface (MCP + optional events). |
| Standalone merit | Hard to extract later once UI and pairing entangle. | “Merit of standalone” is free from day one. |

### Why not “standalone with no Jevons story”

Seamless cockpit use is a product requirement, not a nice-to-have. Without deliberate integration, you get two pairing ceremonies, two status chips, and agents that don’t know “the diagram Jevons is talking about.”

### Integration contract (how it feels like one product)

**Hard boundary — Orthograph owns**

- Document model, op log, render, recognition, snap, Pencil UX
- `orthographd` process, persistence under `~/.orthograph/`
- iPad *Orthograph* app and its pigeon backend role
- MCP tool surface (`orthograph_*`)

**Hard boundary — Jevons owns**

- CEO conversation, threads, cost, policy
- When to open/focus a diagram in service of a task
- Optional chat affordances (“show on pad”, “open diagram for T14”)

**Seam (integration layer — small, explicit)**

| Seam | Mechanism |
|------|-----------|
| Agent tools | Workers and overseer call MCP `orthograph_*` like any other fleet MCP (user-scoped install). |
| Deep link / open | `orthograph://doc/<id>` or `orthograph open <id>`; Jevons can shell/MCP “open this doc.” |
| Presence in cockpit | Optional: jevonsd subscribes to orthographd status (paired?, active doc title, agent ghosts pending) and shows a compact chip in web UI — **read-only projection**, not a second document store. |
| Pairing UX | Prefer **one mental model**: either Orthograph reuses pigeon pairing patterns (QR from `orthographd pair`) *or* a later “Jevons-linked pair” that reuses device identity without merging daemons. Do **not** tunnel orthograph ops through jevonsd as a proxy. |
| Event lane | Optional blurter/orthograph notifications when human must Accept ghost ops; Jevons can also surface those in chat if desired. |
| Memory | Doc ids + export paths indexed by mnemo from sessions; Jevons never stores scene graphs. |
| iOS | **Two apps** (Jevons chat + Orthograph draw). Not a tab inside Jevons. Continuity = shared owner, deep links, simultaneous use on iPad Stage Manager / split view. |

**Anti-patterns**

- Embedding the canvas in Jevons WKWebView “to ship faster”
- Proxying all `orthograph_*` through `jevons_*` wrappers that reimplement semantics
- Storing scenes under `~/.jevons/`
- Making orthographd a subprocess only Jevons can start (agents must reach it cold)

### “Seamless” definition of done (integration)

1. From a Jevons (or any MCP) agent: create/open diagram, draw, observe vector + pixels, push image — no special Jevons-only API required.
2. From Jevons chat: owner can say “open an orthograph diagram for this design” and the overseer does the right `orthograph_*` calls; pad shows the doc.
3. Cockpit can show *that* Orthograph is paired and *which* doc is active (status only).
4. Uninstalling or stopping Jevons does **not** break Orthograph for other agents.
5. Uninstalling Orthograph does **not** break Jevons (tools fail soft / status empty).

### Product naming under this split

- Repo / brew: `orthograph` / `orthographd` (standalone)
- MCP server name: `orthograph`
- Jevons docs: “Orthograph integration” under fleet/providers or companion tools — not a Jevons subsystem chapter that owns the domain model

---

## 16. Open decisions for you

1. **PencilKit vs custom canvas** — custom recommended; confirm.
2. **Propose-mode default** — on or off for agent draws?
3. **Recognition aggressiveness** — auto on every stroke vs explicit?
4. **Public vs private repo** when created under `~/work/github.com/marcelocantos/`?
5. **MVP cut** — Phase 0–1 only first (observable shared canvas), intelligence in Phase 2?
6. **Mac preview** — required in v1 or iPad-only until later?
7. **Product boundary** — standalone + Jevons seam (recommended above); confirm or reject.
8. **iOS** — two apps (recommended) vs Jevons tab later; confirm.

---

## 17. Suggested first milestone

> A brew-installed `orthographd` on the Mac, paired iPad app over pigeon, human Pencil strokes and agent `orthograph_create` sharing one vector document, agent can `orthograph_scene` and `orthograph_render`, and `orthograph_push_image` places a host image on the pad.

That single milestone already changes how you design with agents; recognition and snap make it excellent.

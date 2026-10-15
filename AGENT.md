# AGENT.md — Detective Files (compact project summary)

## What this is
Single-player 2D noir detective game. Native Godot 4.7 client (GDScript),
server-authoritative Go backend. Cases (sealed in order): "The Last Train"
(1947 railway murder) then "The Tunis Ledger" (1947 Tunis hotel murder).
All persons fictional.

## Stack
Backend: Go, **Gin** (`gin.New()`), pgxpool, go-redis, golang-migrate,
`slog`, testify, Argon2id. DB: Postgres 16 + Redis 7 via `compose.yaml`
(services only — API runs on host). Client: Godot 4.7, one global
`game/themes/noir_theme.tres`, 3 autoloads (Session, ApiClient, Nav).

## Run / verify
`make services` → `make api` → `make client`. Checks: `make test`
(24 backend tests), `make check` (Godot headless), `make e2e`
(16-step playthrough), `make shots` (screenshots → `/tmp/dfshots/`).

## Rules that matter
- Server owns all state. Client never sends scores/flags. Spoiler-free
  `GET /cases/:id` (no culprit, responses, or evidence text — enforced by
  `TestNoCulpritLeak`). Accusations: one transaction, idempotent, XP once
  via `PRIMARY KEY (user_id, case_id)`.
- Sessions: 32-byte token, Bearer header, only SHA-256 in Redis + TTL.
- Art: `docs/art-bible.md` is binding (§10 = corkboard look). One Theme;
  per-node styling ONLY via `DFUI`/`BoardUI` copy-from-theme helpers
  (Control-level custom-type lookup is unreliable — do not revert to
  `theme_type_variation`). Art slots: one const each in `ArtPaths`;
  licenses in `game/assets/LICENSES.md` before use. No AI art, no emoji
  icons, CC0/CC-BY/OFL/MIT only (exception: Special Elite is Apache-2.0 —
  mandated by spec, recorded in LICENSES.md).
- Screens are minimal `.tscn` roots + code-built UI. Screenshot rig:
  `game/tools/screenshot.tscn` (`--scene/--shot/--flow/--view/--suspect/--clip`).
- Commissions pending: `game/assets/NEEDED.md` (portraits, backdrops,
  evidence art, logo sting). Placeholders are monogram tiles by design.

## Layout
`backend/{cmd/api,internal/{config,handlers,middleware,services,repository,models,auth,cases,investigations,cache},migrations,tests,content/cases/}`
`game/{scenes,scripts/{api,state,ui,screens},themes,assets,tools}`
`docs/{art-bible,asset-picks,case-the-last-train,API}.md`

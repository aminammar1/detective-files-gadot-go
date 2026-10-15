# Detective Files — The Last Train

Single-player 2D noir detective mystery. Native Godot 4 client (GDScript),
server-authoritative Go backend (Gin + PostgreSQL + Redis). All persons and
cases fictional.

## Run it

```bash
# 1. Services (PostgreSQL + Redis only)
docker compose up -d

# 2. Backend (host; migrations auto-run, case JSON validated at startup)
cd backend
DATABASE_URL="postgres://postgres:postgres@localhost:5432/detectivefiles?sslmode=disable" \
REDIS_ADDR="localhost:6379" \
go run ./cmd/api
# curl localhost:8080/health  -> {"ok":true}

# 3. Client
godot --path game            # or open game/project.godot in the editor
# Point at another host: DF_API_URL=http://host:8080 godot --path game
```

## Layout

- `game/` — Godot 4 client (`scenes/`, `scripts/{api,state,ui,screens,models}`,
  `themes/noir_theme.tres`, `assets/`, `tools/screenshot.tscn`).
- `backend/` — Go API (`cmd/api`, `internal/…`, `migrations/`, `tests/`,
  `content/cases/the_last_train.json`). See `backend/README.md`.
- `docs/` — `art-bible.md`, `asset-picks.md`, `case-the-last-train.md`, `API.md`.
- `compose.yaml` — PostgreSQL 16 + Redis 7 only (persistent `pgdata` volume).

## Verify

```bash
cd backend && go test -count=1 ./...   # 24 tests: auth, sessions, isolation,
                                       # evidence/interrogation rules, accusations,
                                       # idempotency, Redis recovery, case validity
godot --headless --path game --quit    # GDScript parse check
ls /tmp/dfshots/                       # UI screenshots (see game/tools/)
make e2e                               # full manual E2E (needs API + services up)
```

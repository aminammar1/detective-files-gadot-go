# Detective Files — backend

Go + Gin + pgx + go-redis + golang-migrate + slog + testify.
Module `detectivefiles`, rooted at `backend/`. No fake success responses:
every endpoint is backed by Postgres/Redis and covered by tests.

## Run

```bash
# 1. Start Postgres 16 + Redis 7 (repo root)
cd /home/amine/gadot-games/detective-files-gadot-go
docker compose up -d

# 2. Configure (placeholders in .env.example)
cd backend
cp .env.example .env   # defaults already point at compose services
export $(cat .env | xargs)   # or export DATABASE_URL / REDIS_ADDR directly

# 3. Boot (migrations auto-run, case JSON validated at startup)
go run ./cmd/api
# listening on :8080 — try: curl localhost:8080/health
```

Config (env, fail-fast): `PORT` (default 8080), `DATABASE_URL` (required),
`REDIS_ADDR` (required), `SESSION_TTL_HOURS` (default 168),
`RATE_LIMIT_PER_MIN` (default 10).

## Test

```bash
cd backend
go test ./...            # needs compose services on localhost:5432/:6379
TEST_DATABASE_URL=... TEST_REDIS_ADDR=... go test ./...   # overrides
```

If Postgres/Redis are unreachable, integration tests **SKIP** with a clear
message instead of fake-passing (`setupApp` in `tests/helpers_test.go`).
Pure unit tests (hashing, ranks, case validity, deducibility, no-leak) always run.
Note: integration tests use `TRUNCATE users CASCADE` + Redis `FLUSHDB` —
point them at a dev database only.

```bash
go vet ./... && gofmt -l .   # both must be clean
```

## Layout

- `cmd/api/main.go` — boot: config, pgpool, auto-migrations, Redis, cases, Gin.
- `internal/config` — env loading + fail-fast validation.
- `internal/models` — case structs + sanitized public DTOs.
- `internal/cases` — `go:embed` case loading + strict cross-ref validation + `Sanitize`.
- `content/embed.go` — embed shim (the `content/cases/*.json` files are untouched).
- `internal/auth` — Argon2id (64MiB/t=1/p=4/32-byte key) + session tokens (SHA-256 stored).
- `internal/cache` — failure-tolerant Redis: sessions, rate limits, case cache (5 min), leaderboard (60 s).
- `internal/repository` — pgx queries, always scoped by authenticated user id.
- `internal/services` — ranks, deflections, accusation transaction (`Accuse`), case-cache helper.
- `internal/middleware` — bearer auth, `{"error":{"code","message"}}` envelope.
- `internal/handlers` — all endpoints + router.
- `migrations/` — `000001_init` (schema, pgcrypto+citext, indexes), `000002_achievements` (seed).
- `tests/` — unit + HTTP integration suite.
- `../docs/API.md` — full endpoint reference.

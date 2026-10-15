# Detective Files — API v1

Base URL: `http://localhost:8080/api/v1` (health/ready also served at `/health`, `/ready`).

All requests and responses are JSON. Errors always look like:

```json
{"error": {"code": "snake_case", "message": "Human-readable sentence."}}
```

Auth: `Authorization: Bearer <token>` header. No cookies. Sessions live in
Redis (`sess:<sha256hex>` → user id) with `SESSION_TTL_HOURS` TTL.
If Redis is down, authenticated requests 401 (sessions require Redis);
cache-only paths (case metadata, leaderboard) degrade to Postgres.

## Conventions

| Status | Meaning |
|---|---|
| 200 | OK (also: idempotent replay of start/accusation) |
| 201 | Created (register, first start, first accusation) |
| 401 | Missing/bad/expired/revoked token, or bad login |
| 404 | Unknown case, or no investigation started yet |
| 409 | Duplicate registration (field not named), or write to a closed case |
| 422 | Validation failure (bad payload, bad ids, unproven exhibit) |
| 429 | Rate limit (login/register: 10 attempts / minute / IP by default) |
| 503 | `/ready` when Postgres or Redis is down |

## System

### GET /health → 200
```json
{"ok": true}
```

### GET /ready → 200 (or 503)
```json
{"ok": true, "postgres": true, "redis": true}
```

## Auth

### POST /auth/register → 201
Request:
```json
{"email": "sam@graywater.gov", "username": "sam_spade", "password": "trenchcoat99"}
```
Rules: valid email; username `^[a-zA-Z0-9_]{3,20}$`; password 8–72 chars (else `422 validation_error`). Duplicate email **or** username → `409 conflict` (never says which). Creates the player profile too.
Response:
```json
{"user": {"id": "uuid", "email": "sam@graywater.gov", "username": "sam_spade"}, "token": "64-hex-chars"}
```

### POST /auth/login → 200
Request: `{"email": "...", "password": "..."}`. Any failure (unknown email
or wrong password) → `401 {"code":"invalid_credentials","message":"Badge number or password didn't check out."}`. Never reveals which field failed.
Response: same shape as register.

### POST /auth/logout (auth) → 200
Revokes the session. Response: `{"ok": true}`.

### GET /auth/me (auth) → 200
```json
{
  "user": {"id": "uuid", "email": "sam@graywater.gov", "username": "sam_spade"},
  "profile": {"display_name": "sam_spade", "rank": "Rookie", "level": 1, "xp": 0}
}
```

## Player

### GET /player/profile (auth) → 200
```json
{"profile": {"display_name": "Sam Spade", "rank": "Gumshoe", "level": 2, "xp": 1185}}
```

### PATCH /player/profile (auth) → 200
Request: `{"display_name": "Sam Spade"}` (1–30 chars, else `422 validation_error`).
Response: `{"profile": {...}}` updated.

### GET /player/stats (auth) → 200
```json
{"cases_started": 1, "cases_solved": 1, "cases_failed": 0, "total_xp": 160, "level": 1, "rank": "Rookie"}
```

### GET /player/achievements (auth) → 200
All achievement defs; locked rows show `"unlocked_at": null`.
```json
[
  {"id": "started_first_case", "name": "Took the Badge", "description": "...", "unlocked_at": "2026-…"},
  {"id": "perfect_deduction", "name": "Sherlock of Graywater", "description": "...", "unlocked_at": null}
]
```

## Cases

### GET /cases (auth) → 200
Cases after the first carry `"requires_case"` and report `"status": "locked"`
until the required file is closed (any verdict). Sealed files hide their
content: detail/start return `409 case_locked`.
```json
[{"id": "the_last_train", "title": "The Last Train", "difficulty": "introductory",
  "status": "unlocked", "started_at": null, "completed_at": null, "correct": null},
 {"id": "the_tunis_ledger", "title": "The Tunis Ledger", "difficulty": "standard",
  "status": "locked", "requires_case": "the_last_train", "started_at": null,
  "completed_at": null, "correct": null}]
```
After judging: `"status": "completed"`, with `completed_at` and `correct`.

### GET /cases/:caseId (auth) → 200
Spoiler-free. Never includes `culprit_suspect_id`, `accusation_rules`,
`scoring`, topic responses, contradiction responses, or evidence content.
Unknown id → `404 not_found`.
```json
{
  "id": "the_last_train", "title": "The Last Train", "difficulty": "introductory",
  "briefing": "Graywater Police, memorandum…", "audit_letter": "\"Ed — …\"",
  "locations": [{"id": "berth_c", "name": "Berth C — Vane's compartment", "description": "…", "asset": "backgrounds/berth_c_1920.png"}],
  "suspects": [{"id": "edwin_marsh", "name": "Edwin Marsh", "role": "Junior partner, Meridian Freight",
    "description": "…", "dossier": "…", "alibi": "Retired at midnight…",
    "asset": "portraits/edwin_marsh_512.png",
    "topics": [{"id": "t_where", "prompt": "Where were you after midnight?"}]}]
}
```

### POST /cases/:caseId/start (auth) → 201 / 200
Idempotent: first call creates (201), repeats return the existing row (200).
Awards `started_first_case`. Response:
```json
{"investigation": {"id": "uuid", "case_id": "the_last_train", "status": "active",
  "started_at": "2026-…", "completed_at": null}}
```

### GET /cases/:caseId/state (auth) → 200
404 (`not_found`) if never started.
```json
{
  "investigation": {"id": "uuid", "case_id": "the_last_train", "status": "active", "started_at": "…", "completed_at": null},
  "discovered_evidence": [{"id": "flask", "name": "Vane's silver flask", "location_id": "berth_c",
    "description": "…", "proves": "…", "asset": "evidence/flask_800.png", "discovered_at": "…"}],
  "testimonies": [{"id": "tommy_timeline", "title": "Tommy's timeline", "text": "…", "unlocked_at": "…"}],
  "stats": {"evidence_found": 3, "actions": 2, "interrogations": 1}
}
```

### POST /cases/:caseId/actions (auth) → 200
Request: `{"action": "search", "location_id": "berth_c"}`.
Unknown action/location → `422 validation_error`. Discovers ALL undiscovered
evidence at the location; repeats return `"found": []` with `already_known`
ids. First-ever find awards `found_first_evidence`.
```json
{"found": [{"id": "flask", "name": "…", "location_id": "berth_c", "description": "…", "proves": "…", "asset": "…"}],
 "already_known": []}
```

### POST /cases/:caseId/interrogate (auth) → 200
Request: `{"suspect_id": "edwin_marsh", "topic_id": "t_where"}`.
Bad suspect/topic → `422 validation_error`. Persists an `ask` event + action.
```json
{"suspect_id": "edwin_marsh", "topic_id": "t_where", "response": "Retired at midnight…"}
```

### POST /cases/:caseId/present-evidence (auth) → 200
Request: `{"suspect_id": "tommy_beck", "evidence_id": "bond_stub"}`.
Undiscovered exhibit → `422 {"code":"unproven","message":"You don't have that in hand."}`.
Contradiction hit (persists a `present` event; testimony unlock awards `broke_alibi`):
```json
{"contradiction": true, "response": "(He looks at the torn half…)",
 "unlocked_testimony": {"id": "tommy_timeline", "title": "Tommy's timeline", "text": "…"}}
```
No contradiction → in-voice deflection (original copy, never case text):
```json
{"contradiction": false, "response": "Cut the theater, detective…", "unlocked_testimony": null}
```

### POST /cases/:caseId/accusations (auth) → 201 / 200
Single DB transaction. A stored `case_results` row makes it idempotent
(200 replay, no new XP); concurrent double-accusations collapse to one row.
Bad suspect → `422 validation_error`.
```json
{"suspect_id": "edwin_marsh", "case_id": "the_last_train",
 "correct": true, "proven": true, "score": 1600, "xp_awarded": 160,
 "completed": true, "result": "guilty"}
```
Outcome table (`result`: `guilty` | `unproven` | `wrong`):

| Accusation | score | xp_awarded |
|---|---|---|
| Correct + all `required_evidence` discovered + all `required_testimonies` unlocked | `base_correct + 50×#evidence + 100×#testimonies` (max 1600) | `score/10` (integer) |
| Correct person, proof incomplete (`unproven`, `correct: true`) | `1000` (= xp×10 for consistency) | `100` (flat consolation) |
| Wrong suspect (`correct: false`) | `250` (= xp×10 for consistency) | `25` (flat consolation) |

Note: the JSON `consolation_unproven` (100) / `consolation_wrong` (25) are
**XP values** per the case doc; `score` mirrors `xp×10` on those rows.
Judging marks the investigation `completed`, adds XP (recomputing
`level = 1 + xp/1000`, rank: L1 Rookie, L2 Gumshoe, L3 Sleuth, L4 Inspector,
L5+ Chief Inspector), awards `solved_last_train` on a proven solve and
`perfect_deduction` when all 8 evidence + both testimonies were gathered.
Writes to a closed case otherwise → `409 case_closed`.

## Leaderboard

### GET /leaderboard (auth) → 200
Top 20 by XP desc. Cached in Redis 60s (`lb:top`); miss falls back to
Postgres; Redis failure degrades to a miss, never an error.
```json
[{"username": "sam_spade", "display_name": "Sam Spade", "level": 1,
  "rank": "Rookie", "xp": 160, "solved": 1}]
```

## Error codes

`validation_error` (422), `invalid_credentials` (401), `unauthorized` (401),
`not_found` (404), `conflict` (409), `case_closed` (409), `case_locked` (409),
`unproven` (422), `rate_limited` (429), `internal` (500).

Solved achievements are per case (`solved_last_train`, `solved_tunis_ledger`);
`perfect_deduction` is awarded only for a flawless Last Train solve.

package tests

import (
	"bytes"
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"detectivefiles/internal/auth"
)

const caseID = "the_last_train"

// ---- auth flows ----

func TestAuthRegisterLoginMe(t *testing.T) {
	ta := setupApp(t)
	_, token := registerUser(t, ta, "marlowe")

	status, raw := ta.do(t, "GET", "/api/v1/auth/me", token, nil)
	require.Equal(t, 200, status, string(raw))
	m := decode(t, raw)
	require.Contains(t, m, "user")
	require.Contains(t, m, "profile")

	// Logout revokes; the token must then be rejected.
	status, _ = ta.do(t, "POST", "/api/v1/auth/logout", token, nil)
	require.Equal(t, 200, status)
	status, raw = ta.do(t, "GET", "/api/v1/auth/me", token, nil)
	require.Equal(t, 401, status, string(raw))
}

func TestAuthMiddlewareRejects(t *testing.T) {
	ta := setupApp(t)
	status, _ := ta.do(t, "GET", "/api/v1/auth/me", "", nil)
	require.Equal(t, 401, status)
	status, _ = ta.do(t, "GET", "/api/v1/auth/me", "bogus-token", nil)
	require.Equal(t, 401, status)
}

func TestExpiredSessionRejected(t *testing.T) {
	ta := setupApp(t)
	_, token := registerUser(t, ta, "spade")
	// Shorten this session's TTL directly, then let it lapse.
	ctx := context.Background()
	uid, ok := ta.app.Cache.SessionGet(ctx, auth.TokenHash(token))
	require.True(t, ok)
	require.NoError(t, ta.app.Cache.SessionSetWithTTL(ctx, auth.TokenHash(token), uid, time.Second))
	time.Sleep(1200 * time.Millisecond)
	status, _ := ta.do(t, "GET", "/api/v1/auth/me", token, nil)
	require.Equal(t, 401, status)
}

func TestInvalidPayloadsAndDuplicates(t *testing.T) {
	ta := setupApp(t)
	bad := []map[string]any{
		{"email": "not-an-email", "username": "goodname", "password": "trenchcoat99"},
		{"email": "a@b.com", "username": "ab", "password": "trenchcoat99"},
		{"email": "a@b.com", "username": "bad-name!", "password": "trenchcoat99"},
		{"email": "a@b.com", "username": "goodname", "password": "short"},
		{"email": "a@b.com", "username": "goodname"},
	}
	for i, b := range bad {
		status, raw := ta.do(t, "POST", "/api/v1/auth/register", "", b)
		require.Equal(t, 422, status, "case %d: %s", i, string(raw))
		require.Equal(t, "validation_error", decode(t, raw)["error"].(map[string]any)["code"])
	}

	status, _ := ta.do(t, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": "dup@example.com", "username": "unique_one", "password": "trenchcoat99",
	})
	require.Equal(t, 201, status)
	// Duplicate email and duplicate username both 409 without saying which.
	status, raw := ta.do(t, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": "dup@example.com", "username": "other_name", "password": "trenchcoat99",
	})
	require.Equal(t, 409, status, string(raw))
	status, raw = ta.do(t, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": "other@example.com", "username": "unique_one", "password": "trenchcoat99",
	})
	require.Equal(t, 409, status, string(raw))
	for _, r := range []string{string(raw)} {
		require.NotContains(t, r, "email")
		require.NotContains(t, r, "username")
	}

	// Login failure is always the generic 401, never revealing the field.
	status, raw = ta.do(t, "POST", "/api/v1/auth/login", "", map[string]any{
		"email": "dup@example.com", "password": "wrongpassword",
	})
	require.Equal(t, 401, status)
	errObj := decode(t, raw)["error"].(map[string]any)
	require.Equal(t, "invalid_credentials", errObj["code"])
	require.Equal(t, "Badge number or password didn't check out.", errObj["message"])

	status, raw = ta.do(t, "POST", "/api/v1/auth/login", "", map[string]any{
		"email": "nobody@example.com", "password": "wrongpassword",
	})
	require.Equal(t, 401, status)
	require.Equal(t, "invalid_credentials", decode(t, raw)["error"].(map[string]any)["code"])

	// Good login round-trips.
	status, raw = ta.do(t, "POST", "/api/v1/auth/login", "", map[string]any{
		"email": "dup@example.com", "password": "trenchcoat99",
	})
	require.Equal(t, 200, status, string(raw))
	require.Contains(t, decode(t, raw), "token")
}

func TestLoginRateLimited(t *testing.T) {
	ta := setupApp(t)
	ta.app.Cfg.RateLimitPerMin = 3
	for i := 0; i < 3; i++ {
		status, _ := ta.do(t, "POST", "/api/v1/auth/login", "", map[string]any{
			"email": "x@example.com", "password": "wrongpassword",
		})
		require.Equal(t, 401, status)
	}
	status, raw := ta.do(t, "POST", "/api/v1/auth/login", "", map[string]any{
		"email": "x@example.com", "password": "wrongpassword",
	})
	require.Equal(t, 429, status, string(raw))
}

// ---- player isolation ----

func TestPlayerDataIsolation(t *testing.T) {
	ta := setupApp(t)
	_, tokA := registerUser(t, ta, "archer")
	_, tokB := registerUser(t, ta, "hammer")

	status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tokA, nil)
	require.Equal(t, 201, status)
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/actions", tokA,
		map[string]any{"action": "search", "location_id": "berth_c"})
	require.Equal(t, 200, status)

	// B never started: state is 404, and B's list shows unlocked.
	status, _ = ta.do(t, "GET", "/api/v1/cases/"+caseID+"/state", tokB, nil)
	require.Equal(t, 404, status)
	status, raw := ta.do(t, "GET", "/api/v1/cases", tokB, nil)
	require.Equal(t, 200, status)
	var list []map[string]any
	require.NoError(t, jsonUnmarshal(raw, &list))
	require.Len(t, list, 2)
	byID := map[string]map[string]any{}
	for _, e := range list {
		byID[e["id"].(string)] = e
	}
	require.Equal(t, "unlocked", byID["the_last_train"]["status"])
	require.Equal(t, "locked", byID["the_tunis_ledger"]["status"])

	// A's state shows evidence; B has none anywhere.
	status, raw = ta.do(t, "GET", "/api/v1/cases/"+caseID+"/state", tokA, nil)
	require.Equal(t, 200, status)
	require.Len(t, decode(t, raw)["discovered_evidence"].([]any), 3)
}

// ---- evidence rules ----

func TestEvidenceSearchRules(t *testing.T) {
	ta := setupApp(t)
	_, tok := registerUser(t, ta, "chan")

	status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 201, status)

	// Unknown action and unknown location are 422.
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/actions", tok,
		map[string]any{"action": "ransack", "location_id": "berth_c"})
	require.Equal(t, 422, status)
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/actions", tok,
		map[string]any{"action": "search", "location_id": "engine_room"})
	require.Equal(t, 422, status)

	status, raw := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/actions", tok,
		map[string]any{"action": "search", "location_id": "berth_c"})
	require.Equal(t, 200, status)
	m := decode(t, raw)
	require.Len(t, m["found"].([]any), 3)

	// Rediscovery is idempotent: nothing new, no duplicate rows.
	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/actions", tok,
		map[string]any{"action": "search", "location_id": "berth_c"})
	require.Equal(t, 200, status)
	m = decode(t, raw)
	require.Len(t, m["found"].([]any), 0)
	require.Len(t, m["already_known"].([]any), 3)

	var n int
	require.NoError(t, ta.app.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM discovered_evidence`).Scan(&n))
	require.Equal(t, 3, n)
}

// ---- interrogation + presenting ----

func TestInterrogateValidation(t *testing.T) {
	ta := setupApp(t)
	_, tok := registerUser(t, ta, "mason")
	status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 201, status)

	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/interrogate", tok,
		map[string]any{"suspect_id": "no_such_person", "topic_id": "t_where"})
	require.Equal(t, 422, status)
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/interrogate", tok,
		map[string]any{"suspect_id": "edwin_marsh", "topic_id": "no_such_topic"})
	require.Equal(t, 422, status)

	status, raw := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/interrogate", tok,
		map[string]any{"suspect_id": "edwin_marsh", "topic_id": "t_where"})
	require.Equal(t, 200, status)
	require.Contains(t, decode(t, raw)["response"], "Retired at midnight")
}

func TestPresentEvidenceRules(t *testing.T) {
	ta := setupApp(t)
	_, tok := registerUser(t, ta, "wolfe")
	status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 201, status)

	// Undiscovered evidence cannot be presented: 422 unproven.
	status, raw := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/present-evidence", tok,
		map[string]any{"suspect_id": "tommy_beck", "evidence_id": "bond_stub"})
	require.Equal(t, 422, status, string(raw))
	require.Equal(t, "unproven", decode(t, raw)["error"].(map[string]any)["code"])

	// Search everywhere, then present a non-contradiction: deflection.
	for _, loc := range []string{"berth_c", "dining_car", "vestibule_car3", "observation_deck"} {
		status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/actions", tok,
			map[string]any{"action": "search", "location_id": loc})
		require.Equal(t, 200, status)
	}
	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/present-evidence", tok,
		map[string]any{"suspect_id": "edwin_marsh", "evidence_id": "flask"})
	require.Equal(t, 200, status)
	m := decode(t, raw)
	require.Equal(t, false, m["contradiction"])
	require.NotEmpty(t, m["response"])

	// Real contradiction unlocks testimony.
	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/present-evidence", tok,
		map[string]any{"suspect_id": "tommy_beck", "evidence_id": "bond_stub"})
	require.Equal(t, 200, status)
	m = decode(t, raw)
	require.Equal(t, true, m["contradiction"])
	require.NotNil(t, m["unlocked_testimony"])
	require.Equal(t, "tommy_timeline", m["unlocked_testimony"].(map[string]any)["id"])

	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/present-evidence", tok,
		map[string]any{"suspect_id": "vivienne_hart", "evidence_id": "matchbook"})
	require.Equal(t, 200, status)
	require.Equal(t, true, decode(t, raw)["contradiction"])

	// State shows both testimonies with unlock times.
	status, raw = ta.do(t, "GET", "/api/v1/cases/"+caseID+"/state", tok, nil)
	require.Equal(t, 200, status)
	require.Len(t, decode(t, raw)["testimonies"].([]any), 2)
}

// ---- accusations ----

func fullSolve(t *testing.T, ta *testApp, tok string) {
	t.Helper()
	for _, loc := range []string{"berth_c", "dining_car", "vestibule_car3", "observation_deck"} {
		status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/actions", tok,
			map[string]any{"action": "search", "location_id": loc})
		require.Equal(t, 200, status)
	}
	for _, p := range [][2]string{{"tommy_beck", "bond_stub"}, {"vivienne_hart", "matchbook"}} {
		status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/present-evidence", tok,
			map[string]any{"suspect_id": p[0], "evidence_id": p[1]})
		require.Equal(t, 200, status)
	}
}

func profileXP(t *testing.T, ta *testApp, tok string) int {
	t.Helper()
	status, raw := ta.do(t, "GET", "/api/v1/player/profile", tok, nil)
	require.Equal(t, 200, status)
	return int(decode(t, raw)["profile"].(map[string]any)["xp"].(float64))
}

func TestCorrectAccusationGuiltyAndIdempotent(t *testing.T) {
	ta := setupApp(t)
	_, tok := registerUser(t, ta, "holmes")
	status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 201, status)
	fullSolve(t, ta, tok)

	status, raw := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/accusations", tok,
		map[string]any{"suspect_id": "edwin_marsh"})
	require.Equal(t, 201, status, string(raw))
	m := decode(t, raw)
	require.Equal(t, true, m["correct"])
	require.Equal(t, true, m["proven"])
	require.Equal(t, "guilty", m["result"])
	require.Equal(t, float64(1600), m["score"], "1000 + 50*8 + 100*2")
	require.Equal(t, float64(160), m["xp_awarded"])
	require.Equal(t, 160, profileXP(t, ta, tok))

	// Replay: same row, no new XP (200, one row, xp unchanged).
	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/accusations", tok,
		map[string]any{"suspect_id": "edwin_marsh"})
	require.Equal(t, 200, status, string(raw))
	require.Equal(t, float64(160), decode(t, raw)["xp_awarded"])
	require.Equal(t, 160, profileXP(t, ta, tok))
	var rows int
	require.NoError(t, ta.app.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM case_results`).Scan(&rows))
	require.Equal(t, 1, rows)

	// Perfect deduction + solve achievements unlocked.
	status, raw = ta.do(t, "GET", "/api/v1/player/achievements", tok, nil)
	require.Equal(t, 200, status)
	unlocked := map[string]bool{}
	var achs []map[string]any
	require.NoError(t, jsonUnmarshal(raw, &achs))
	for _, a := range achs {
		if a["unlocked_at"] != nil {
			unlocked[a["id"].(string)] = true
		}
	}
	for _, id := range []string{"started_first_case", "found_first_evidence", "broke_alibi", "solved_last_train", "perfect_deduction"} {
		require.True(t, unlocked[id], "achievement %q should be unlocked", id)
	}
}

func TestUnprovenAndWrongAccusations(t *testing.T) {
	ta := setupApp(t)
	_, tokU := registerUser(t, ta, "unproven")
	status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tokU, nil)
	require.Equal(t, 201, status)
	// Right person, empty hands: unproven, 100 XP, score mirrors xp*10.
	status, raw := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/accusations", tokU,
		map[string]any{"suspect_id": "edwin_marsh"})
	require.Equal(t, 201, status, string(raw))
	m := decode(t, raw)
	require.Equal(t, true, m["correct"])
	require.Equal(t, false, m["proven"])
	require.Equal(t, "unproven", m["result"])
	require.Equal(t, float64(100), m["xp_awarded"])
	require.Equal(t, float64(1000), m["score"])
	require.Equal(t, 100, profileXP(t, ta, tokU))

	_, tokW := registerUser(t, ta, "wrongguy")
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tokW, nil)
	require.Equal(t, 201, status)
	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/accusations", tokW,
		map[string]any{"suspect_id": "vivienne_hart"})
	require.Equal(t, 201, status, string(raw))
	m = decode(t, raw)
	require.Equal(t, false, m["correct"])
	require.Equal(t, false, m["proven"])
	require.Equal(t, "wrong", m["result"])
	require.Equal(t, float64(25), m["xp_awarded"])
	require.Equal(t, float64(250), m["score"])
	require.Equal(t, 25, profileXP(t, ta, tokW))

	// Invalid suspect is 422.
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/accusations", tokW,
		map[string]any{"suspect_id": "ghost"})
	require.Equal(t, 422, status)
}

func TestConcurrentAccusationsAwardOnce(t *testing.T) {
	ta := setupApp(t)
	_, tok := registerUser(t, ta, "racer")
	status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 201, status)
	fullSolve(t, ta, tok)

	const n = 8
	var wg sync.WaitGroup
	type res struct {
		status int
		raw    []byte
	}
	results := make([]res, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// httptest + pgxpool are goroutine-safe; no require calls here
			// (assertions run in the test goroutine after Wait).
			router := ta.app.NewRouter()
			body := []byte(`{"suspect_id":"edwin_marsh"}`)
			req := httptest.NewRequest("POST", "/api/v1/cases/"+caseID+"/accusations", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tok)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			results[i] = res{w.Code, w.Body.Bytes()}
		}(i)
	}
	wg.Wait()
	for i, r := range results {
		require.Contains(t, []int{200, 201}, r.status, "goroutine %d: %s", i, string(r.raw))
		require.Equal(t, "guilty", decode(t, r.raw)["result"])
	}
	// Exactly one row and exactly one XP award despite the race.
	var rows, total int
	require.NoError(t, ta.app.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*), COALESCE(SUM(xp_awarded),0) FROM case_results`).Scan(&rows, &total))
	require.Equal(t, 1, rows)
	require.Equal(t, 160, total)
	require.Equal(t, 160, profileXP(t, ta, tok))
}

// ---- redis resilience ----

func TestRedisFlushKeepsProgress(t *testing.T) {
	ta := setupApp(t)
	// Register with a known email so we can log back in after the flush.
	status, raw := ta.do(t, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": "persistent@example.com", "username": "persistent", "password": "trenchcoat99",
	})
	require.Equal(t, 201, status, string(raw))
	tok := decode(t, raw)["token"].(string)

	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 201, status)
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/actions", tok,
		map[string]any{"action": "search", "location_id": "berth_c"})
	require.Equal(t, 200, status)

	require.NoError(t, ta.app.Cache.Flush(context.Background()))

	// Sessions live in Redis, so the old token is gone — but PG progress is
	// intact. Log back in and the evidence is still there.
	status, raw = ta.do(t, "POST", "/api/v1/auth/login", "", map[string]any{
		"email": "persistent@example.com", "password": "trenchcoat99",
	})
	require.Equal(t, 200, status, string(raw))
	tok = decode(t, raw)["token"].(string)

	status, raw = ta.do(t, "GET", "/api/v1/cases/"+caseID+"/state", tok, nil)
	require.Equal(t, 200, status, "progress must survive Redis loss (PG is source of truth)")
	require.Len(t, decode(t, raw)["discovered_evidence"].([]any), 3)

	// Leaderboard still serves from PG on cache miss.
	status, raw = ta.do(t, "GET", "/api/v1/leaderboard", tok, nil)
	require.Equal(t, 200, status, string(raw))
}

// TestCacheFailureTolerance proves every Redis consumer fails safe when the
// server is gone: misses (not errors), fail-open rate limits, and no panics
// on writes. HTTP-level PG fallback is covered by TestRedisFlushKeepsProgress.
// NOTE: session auth intentionally still 401s with Redis down — sessions live
// in Redis, and failing open there would let anyone in.
func TestCacheFailureTolerance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	bad := brokenCache(t)

	_, ok := bad.GetLeaderboard(ctx)
	require.False(t, ok, "leaderboard read must degrade to a miss")
	bad.SetLeaderboard(ctx, `[]`) // must not panic

	_, ok = bad.GetCase(ctx, caseID)
	require.False(t, ok, "case read must degrade to a miss")
	bad.SetCase(ctx, caseID, `{}`) // must not panic

	require.True(t, bad.AllowRateLimit(ctx, "login", "127.0.0.1", 1), "rate limit must fail open")

	_, ok = bad.SessionGet(ctx, "deadbeef")
	require.False(t, ok, "unknown/dead session must not authenticate")
}

// ---- leaderboard ordering ----

func TestLeaderboardOrdering(t *testing.T) {
	ta := setupApp(t)
	_, tokTop := registerUser(t, ta, "topdog")
	status, _ := ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tokTop, nil)
	require.Equal(t, 201, status)
	fullSolve(t, ta, tokTop)
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/accusations", tokTop,
		map[string]any{"suspect_id": "edwin_marsh"})
	require.Equal(t, 201, status)
	_, tokLow := registerUser(t, ta, "lowdog")

	status, raw := ta.do(t, "GET", "/api/v1/leaderboard", tokLow, nil)
	require.Equal(t, 200, status)
	var rows []map[string]any
	require.NoError(t, jsonUnmarshal(raw, &rows))
	require.GreaterOrEqual(t, len(rows), 2)
	first := rows[0]
	require.GreaterOrEqual(t, first["xp"].(float64), rows[1]["xp"].(float64))
	for _, r := range rows {
		for _, k := range []string{"username", "display_name", "level", "rank", "xp", "solved"} {
			require.Contains(t, r, k)
		}
	}
	// Cached second hit agrees.
	status, raw2 := ta.do(t, "GET", "/api/v1/leaderboard", tokLow, nil)
	require.Equal(t, 200, status)
	require.JSONEq(t, string(raw), string(raw2))
}

// ---- misc endpoints ----

func TestHealthReadyCasesAndStats(t *testing.T) {
	ta := setupApp(t)
	status, raw := ta.do(t, "GET", "/health", "", nil)
	require.Equal(t, 200, status, string(raw))
	status, raw = ta.do(t, "GET", "/ready", "", nil)
	require.Equal(t, 200, status, string(raw))
	require.Equal(t, true, decode(t, raw)["postgres"])

	_, tok := registerUser(t, ta, "statman")
	status, raw = ta.do(t, "GET", "/api/v1/cases", tok, nil)
	require.Equal(t, 200, status)
	var list []map[string]any
	require.NoError(t, jsonUnmarshal(raw, &list))
	require.Len(t, list, 2)
	require.Equal(t, "the_last_train", list[0]["id"])
	require.Equal(t, "the_tunis_ledger", list[1]["id"])
	require.Equal(t, "locked", list[1]["status"])

	status, raw = ta.do(t, "GET", "/api/v1/cases/"+caseID, tok, nil)
	require.Equal(t, 200, status)
	detail := decode(t, raw)
	require.NotContains(t, detail, "culprit_suspect_id")
	require.Len(t, detail["suspects"].([]any), 3)

	// Unknown case ids 404.
	status, _ = ta.do(t, "GET", "/api/v1/cases/nope", tok, nil)
	require.Equal(t, 404, status)
	status, _ = ta.do(t, "GET", "/api/v1/cases/nope/state", tok, nil)
	require.Equal(t, 404, status)

	// Start is idempotent: second call returns the same investigation.
	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 201, status)
	first := decode(t, raw)["investigation"].(map[string]any)["id"]
	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 200, status)
	require.Equal(t, first, decode(t, raw)["investigation"].(map[string]any)["id"])

	// Profile patch validation + stats shape.
	status, _ = ta.do(t, "PATCH", "/api/v1/player/profile", tok, map[string]any{"display_name": ""})
	require.Equal(t, 422, status)
	status, raw = ta.do(t, "PATCH", "/api/v1/player/profile", tok, map[string]any{"display_name": "Sam Spade"})
	require.Equal(t, 200, status)
	require.Equal(t, "Sam Spade", decode(t, raw)["profile"].(map[string]any)["display_name"])

	status, raw = ta.do(t, "GET", "/api/v1/player/stats", tok, nil)
	require.Equal(t, 200, status)
	stats := decode(t, raw)
	for _, k := range []string{"cases_started", "cases_solved", "cases_failed", "total_xp", "level", "rank"} {
		require.Contains(t, stats, k)
	}
	require.Equal(t, float64(1), stats["cases_started"])
}

func TestCaseLockedFlow(t *testing.T) {
	ta := setupApp(t)
	_, tok := registerUser(t, ta, "lockpick")

	// Tunis file is sealed before the train is closed.
	status, raw := ta.do(t, "GET", "/api/v1/cases", tok, nil)
	require.Equal(t, 200, status)
	var list []map[string]any
	require.NoError(t, jsonUnmarshal(raw, &list))
	byID := map[string]map[string]any{}
	for _, e := range list {
		byID[e["id"].(string)] = e
	}
	require.Equal(t, "locked", byID["the_tunis_ledger"]["status"])
	status, raw = ta.do(t, "POST", "/api/v1/cases/the_tunis_ledger/start", tok, nil)
	require.Equal(t, 409, status, string(raw))
	require.Equal(t, "case_locked", decode(t, raw)["error"].(map[string]any)["code"])
	status, _ = ta.do(t, "GET", "/api/v1/cases/the_tunis_ledger", tok, nil)
	require.Equal(t, 409, status)

	// Close the train (wrong verdict still closes the file).
	status, _ = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/start", tok, nil)
	require.Equal(t, 201, status)
	status, raw = ta.do(t, "POST", "/api/v1/cases/"+caseID+"/accusations", tok,
		map[string]any{"suspect_id": "tommy_beck"})
	require.Equal(t, 201, status, string(raw))

	// Seal lifts: start, full solve, guilty + achievement.
	status, raw = ta.do(t, "GET", "/api/v1/cases", tok, nil)
	require.Equal(t, 200, status)
	require.NoError(t, jsonUnmarshal(raw, &list))
	byID = map[string]map[string]any{}
	for _, e := range list {
		byID[e["id"].(string)] = e
	}
	require.Equal(t, "unlocked", byID["the_tunis_ledger"]["status"])
	status, _ = ta.do(t, "POST", "/api/v1/cases/the_tunis_ledger/start", tok, nil)
	require.Equal(t, 201, status)
	for _, loc := range []string{"hotel_office", "hotel_bar", "marina_quay", "cinema_etoile", "service_stair"} {
		status, _ = ta.do(t, "POST", "/api/v1/cases/the_tunis_ledger/actions", tok,
			map[string]any{"action": "search", "location_id": loc})
		require.Equal(t, 200, status)
	}
	for _, p := range [][2]string{{"bilal_haddad", "banknote"}, {"yasmine_benammar", "stub"}} {
		status, _ = ta.do(t, "POST", "/api/v1/cases/the_tunis_ledger/present-evidence", tok,
			map[string]any{"suspect_id": p[0], "evidence_id": p[1]})
		require.Equal(t, 200, status)
	}
	status, raw = ta.do(t, "POST", "/api/v1/cases/the_tunis_ledger/accusations", tok,
		map[string]any{"suspect_id": "marco_silvestri"})
	require.Equal(t, 201, status, string(raw))
	m := decode(t, raw)
	require.Equal(t, "guilty", m["result"])
	status, raw = ta.do(t, "GET", "/api/v1/player/achievements", tok, nil)
	require.Equal(t, 200, status)
	var achs []map[string]any
	require.NoError(t, jsonUnmarshal(raw, &achs))
	found := false
	for _, a := range achs {
		if a["id"] == "solved_tunis_ledger" && a["unlocked_at"] != nil {
			found = true
		}
	}
	require.True(t, found, "solved_tunis_ledger should be unlocked")
}

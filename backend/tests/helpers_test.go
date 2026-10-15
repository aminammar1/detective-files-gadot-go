package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"detectivefiles/internal/cache"
	"detectivefiles/internal/cases"
	"detectivefiles/internal/config"
	"detectivefiles/internal/handlers"
	"detectivefiles/internal/repository"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://postgres:postgres@localhost:5432/detectivefiles?sslmode=disable"
}

func testRedisAddr(t *testing.T) string {
	t.Helper()
	if v := os.Getenv("TEST_REDIS_ADDR"); v != "" {
		return v
	}
	return "localhost:6379"
}

type testApp struct {
	app *handlers.App
}

// setupApp connects to Postgres + Redis (compose services at localhost by
// default), migrates, wipes player data, and flushes Redis. If either
// dependency is unreachable the test SKIPS with a clear message — never a
// fake pass.
func setupApp(t *testing.T) *testApp {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, testDatabaseURL(t))
	if err != nil {
		t.Skipf("postgres unavailable (%v); start compose services to run integration tests", err)
	}
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("postgres unreachable (%v); start compose services to run integration tests", err)
	}

	m, err := migrate.New("file://../migrations", testDatabaseURL(t))
	if err != nil {
		t.Fatalf("migrate init: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}
	srcErr, dbErr := m.Close()
	_ = srcErr
	_ = dbErr

	if _, err := pool.Exec(ctx, `TRUNCATE users CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ch := cache.New(testRedisAddr(t), log)
	rctx, rcancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer rcancel()
	if err := ch.Ping(rctx); err != nil {
		t.Skipf("redis unreachable (%v); start compose services to run integration tests", err)
	}
	if err := ch.Flush(rctx); err != nil {
		t.Fatalf("redis flush: %v", err)
	}

	reg, err := cases.Load()
	if err != nil {
		t.Fatalf("load cases: %v", err)
	}
	cfg := config.Config{
		Port: "0", DatabaseURL: testDatabaseURL(t), RedisAddr: testRedisAddr(t),
		SessionTTLHrs: 168, RateLimitPerMin: 50,
	}
	app := &handlers.App{
		Cfg: cfg, Pool: pool, Repo: repository.New(pool),
		Cache: ch, Cases: reg, Log: log,
	}
	return &testApp{app: app}
}

func (ta *testApp) do(t *testing.T, method, path, token string, body any) (int, []byte) {
	t.Helper()
	var rdr *bytes.Reader
	if body == nil {
		rdr = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		rdr = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	ta.app.NewRouter().ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

func decode(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(raw, &m))
	return m
}

var userSeq int

func registerUser(t *testing.T, ta *testApp, username string) (userID, token string) {
	t.Helper()
	userSeq++
	email := fmt.Sprintf("%s_%d_%d@example.com", username, userSeq, time.Now().UnixNano())
	status, raw := ta.do(t, "POST", "/api/v1/auth/register", "", map[string]any{
		"email": email, "username": fmt.Sprintf("%s_%d", username, userSeq), "password": "trenchcoat99",
	})
	require.Equal(t, 201, status, string(raw))
	m := decode(t, raw)
	return m["user"].(map[string]any)["id"].(string), m["token"].(string)
}

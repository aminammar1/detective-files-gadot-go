package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache wraps Redis with failure tolerance: every read falls back to a miss
// and every write failure is logged, never fatal. Postgres stays the source
// of truth for all player progress.
type Cache struct {
	rdb *redis.Client
	log *slog.Logger
}

// New creates a Cache. Use NewWithAddr in tests to point at a bad host.
func New(addr string, log *slog.Logger) *Cache {
	return NewWithClient(redis.NewClient(&redis.Options{Addr: addr}), log)
}

// NewWithClient wraps an existing client (tests).
func NewWithClient(rdb *redis.Client, log *slog.Logger) *Cache {
	if log == nil {
		log = slog.Default()
	}
	return &Cache{rdb: rdb, log: log}
}

// Ping checks Redis health for /ready.
func (c *Cache) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// ---- Sessions: sess:<sha256hex> = user_id ----

func (c *Cache) SessionSet(ctx context.Context, tokenHash, userID string, ttl time.Duration) error {
	err := c.rdb.Set(ctx, "sess:"+tokenHash, userID, ttl).Err()
	if err != nil {
		c.log.Warn("redis session set failed", "err", err)
	}
	return err
}

// SessionSetWithTTL is SessionSet with explicit TTL (tests use short TTLs).
func (c *Cache) SessionSetWithTTL(ctx context.Context, tokenHash, userID string, ttl time.Duration) error {
	return c.SessionSet(ctx, tokenHash, userID, ttl)
}

func (c *Cache) SessionGet(ctx context.Context, tokenHash string) (string, bool) {
	v, err := c.rdb.Get(ctx, "sess:"+tokenHash).Result()
	if err != nil {
		if err != redis.Nil {
			c.log.Warn("redis session get failed", "err", err)
		}
		return "", false
	}
	return v, true
}

func (c *Cache) SessionDel(ctx context.Context, tokenHash string) {
	if err := c.rdb.Del(ctx, "sess:"+tokenHash).Err(); err != nil {
		c.log.Warn("redis session del failed", "err", err)
	}
}

// ---- Rate limits: rl:<scope>:<ip>, fixed window ----

// AllowRateLimit increments the window counter and reports whether the
// attempt is within limit. Redis failure fails open (allows) and logs.
func (c *Cache) AllowRateLimit(ctx context.Context, scope, ip string, limit int) bool {
	key := "rl:" + scope + ":" + ip
	n, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		c.log.Warn("redis rate limit failed, allowing", "err", err)
		return true
	}
	if n == 1 {
		if err := c.rdb.Expire(ctx, key, time.Minute).Err(); err != nil {
			c.log.Warn("redis rate limit expire failed", "err", err)
		}
	}
	return n <= int64(limit)
}

// ---- Case metadata cache (5 min) ----

func (c *Cache) GetCase(ctx context.Context, id string) (string, bool) {
	v, err := c.rdb.Get(ctx, "case:"+id).Result()
	if err != nil {
		if err != redis.Nil {
			c.log.Warn("redis case get failed", "err", err)
		}
		return "", false
	}
	return v, true
}

func (c *Cache) SetCase(ctx context.Context, id, payload string) {
	if err := c.rdb.Set(ctx, "case:"+id, payload, 5*time.Minute).Err(); err != nil {
		c.log.Warn("redis case set failed", "err", err)
	}
}

// ---- Leaderboard cache (60 s) ----

func (c *Cache) GetLeaderboard(ctx context.Context) (string, bool) {
	v, err := c.rdb.Get(ctx, "lb:top").Result()
	if err != nil {
		if err != redis.Nil {
			c.log.Warn("redis leaderboard get failed", "err", err)
		}
		return "", false
	}
	return v, true
}

func (c *Cache) SetLeaderboard(ctx context.Context, payload string) {
	if err := c.rdb.Set(ctx, "lb:top", payload, 60*time.Second).Err(); err != nil {
		c.log.Warn("redis leaderboard set failed", "err", err)
	}
}

// Flush clears the current DB (tests only).
func (c *Cache) Flush(ctx context.Context) error {
	return c.rdb.FlushDB(ctx).Err()
}

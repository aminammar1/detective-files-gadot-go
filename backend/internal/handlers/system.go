package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"detectivefiles/internal/middleware"
)

// GET /health and /api/v1/health
func (a *App) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /ready and /api/v1/ready
func (a *App) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	pgOK := a.Pool.Ping(ctx) == nil
	redisOK := a.Cache.Ping(ctx) == nil
	status := http.StatusOK
	if !pgOK || !redisOK {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{"ok": pgOK && redisOK, "postgres": pgOK, "redis": redisOK})
}

// GET /api/v1/leaderboard — top 20 by XP, cached 60s, PG fallback.
func (a *App) Leaderboard(c *gin.Context) {
	ctx := c.Request.Context()
	if raw, ok := a.Cache.GetLeaderboard(ctx); ok && raw != "" {
		var cached []gin.H
		if err := json.Unmarshal([]byte(raw), &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}
	rows, err := a.Repo.Leaderboard(ctx, 20)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, l := range rows {
		out = append(out, gin.H{
			"username": l.Username, "display_name": l.DisplayName,
			"level": l.Level, "rank": l.Rank, "xp": l.XP, "solved": l.Solved,
		})
	}
	if raw, err := json.Marshal(out); err == nil {
		a.Cache.SetLeaderboard(ctx, string(raw))
	}
	c.JSON(http.StatusOK, out)
}

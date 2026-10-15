package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"detectivefiles/internal/middleware"
)

func profileJSON(displayName, rank string, level, xp int) gin.H {
	return gin.H{"display_name": displayName, "rank": rank, "level": level, "xp": xp}
}

// GET /api/v1/player/profile
func (a *App) GetProfile(c *gin.Context) {
	p, err := a.Repo.GetProfile(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": profileJSON(p.DisplayName, p.Rank, p.Level, p.XP)})
}

// PATCH /api/v1/player/profile
func (a *App) PatchProfile(c *gin.Context) {
	var body struct {
		DisplayName string `json:"display_name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, 422, "validation_error", "That paperwork is incomplete.")
		return
	}
	if len(body.DisplayName) < 1 || len(body.DisplayName) > 30 {
		middleware.Error(c, 422, "validation_error", "Display name must be 1 to 30 characters.")
		return
	}
	p, err := a.Repo.UpdateDisplayName(c.Request.Context(), middleware.UserID(c), body.DisplayName)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": profileJSON(p.DisplayName, p.Rank, p.Level, p.XP)})
}

// GET /api/v1/player/stats
func (a *App) Stats(c *gin.Context) {
	ctx := c.Request.Context()
	uid := middleware.UserID(c)
	started, err := a.Repo.CountInvestigations(ctx, uid)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	t, f := true, false
	solved, err := a.Repo.CountResults(ctx, uid, &t)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	failed, err := a.Repo.CountResults(ctx, uid, &f)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	p, err := a.Repo.GetProfile(ctx, uid)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"cases_started": started,
		"cases_solved":  solved,
		"cases_failed":  failed,
		"total_xp":      p.XP,
		"level":         p.Level,
		"rank":          p.Rank,
	})
}

// GET /api/v1/player/achievements
func (a *App) Achievements(c *gin.Context) {
	list, err := a.Repo.ListAchievements(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, ach := range list {
		var unlocked any
		if ach.UnlockedAt != nil {
			unlocked = ach.UnlockedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		}
		out = append(out, gin.H{
			"id": ach.ID, "name": ach.Name, "description": ach.Description, "unlocked_at": unlocked,
		})
	}
	c.JSON(http.StatusOK, out)
}

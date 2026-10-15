package handlers

import (
	"github.com/gin-gonic/gin"

	"detectivefiles/internal/middleware"
)

// NewRouter wires every route. Health/ready live at the root AND under
// /api/v1 so both conventions work.
func (a *App) NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", a.Health)
	r.GET("/ready", a.Ready)

	v1 := r.Group("/api/v1")
	v1.GET("/health", a.Health)
	v1.GET("/ready", a.Ready)

	v1.POST("/auth/register", a.Register)
	v1.POST("/auth/login", a.Login)

	authed := v1.Group("")
	authed.Use(middleware.Auth(a.Cache))
	authed.POST("/auth/logout", a.Logout)
	authed.GET("/auth/me", a.Me)

	authed.GET("/player/profile", a.GetProfile)
	authed.PATCH("/player/profile", a.PatchProfile)
	authed.GET("/player/stats", a.Stats)
	authed.GET("/player/achievements", a.Achievements)

	authed.GET("/cases", a.ListCases)
	authed.GET("/cases/:caseId", a.GetCase)
	authed.POST("/cases/:caseId/start", a.StartCase)
	authed.GET("/cases/:caseId/state", a.CaseState)
	authed.POST("/cases/:caseId/actions", a.CaseAction)
	authed.POST("/cases/:caseId/interrogate", a.Interrogate)
	authed.POST("/cases/:caseId/present-evidence", a.PresentEvidence)
	authed.POST("/cases/:caseId/accusations", a.Accuse)

	authed.GET("/leaderboard", a.Leaderboard)
	return r
}

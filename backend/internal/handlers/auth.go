package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"detectivefiles/internal/auth"
	"detectivefiles/internal/middleware"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func userJSON(id, email, username string) gin.H {
	return gin.H{"id": id, "email": email, "username": username}
}

// POST /api/v1/auth/register
func (a *App) Register(c *gin.Context) {
	if !a.Cache.AllowRateLimit(c.Request.Context(), "register", c.ClientIP(), a.Cfg.RateLimitPerMin) {
		middleware.Error(c, http.StatusTooManyRequests, "rate_limited", "Too many attempts. Cool your heels a minute.")
		return
	}
	var body struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, 422, "validation_error", "That paperwork is incomplete.")
		return
	}
	if !validEmail(body.Email) {
		middleware.Error(c, 422, "validation_error", "That email address doesn't check out.")
		return
	}
	if !usernameRe.MatchString(body.Username) {
		middleware.Error(c, 422, "validation_error", "Username must be 3-20 letters, numbers, or underscores.")
		return
	}
	if len(body.Password) < 8 || len(body.Password) > 72 {
		middleware.Error(c, 422, "validation_error", "Password must be 8 to 72 characters.")
		return
	}
	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	u, err := a.Repo.CreateUser(c.Request.Context(), body.Email, body.Username, hash)
	if err != nil {
		if isUniqueViolation(err) {
			// Never say which field collided.
			middleware.Error(c, http.StatusConflict, "conflict", "Those credentials are already on file.")
			return
		}
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	if _, err := a.Repo.CreateProfile(c.Request.Context(), u.ID, u.Username); err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	token, err := auth.NewToken()
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	ttl := time.Duration(a.Cfg.SessionTTLHrs) * time.Hour
	if err := a.Cache.SessionSet(c.Request.Context(), auth.TokenHash(token), u.ID, ttl); err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": userJSON(u.ID, u.Email, u.Username), "token": token})
}

// POST /api/v1/auth/login
func (a *App) Login(c *gin.Context) {
	if !a.Cache.AllowRateLimit(c.Request.Context(), "login", c.ClientIP(), a.Cfg.RateLimitPerMin) {
		middleware.Error(c, http.StatusTooManyRequests, "rate_limited", "Too many attempts. Cool your heels a minute.")
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Email == "" || body.Password == "" {
		middleware.Error(c, 422, "validation_error", "That paperwork is incomplete.")
		return
	}
	const generic = "Badge number or password didn't check out."
	u, err := a.Repo.FindUserByEmail(c.Request.Context(), body.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			middleware.Error(c, http.StatusUnauthorized, "invalid_credentials", generic)
			return
		}
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	ok, err := auth.VerifyPassword(u.PasswordHash, body.Password)
	if err != nil || !ok {
		// Never reveal whether the email or the password failed.
		middleware.Error(c, http.StatusUnauthorized, "invalid_credentials", generic)
		return
	}
	token, err := auth.NewToken()
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	ttl := time.Duration(a.Cfg.SessionTTLHrs) * time.Hour
	if err := a.Cache.SessionSet(c.Request.Context(), auth.TokenHash(token), u.ID, ttl); err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": userJSON(u.ID, u.Email, u.Username), "token": token})
}

// POST /api/v1/auth/logout
func (a *App) Logout(c *gin.Context) {
	h := c.GetHeader("Authorization")
	token := ""
	if len(h) > 7 && h[:7] == "Bearer " {
		token = h[7:]
	}
	if token != "" {
		a.Cache.SessionDel(c.Request.Context(), auth.TokenHash(token))
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GET /api/v1/auth/me
func (a *App) Me(c *gin.Context) {
	uid := middleware.UserID(c)
	u, err := a.Repo.FindUserByID(c.Request.Context(), uid)
	if err != nil {
		middleware.Error(c, 404, "not_found", "No file on that badge.")
		return
	}
	p, err := a.Repo.GetProfile(c.Request.Context(), uid)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": userJSON(u.ID, u.Email, u.Username),
		"profile": gin.H{
			"display_name": p.DisplayName, "rank": p.Rank, "level": p.Level, "xp": p.XP,
		},
	})
}

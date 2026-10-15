package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"detectivefiles/internal/auth"
	"detectivefiles/internal/cache"
)

const userIDKey = "userID"

// Error writes the consistent JSON error envelope.
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

// UserID returns the authenticated user id from context.
func UserID(c *gin.Context) string {
	v, _ := c.Get(userIDKey)
	s, _ := v.(string)
	return s
}

// Auth requires `Authorization: Bearer <token>` and resolves the session
// from Redis. Missing or unknown tokens are 401; raw tokens are never logged.
func Auth(ch *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		token := ""
		if strings.HasPrefix(h, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
		}
		if token == "" {
			Error(c, 401, "unauthorized", "Show your badge first.")
			return
		}
		uid, ok := ch.SessionGet(c.Request.Context(), auth.TokenHash(token))
		if !ok || uid == "" {
			Error(c, 401, "unauthorized", "That badge doesn't check out.")
			return
		}
		c.Set(userIDKey, uid)
		c.Next()
	}
}

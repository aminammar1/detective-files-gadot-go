package handlers

import (
	"log/slog"
	"net/mail"
	"regexp"
	"strings"

	"detectivefiles/internal/cache"
	"detectivefiles/internal/cases"
	"detectivefiles/internal/config"
	"detectivefiles/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// App wires all dependencies for handlers.
type App struct {
	Cfg   config.Config
	Pool  *pgxpool.Pool
	Repo  *repository.Repo
	Cache *cache.Cache
	Cases *cases.Registry
	Log   *slog.Logger
}

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

func validEmail(e string) bool {
	if len(e) > 254 || !strings.Contains(e, "@") {
		return false
	}
	_, err := mail.ParseAddress(e)
	return err == nil
}

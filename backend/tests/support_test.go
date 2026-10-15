package tests

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"detectivefiles/internal/cache"
)

func jsonUnmarshal(raw []byte, v any) error {
	return json.Unmarshal(raw, v)
}

// brokenCache returns a Cache pointed at a port where nothing listens, so
// every Redis call fails and the failure-tolerant paths engage.
func brokenCache(t *testing.T) *cache.Cache {
	t.Helper()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return cache.New("127.0.0.1:1", log)
}

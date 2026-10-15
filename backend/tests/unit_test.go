package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"detectivefiles/internal/auth"
	"detectivefiles/internal/cases"
	"detectivefiles/internal/investigations"
	"detectivefiles/internal/services"
)

func TestPasswordHashingRoundTrip(t *testing.T) {
	h, err := auth.HashPassword("trenchcoat99")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(h, "$argon2id$v=19$"), "storable PHC format, got %q", h)
	ok, err := auth.VerifyPassword(h, "trenchcoat99")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = auth.VerifyPassword(h, "wrongpassword")
	require.NoError(t, err)
	require.False(t, ok)
	// Random salt: two hashes of the same password differ.
	h2, err := auth.HashPassword("trenchcoat99")
	require.NoError(t, err)
	require.NotEqual(t, h, h2)
	// Malformed hash is an error, never a match.
	_, err = auth.VerifyPassword("not-a-hash", "trenchcoat99")
	require.Error(t, err)
}

func TestTokenHashing(t *testing.T) {
	tok, err := auth.NewToken()
	require.NoError(t, err)
	require.Len(t, tok, 64, "32 bytes hex-encoded")
	require.Len(t, auth.TokenHash(tok), 64)
	require.NotContains(t, auth.TokenHash(tok), tok)
}

func TestRankThresholds(t *testing.T) {
	require.Equal(t, 1, services.LevelForXP(0))
	require.Equal(t, 1, services.LevelForXP(999))
	require.Equal(t, 2, services.LevelForXP(1000))
	require.Equal(t, 5, services.LevelForXP(4000))
	require.Equal(t, "Rookie", services.RankForLevel(1))
	require.Equal(t, "Gumshoe", services.RankForLevel(2))
	require.Equal(t, "Sleuth", services.RankForLevel(3))
	require.Equal(t, "Inspector", services.RankForLevel(4))
	require.Equal(t, "Chief Inspector", services.RankForLevel(5))
	require.Equal(t, "Chief Inspector", services.RankForLevel(12))
}

func TestProvenMath(t *testing.T) {
	reg, err := cases.Load()
	require.NoError(t, err)
	cf := reg.Get("the_last_train")
	require.NotNil(t, cf)
	disc := map[string]bool{}
	for _, id := range cf.AccusationRules.RequiredEvidence {
		disc[id] = true
	}
	unlocked := map[string]string{}
	for _, id := range cf.AccusationRules.RequiredTestimonies {
		unlocked[id] = "2026-01-01T00:00:00Z"
	}
	require.True(t, investigations.Proven(cf, cf.CulpritSuspectID, disc, unlocked))
	require.False(t, investigations.Proven(cf, "vivienne_hart", disc, unlocked))
	delete(disc, cf.AccusationRules.RequiredEvidence[0])
	require.False(t, investigations.Proven(cf, cf.CulpritSuspectID, disc, unlocked))
}

func TestDeflectionsDistinctAndOriginal(t *testing.T) {
	c, err := cases.Load()
	require.NoError(t, err)
	cf := c.Get("the_last_train")
	require.NotNil(t, cf)
	seen := map[string]bool{}
	for _, s := range cf.Suspects {
		d := services.Deflection(s.ID)
		require.NotEmpty(t, d)
		require.False(t, seen[d], "deflections must be distinct per suspect")
		seen[d] = true
		// Must not reuse case copy: no topic or contradiction response may
		// contain the deflection or vice versa.
		for _, topic := range s.Topics {
			require.NotContains(t, topic.Response, d)
			require.NotContains(t, d, topic.Response)
		}
		for _, con := range s.Contradictions {
			require.NotContains(t, con.Response, d)
			require.NotContains(t, d, con.Response)
		}
	}
}

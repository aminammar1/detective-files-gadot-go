package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"detectivefiles/internal/cache"
	"detectivefiles/internal/cases"
	"detectivefiles/internal/investigations"
	"detectivefiles/internal/models"
	"detectivefiles/internal/repository"
)

// ---- ranks ----

// LevelForXP: level = 1 + xp/1000 (integer division).
func LevelForXP(xp int) int {
	if xp < 0 {
		xp = 0
	}
	return 1 + xp/1000
}

// RankForLevel: L1 Rookie, L2 Gumshoe, L3 Sleuth, L4 Inspector, L5+ Chief Inspector.
func RankForLevel(level int) string {
	switch {
	case level <= 1:
		return "Rookie"
	case level == 2:
		return "Gumshoe"
	case level == 3:
		return "Sleuth"
	case level == 4:
		return "Inspector"
	default:
		return "Chief Inspector"
	}
}

// ---- deflections (original copy; never reuse case text) ----

// Deflection returns an in-voice brush-off for presenting evidence that
// triggers no contradiction. One distinct voice per suspect.
func Deflection(suspectID string) string {
	switch suspectID {
	case "edwin_marsh":
		return "Cut the theater, detective. That proves nothing about me or my night. Bring paper that matters, or bill your hours elsewhere."
	case "vivienne_hart":
		return "Oh, sugar, you wave that around like a stage prop — cute, but it don't sing a single note about me."
	case "tommy_beck":
		return "I — I never seen that, sir, I swear, sir, please don't write down that I touched it, sir!"
	default:
		return "That doesn't prove a thing about me, detective. Try again with something real."
	}
}

// ---- accusation outcome ----

type AccusationOutcome struct {
	SuspectID string `json:"suspect_id"`
	CaseID    string `json:"case_id"`
	Correct   bool   `json:"correct"`
	Proven    bool   `json:"proven"`
	Score     int    `json:"score"`
	XPAwarded int    `json:"xp_awarded"`
	Completed bool   `json:"completed"`
	Result    string `json:"result"` // guilty | unproven | wrong
}

// unlockedTestimonies derives unlocked testimony ids (with unlock time) from
// present-events: an event unlocks T when its (suspect,evidence) matches a
// contradiction whose unlocks_testimony is T.
func unlockedTestimonies(c *models.CaseFile, events []repository.InterrogationEvent) map[string]string {
	byEvidence := map[string]map[string]*string{} // suspect -> evidence -> testimony
	for i := range c.Suspects {
		s := &c.Suspects[i]
		m := map[string]*string{}
		for j := range s.Contradictions {
			con := &s.Contradictions[j]
			m[con.EvidenceID] = con.UnlocksTestimony
		}
		byEvidence[s.ID] = m
	}
	out := map[string]string{}
	for _, e := range events {
		if e.Kind != "present" || e.EvidenceID == nil {
			continue
		}
		if tid := byEvidence[e.SuspectID][*e.EvidenceID]; tid != nil && *tid != "" {
			ts := e.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
			if prev, ok := out[*tid]; !ok || ts < prev {
				out[*tid] = ts
			}
		}
	}
	return out
}

// UnlockedTestimonies is the exported helper for state responses.
func UnlockedTestimonies(c *models.CaseFile, events []repository.InterrogationEvent) map[string]string {
	return unlockedTestimonies(c, events)
}

// Accuse runs the accusation inside a single DB transaction. An existing
// result row makes it idempotent: the stored row returns with no new XP.
// A PK conflict on insert (concurrent double-accuse) rolls back and returns
// the winner's row, so XP is awarded exactly once.
func Accuse(ctx context.Context, pool *pgxpool.Pool, repo *repository.Repo, c *models.CaseFile, userID, caseID, suspectID string, log *slog.Logger) (*AccusationOutcome, bool, error) {
	// Fast path: already judged.
	if res, err := repo.FindResult(ctx, userID, caseID); err == nil {
		return outcomeFromRow(res), false, nil
	}

	discovered, err := repo.ListDiscovered(ctx, mustInvestigationID(ctx, pool, userID, caseID))
	if err != nil {
		return nil, false, err
	}
	// Need investigation id for events; fetch investigation directly.
	iv, err := repo.FindInvestigation(ctx, userID, caseID)
	if err != nil {
		return nil, false, err
	}
	events, err := repo.ListEvents(ctx, iv.ID)
	if err != nil {
		return nil, false, err
	}

	discSet := map[string]bool{}
	for _, d := range discovered {
		discSet[d.EvidenceID] = true
	}
	unlocked := unlockedTestimonies(c, events)

	correct := suspectID == c.CulpritSuspectID
	proven := investigations.Proven(c, suspectID, discSet, unlocked)

	var score, xp int
	var result string
	switch {
	case correct && proven:
		score = c.Scoring.BaseCorrect + c.Scoring.PerEvidence*len(discSet) + c.Scoring.PerTestimony*len(unlocked)
		xp = score / 10
		result = "guilty"
	case correct:
		// Right person, unproven: consolations are XP values per the case
		// doc; score mirrors xp*10 for consistency.
		xp = c.Scoring.ConsolationUnproven
		score = xp * 10
		result = "unproven"
	default:
		xp = c.Scoring.ConsolationWrong
		score = xp * 10
		result = "wrong"
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM case_results WHERE user_id=$1 AND case_id=$2)`, userID, caseID).Scan(&exists)
	if err != nil {
		return nil, false, err
	}
	if exists {
		// Lost a race or double-submit: return the stored row.
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Warn("accuse rollback failed", "err", err)
		}
		res, err := repo.FindResult(ctx, userID, caseID)
		if err != nil {
			return nil, false, err
		}
		return outcomeFromRow(res), false, nil
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO case_results(user_id,case_id,suspect_id,correct,proven,score,xp_awarded)
		 VALUES($1,$2,$3,$4,$5,$6,$7)`, userID, caseID, suspectID, correct, proven, score, xp); err != nil {
		// Concurrent insert won the race: fall back to its row.
		_ = tx.Rollback(ctx)
		if res, rerr := repo.FindResult(ctx, userID, caseID); rerr == nil {
			return outcomeFromRow(res), false, nil
		}
		return nil, false, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE investigations SET status='completed',completed_at=now()
		 WHERE user_id=$1 AND case_id=$2`, userID, caseID); err != nil {
		return nil, false, err
	}
	// XP + level/rank recompute. Read-then-write inside the same tx.
	var curXP int
	if err := tx.QueryRow(ctx, `SELECT xp FROM player_profiles WHERE user_id=$1`, userID).Scan(&curXP); err != nil {
		return nil, false, err
	}
	newXP := curXP + xp
	newLevel := LevelForXP(newXP)
	if _, err := tx.Exec(ctx, `UPDATE player_profiles SET xp=$2,level=$3,rank=$4,updated_at=now() WHERE user_id=$1`,
		userID, newXP, newLevel, RankForLevel(newLevel)); err != nil {
		return nil, false, err
	}
	if correct && proven {
		// Per-case solved achievement: "the_last_train" -> "solved_last_train",
		// "the_tunis_ledger" -> "solved_tunis_ledger". Awarded only if the id
		// is seeded — the FK-safe conditional keeps future cases from breaking.
		solvedID := "solved_" + strings.TrimPrefix(caseID, "the_")
		if _, err := tx.Exec(ctx,
			`INSERT INTO player_achievements(user_id,achievement_id)
			 SELECT $1,$2 WHERE EXISTS (SELECT 1 FROM achievements WHERE id=$2)
			 ON CONFLICT DO NOTHING`,
			userID, solvedID); err != nil {
			return nil, false, err
		}
		if caseID == "the_last_train" && len(discSet) == len(c.Evidence) && len(unlocked) == len(c.Testimonies) {
			if _, err := tx.Exec(ctx,
				`INSERT INTO player_achievements(user_id,achievement_id) VALUES($1,'perfect_deduction') ON CONFLICT DO NOTHING`,
				userID); err != nil {
				return nil, false, err
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}
	return &AccusationOutcome{
		SuspectID: suspectID, CaseID: caseID, Correct: correct, Proven: proven,
		Score: score, XPAwarded: xp, Completed: true, Result: result,
	}, true, nil
}

func outcomeFromRow(r *repository.CaseResult) *AccusationOutcome {
	result := "wrong"
	switch {
	case r.Correct && r.Proven:
		result = "guilty"
	case r.Correct:
		result = "unproven"
	}
	return &AccusationOutcome{
		SuspectID: r.SuspectID, CaseID: r.CaseID, Correct: r.Correct, Proven: r.Proven,
		Score: r.Score, XPAwarded: r.XPAwarded, Completed: true, Result: result,
	}
}

func mustInvestigationID(ctx context.Context, pool *pgxpool.Pool, userID, caseID string) string {
	var id string
	_ = pool.QueryRow(ctx, `SELECT id FROM investigations WHERE user_id=$1 AND case_id=$2`, userID, caseID).Scan(&id)
	return id
}

// ---- case metadata cache (5 min, Redis failure tolerant) ----

// CachedDetailJSON returns the sanitized case detail JSON, caching the
// encoded payload in Redis for 5 minutes. A cache miss or Redis failure
// falls back to re-encoding the embedded registry copy.
func CachedDetailJSON(ctx context.Context, ch *cache.Cache, reg *cases.Registry, id string) ([]byte, error) {
	if raw, ok := ch.GetCase(ctx, id); ok && raw != "" {
		return []byte(raw), nil
	}
	c := reg.Get(id)
	if c == nil {
		return nil, fmt.Errorf("unknown case %q", id)
	}
	raw, err := json.Marshal(cases.Sanitize(c))
	if err != nil {
		return nil, err
	}
	ch.SetCase(ctx, id, string(raw))
	return raw, nil
}

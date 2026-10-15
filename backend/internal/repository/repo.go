package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Row types.
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Profile struct {
	UserID      string
	DisplayName string
	XP          int
	Level       int
	Rank        string
	UpdatedAt   time.Time
}

type Investigation struct {
	ID          string
	UserID      string
	CaseID      string
	Status      string
	StartedAt   time.Time
	CompletedAt *time.Time
}

type DiscoveredRow struct {
	EvidenceID   string
	DiscoveredAt time.Time
}

type InterrogationEvent struct {
	ID         string
	SuspectID  string
	TopicID    *string
	EvidenceID *string
	Kind       string
	Response   string
	CreatedAt  time.Time
}

type CaseResult struct {
	UserID      string
	CaseID      string
	SuspectID   string
	Correct     bool
	Proven      bool
	Score       int
	XPAwarded   int
	CompletedAt time.Time
}

type Achievement struct {
	ID          string
	Name        string
	Description string
	UnlockedAt  *time.Time
}

type LeaderRow struct {
	Username    string
	DisplayName string
	Level       int
	Rank        string
	XP          int
	Solved      int
}

type Repo struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo { return &Repo{Pool: pool} }

// ---- users / profiles ----

func (r *Repo) CreateUser(ctx context.Context, email, username, hash string) (*User, error) {
	u := &User{}
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO users(email,username,password_hash) VALUES($1,$2,$3)
		 RETURNING id,email,username,password_hash,created_at`,
		email, username, hash).Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (r *Repo) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := r.Pool.QueryRow(ctx,
		`SELECT id,email,username,password_hash,created_at FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (r *Repo) FindUserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := r.Pool.QueryRow(ctx,
		`SELECT id,email,username,password_hash,created_at FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func (r *Repo) CreateProfile(ctx context.Context, userID, displayName string) (*Profile, error) {
	p := &Profile{}
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO player_profiles(user_id,display_name) VALUES($1,$2)
		 RETURNING user_id,display_name,xp,level,rank,updated_at`,
		userID, displayName).Scan(&p.UserID, &p.DisplayName, &p.XP, &p.Level, &p.Rank, &p.UpdatedAt)
	return p, err
}

func (r *Repo) GetProfile(ctx context.Context, userID string) (*Profile, error) {
	p := &Profile{}
	err := r.Pool.QueryRow(ctx,
		`SELECT user_id,display_name,xp,level,rank,updated_at FROM player_profiles WHERE user_id=$1`, userID).
		Scan(&p.UserID, &p.DisplayName, &p.XP, &p.Level, &p.Rank, &p.UpdatedAt)
	return p, err
}

func (r *Repo) UpdateDisplayName(ctx context.Context, userID, name string) (*Profile, error) {
	p := &Profile{}
	err := r.Pool.QueryRow(ctx,
		`UPDATE player_profiles SET display_name=$2,updated_at=now() WHERE user_id=$1
		 RETURNING user_id,display_name,xp,level,rank,updated_at`,
		userID, name).Scan(&p.UserID, &p.DisplayName, &p.XP, &p.Level, &p.Rank, &p.UpdatedAt)
	return p, err
}

// AddXP adds XP and recomputes level/rank atomically.
func (r *Repo) AddXP(ctx context.Context, tx pgx.Tx, userID string, xp int, level int, rank string) error {
	var q pgx.Row
	if tx != nil {
		q = tx.QueryRow(ctx, `UPDATE player_profiles SET xp=xp+$2,level=$3,rank=$4,updated_at=now()
			WHERE user_id=$1 RETURNING xp`, userID, xp, level, rank)
	} else {
		q = r.Pool.QueryRow(ctx, `UPDATE player_profiles SET xp=xp+$2,level=$3,rank=$4,updated_at=now()
			WHERE user_id=$1 RETURNING xp`, userID, xp, level, rank)
	}
	var ignored int
	return q.Scan(&ignored)
}

// ---- investigations ----

func (r *Repo) FindInvestigation(ctx context.Context, userID, caseID string) (*Investigation, error) {
	iv := &Investigation{}
	err := r.Pool.QueryRow(ctx,
		`SELECT id,user_id,case_id,status,started_at,completed_at FROM investigations
		 WHERE user_id=$1 AND case_id=$2`, userID, caseID).
		Scan(&iv.ID, &iv.UserID, &iv.CaseID, &iv.Status, &iv.StartedAt, &iv.CompletedAt)
	return iv, err
}

func (r *Repo) CreateInvestigation(ctx context.Context, userID, caseID string) (*Investigation, error) {
	iv := &Investigation{}
	err := r.Pool.QueryRow(ctx,
		`INSERT INTO investigations(user_id,case_id) VALUES($1,$2)
		 RETURNING id,user_id,case_id,status,started_at,completed_at`,
		userID, caseID).Scan(&iv.ID, &iv.UserID, &iv.CaseID, &iv.Status, &iv.StartedAt, &iv.CompletedAt)
	return iv, err
}

func (r *Repo) CountInvestigations(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM investigations WHERE user_id=$1`, userID).Scan(&n)
	return n, err
}

// ---- discovered evidence ----

func (r *Repo) InsertDiscovered(ctx context.Context, investigationID string, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.Pool.Query(ctx,
		`INSERT INTO discovered_evidence(investigation_id,evidence_id)
		 SELECT $1, unnest($2::text[]) ON CONFLICT DO NOTHING RETURNING evidence_id`, investigationID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var inserted []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		inserted = append(inserted, id)
	}
	return inserted, rows.Err()
}

func (r *Repo) ListDiscovered(ctx context.Context, investigationID string) ([]DiscoveredRow, error) {
	rows, err := r.Pool.Query(ctx,
		`SELECT evidence_id,discovered_at FROM discovered_evidence WHERE investigation_id=$1 ORDER BY discovered_at`,
		investigationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DiscoveredRow
	for rows.Next() {
		var d DiscoveredRow
		if err := rows.Scan(&d.EvidenceID, &d.DiscoveredAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Repo) CountDiscovered(ctx context.Context, investigationID string) (int, error) {
	var n int
	err := r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM discovered_evidence WHERE investigation_id=$1`, investigationID).Scan(&n)
	return n, err
}

// ---- events / actions ----

func (r *Repo) InsertEvent(ctx context.Context, investigationID, suspectID string, topicID, evidenceID *string, kind, response string) error {
	_, err := r.Pool.Exec(ctx,
		`INSERT INTO interrogation_events(investigation_id,suspect_id,topic_id,evidence_id,kind,response_text)
		 VALUES($1,$2,$3,$4,$5,$6)`, investigationID, suspectID, topicID, evidenceID, kind, response)
	return err
}

func (r *Repo) ListEvents(ctx context.Context, investigationID string) ([]InterrogationEvent, error) {
	rows, err := r.Pool.Query(ctx,
		`SELECT id,suspect_id,topic_id,evidence_id,kind,response_text,created_at
		 FROM interrogation_events WHERE investigation_id=$1 ORDER BY created_at`, investigationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InterrogationEvent
	for rows.Next() {
		var e InterrogationEvent
		if err := rows.Scan(&e.ID, &e.SuspectID, &e.TopicID, &e.EvidenceID, &e.Kind, &e.Response, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repo) CountEvents(ctx context.Context, investigationID, kind string) (int, error) {
	var n int
	var err error
	if kind == "" {
		err = r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM interrogation_events WHERE investigation_id=$1`, investigationID).Scan(&n)
	} else {
		err = r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM interrogation_events WHERE investigation_id=$1 AND kind=$2`, investigationID, kind).Scan(&n)
	}
	return n, err
}

func (r *Repo) LogAction(ctx context.Context, investigationID, action, payload string) error {
	_, err := r.Pool.Exec(ctx,
		`INSERT INTO investigation_actions(investigation_id,action,payload) VALUES($1,$2,$3::jsonb)`,
		investigationID, action, payload)
	return err
}

func (r *Repo) CountActions(ctx context.Context, investigationID string) (int, error) {
	var n int
	err := r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM investigation_actions WHERE investigation_id=$1`, investigationID).Scan(&n)
	return n, err
}

// ---- results (accusations run in a single transaction in services) ----

func (r *Repo) FindResult(ctx context.Context, userID, caseID string) (*CaseResult, error) {
	res := &CaseResult{}
	err := r.Pool.QueryRow(ctx,
		`SELECT user_id,case_id,suspect_id,correct,proven,score,xp_awarded,completed_at
		 FROM case_results WHERE user_id=$1 AND case_id=$2`, userID, caseID).
		Scan(&res.UserID, &res.CaseID, &res.SuspectID, &res.Correct, &res.Proven, &res.Score, &res.XPAwarded, &res.CompletedAt)
	return res, err
}

func (r *Repo) CountResults(ctx context.Context, userID string, correct *bool) (int, error) {
	var n int
	var err error
	if correct == nil {
		err = r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM case_results WHERE user_id=$1`, userID).Scan(&n)
	} else {
		err = r.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM case_results WHERE user_id=$1 AND correct=$2`, userID, *correct).Scan(&n)
	}
	return n, err
}

// ---- achievements ----

func (r *Repo) Award(ctx context.Context, userID, achievementID string) error {
	_, err := r.Pool.Exec(ctx,
		`INSERT INTO player_achievements(user_id,achievement_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,
		userID, achievementID)
	return err
}

func (r *Repo) ListAchievements(ctx context.Context, userID string) ([]Achievement, error) {
	rows, err := r.Pool.Query(ctx,
		`SELECT a.id,a.name,a.description,pa.unlocked_at
		 FROM achievements a LEFT JOIN player_achievements pa
		   ON pa.achievement_id=a.id AND pa.user_id=$1 ORDER BY a.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Achievement
	for rows.Next() {
		var a Achievement
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.UnlockedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---- leaderboard ----

func (r *Repo) Leaderboard(ctx context.Context, limit int) ([]LeaderRow, error) {
	rows, err := r.Pool.Query(ctx,
		`SELECT u.username,p.display_name,p.level,p.rank,p.xp,
		        COALESCE((SELECT COUNT(*) FROM case_results r WHERE r.user_id=u.id AND r.correct),0) AS solved
		 FROM users u JOIN player_profiles p ON p.user_id=u.id
		 ORDER BY p.xp DESC, u.created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LeaderRow
	for rows.Next() {
		var l LeaderRow
		if err := rows.Scan(&l.Username, &l.DisplayName, &l.Level, &l.Rank, &l.XP, &l.Solved); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

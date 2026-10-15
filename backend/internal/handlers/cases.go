package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"detectivefiles/internal/middleware"
	"detectivefiles/internal/services"
)

// GET /api/v1/cases
func (a *App) ListCases(c *gin.Context) {
	ctx := c.Request.Context()
	uid := middleware.UserID(c)
	out := []gin.H{}
	for _, id := range a.Cases.IDs() {
		cf := a.Cases.Get(id)
		entry := gin.H{
			"id": id, "title": cf.Title, "difficulty": cf.Difficulty,
			"status": "unlocked", "started_at": nil, "completed_at": nil, "correct": nil,
		}
		if cf.RequiresCase != "" {
			entry["status"] = "locked"
			entry["requires_case"] = cf.RequiresCase
		}
		if iv, err := a.Repo.FindInvestigation(ctx, uid, id); err == nil {
			entry["started_at"] = iv.StartedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		} else if !errors.Is(err, pgx.ErrNoRows) {
			middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
			return
		}
		if res, err := a.Repo.FindResult(ctx, uid, id); err == nil {
			entry["status"] = "completed"
			entry["completed_at"] = res.CompletedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
			entry["correct"] = res.Correct
		} else if !errors.Is(err, pgx.ErrNoRows) {
			middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
			return
		} else if cf.RequiresCase == "" {
			entry["status"] = "unlocked"
		} else if _, rerr := a.Repo.FindResult(ctx, uid, cf.RequiresCase); rerr == nil {
			entry["status"] = "unlocked"
		} else if !errors.Is(rerr, pgx.ErrNoRows) {
			middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
			return
		}
		out = append(out, entry)
	}
	c.JSON(http.StatusOK, out)
}

// GET /api/v1/cases/:caseId — spoiler-free.
func (a *App) GetCase(c *gin.Context) {
	cf := a.requireCase(c)
	if cf == nil {
		return
	}
	if !a.requireUnlocked(c, cf) {
		return
	}
	raw, err := services.CachedDetailJSON(c.Request.Context(), a.Cache, a.Cases, cf.ID)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	c.Data(http.StatusOK, "application/json", raw)
}

// POST /api/v1/cases/:caseId/start — idempotent.
func (a *App) StartCase(c *gin.Context) {
	cf := a.requireCase(c)
	if cf == nil {
		return
	}
	if !a.requireUnlocked(c, cf) {
		return
	}
	ctx := c.Request.Context()
	uid := middleware.UserID(c)
	if iv, err := a.Repo.FindInvestigation(ctx, uid, cf.ID); err == nil {
		c.JSON(http.StatusOK, gin.H{"investigation": investigationJSON(iv)})
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	iv, err := a.Repo.CreateInvestigation(ctx, uid, cf.ID)
	if err != nil {
		// Concurrent start won the race: return the existing row.
		if existing, rerr := a.Repo.FindInvestigation(ctx, uid, cf.ID); rerr == nil {
			c.JSON(http.StatusOK, gin.H{"investigation": investigationJSON(existing)})
			return
		}
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	if err := a.Repo.LogAction(ctx, iv.ID, "start", `{"case_id":`+quote(cf.ID)+`}`); err != nil {
		a.Log.Warn("log action failed", "err", err)
	}
	if err := a.Repo.Award(ctx, uid, "started_first_case"); err != nil {
		a.Log.Warn("award failed", "err", err)
	}
	c.JSON(http.StatusCreated, gin.H{"investigation": investigationJSON(iv)})
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// GET /api/v1/cases/:caseId/state
func (a *App) CaseState(c *gin.Context) {
	cf := a.requireCase(c)
	if cf == nil {
		return
	}
	iv := a.requireInvestigation(c, cf.ID)
	if iv == nil {
		return
	}
	ctx := c.Request.Context()
	disc, err := a.Repo.ListDiscovered(ctx, iv.ID)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	events, err := a.Repo.ListEvents(ctx, iv.ID)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	evOut := []gin.H{}
	for _, d := range disc {
		if e := evidenceByID(cf, d.EvidenceID); e != nil {
			evOut = append(evOut, gin.H{
				"id": e.ID, "name": e.Name, "location_id": e.LocationID,
				"description": e.Description, "proves": e.Proves, "asset": e.Asset,
				"discovered_at": d.DiscoveredAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
			})
		}
	}
	if evOut == nil {
		evOut = []gin.H{}
	}
	unlocked := services.UnlockedTestimonies(cf, events)
	testOut := []gin.H{}
	for _, t := range cf.Testimonies {
		if ts, ok := unlocked[t.ID]; ok {
			testOut = append(testOut, gin.H{
				"id": t.ID, "title": t.Title, "text": t.Text, "unlocked_at": ts,
			})
		}
	}
	if testOut == nil {
		testOut = []gin.H{}
	}
	actions, err := a.Repo.CountActions(ctx, iv.ID)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	interrogations, err := a.Repo.CountEvents(ctx, iv.ID, "")
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"investigation":       investigationJSON(iv),
		"discovered_evidence": evOut,
		"testimonies":         testOut,
		"stats": gin.H{
			"evidence_found": len(evOut), "actions": actions, "interrogations": interrogations,
		},
	})
}

// POST /api/v1/cases/:caseId/actions
func (a *App) CaseAction(c *gin.Context) {
	cf := a.requireCase(c)
	if cf == nil {
		return
	}
	iv := a.requireInvestigation(c, cf.ID)
	if iv == nil || !guardOpen(c, iv) {
		return
	}
	var body struct {
		Action     string `json:"action"`
		LocationID string `json:"location_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, 422, "validation_error", "That paperwork is incomplete.")
		return
	}
	if body.Action != "search" {
		middleware.Error(c, 422, "validation_error", "Unknown action. The only move here is to search.")
		return
	}
	known := false
	for _, l := range cf.Locations {
		if l.ID == body.LocationID {
			known = true
			break
		}
	}
	if !known {
		middleware.Error(c, 422, "validation_error", "No such car on this train.")
		return
	}
	ctx := c.Request.Context()
	disc, err := a.Repo.ListDiscovered(ctx, iv.ID)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	have := map[string]bool{}
	for _, d := range disc {
		have[d.EvidenceID] = true
	}
	var candidates []string
	for _, e := range cf.Evidence {
		if e.LocationID == body.LocationID && !have[e.ID] {
			candidates = append(candidates, e.ID)
		}
	}
	inserted, err := a.Repo.InsertDiscovered(ctx, iv.ID, candidates)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	if len(inserted) > 0 && len(disc) == 0 {
		if err := a.Repo.Award(ctx, middleware.UserID(c), "found_first_evidence"); err != nil {
			a.Log.Warn("award failed", "err", err)
		}
	}
	newSet := map[string]bool{}
	for _, id := range inserted {
		newSet[id] = true
	}
	found := []gin.H{}
	already := []string{}
	for _, e := range cf.Evidence {
		if e.LocationID != body.LocationID {
			continue
		}
		if newSet[e.ID] {
			found = append(found, gin.H{
				"id": e.ID, "name": e.Name, "location_id": e.LocationID,
				"description": e.Description, "proves": e.Proves, "asset": e.Asset,
			})
		} else if have[e.ID] {
			already = append(already, e.ID)
		}
	}
	if found == nil {
		found = []gin.H{}
	}
	if already == nil {
		already = []string{}
	}
	_ = a.Repo.LogAction(ctx, iv.ID, "search", `{"location_id":`+quote(body.LocationID)+`}`)
	c.JSON(http.StatusOK, gin.H{"found": found, "already_known": already})
}

// POST /api/v1/cases/:caseId/interrogate
func (a *App) Interrogate(c *gin.Context) {
	cf := a.requireCase(c)
	if cf == nil {
		return
	}
	iv := a.requireInvestigation(c, cf.ID)
	if iv == nil || !guardOpen(c, iv) {
		return
	}
	var body struct {
		SuspectID string `json:"suspect_id"`
		TopicID   string `json:"topic_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, 422, "validation_error", "That paperwork is incomplete.")
		return
	}
	s := suspectByID(cf, body.SuspectID)
	if s == nil {
		middleware.Error(c, 422, "validation_error", "That person isn't on this train.")
		return
	}
	t := topicByID(s, body.TopicID)
	if t == nil {
		middleware.Error(c, 422, "validation_error", "That line of questioning goes nowhere.")
		return
	}
	ctx := c.Request.Context()
	if err := a.Repo.InsertEvent(ctx, iv.ID, s.ID, &t.ID, nil, "ask", t.Response); err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	_ = a.Repo.LogAction(ctx, iv.ID, "interrogate", `{"suspect_id":`+quote(s.ID)+`,"topic_id":`+quote(t.ID)+`}`)
	c.JSON(http.StatusOK, gin.H{"suspect_id": s.ID, "topic_id": t.ID, "response": t.Response})
}

// POST /api/v1/cases/:caseId/present-evidence
func (a *App) PresentEvidence(c *gin.Context) {
	cf := a.requireCase(c)
	if cf == nil {
		return
	}
	iv := a.requireInvestigation(c, cf.ID)
	if iv == nil || !guardOpen(c, iv) {
		return
	}
	var body struct {
		SuspectID  string `json:"suspect_id"`
		EvidenceID string `json:"evidence_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, 422, "validation_error", "That paperwork is incomplete.")
		return
	}
	s := suspectByID(cf, body.SuspectID)
	if s == nil {
		middleware.Error(c, 422, "validation_error", "That person isn't on this train.")
		return
	}
	if evidenceByID(cf, body.EvidenceID) == nil {
		middleware.Error(c, 422, "validation_error", "That exhibit isn't in the case file.")
		return
	}
	ctx := c.Request.Context()
	uid := middleware.UserID(c)
	disc, err := a.Repo.ListDiscovered(ctx, iv.ID)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	has := false
	for _, d := range disc {
		if d.EvidenceID == body.EvidenceID {
			has = true
			break
		}
	}
	if !has {
		middleware.Error(c, 422, "unproven", "You don't have that in hand.")
		return
	}
	for _, con := range s.Contradictions {
		if con.EvidenceID != body.EvidenceID {
			continue
		}
		eid := body.EvidenceID
		if err := a.Repo.InsertEvent(ctx, iv.ID, s.ID, nil, &eid, "present", con.Response); err != nil {
			middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
			return
		}
		_ = a.Repo.LogAction(ctx, iv.ID, "present", `{"suspect_id":`+quote(s.ID)+`,"evidence_id":`+quote(eid)+`}`)
		var unlocked any
		if con.UnlocksTestimony != nil && *con.UnlocksTestimony != "" {
			if t := testimonyByID(cf, *con.UnlocksTestimony); t != nil {
				unlocked = gin.H{"id": t.ID, "title": t.Title, "text": t.Text}
				if err := a.Repo.Award(ctx, uid, "broke_alibi"); err != nil {
					a.Log.Warn("award failed", "err", err)
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"contradiction": true, "response": con.Response, "unlocked_testimony": unlocked,
		})
		return
	}
	// No contradiction: in-voice deflection, original copy.
	deflect := services.Deflection(s.ID)
	eid := body.EvidenceID
	if err := a.Repo.InsertEvent(ctx, iv.ID, s.ID, nil, &eid, "present", deflect); err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	_ = a.Repo.LogAction(ctx, iv.ID, "present", `{"suspect_id":`+quote(s.ID)+`,"evidence_id":`+quote(eid)+`}`)
	c.JSON(http.StatusOK, gin.H{
		"contradiction": false, "response": deflect, "unlocked_testimony": nil,
	})
}

// POST /api/v1/cases/:caseId/accusations
func (a *App) Accuse(c *gin.Context) {
	cf := a.requireCase(c)
	if cf == nil {
		return
	}
	iv := a.requireInvestigation(c, cf.ID)
	if iv == nil {
		return
	}
	var body struct {
		SuspectID string `json:"suspect_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Error(c, 422, "validation_error", "That paperwork is incomplete.")
		return
	}
	if suspectByID(cf, body.SuspectID) == nil {
		middleware.Error(c, 422, "validation_error", "That person isn't on this train.")
		return
	}
	out, created, err := services.Accuse(c.Request.Context(), a.Pool, a.Repo, cf, middleware.UserID(c), cf.ID, body.SuspectID, a.Log)
	if err != nil {
		middleware.Error(c, 500, "internal", "The precinct is having trouble. Try again.")
		return
	}
	if created {
		c.JSON(http.StatusCreated, out)
		return
	}
	c.JSON(http.StatusOK, out)
}

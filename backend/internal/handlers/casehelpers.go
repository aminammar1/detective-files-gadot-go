package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"detectivefiles/internal/middleware"
	"detectivefiles/internal/models"
	"detectivefiles/internal/repository"
)

func (a *App) requireCase(c *gin.Context) *models.CaseFile {
	cf := a.Cases.Get(c.Param("caseId"))
	if cf == nil {
		middleware.Error(c, http.StatusNotFound, "not_found", "No such case file.")
		return nil
	}
	return cf
}

// requireUnlocked rejects sealed follow-up cases: a case with requires_case
// stays sealed until the caller has closed the previous file (any verdict).
func (a *App) requireUnlocked(c *gin.Context, cf *models.CaseFile) bool {
	if cf.RequiresCase == "" {
		return true
	}
	if _, err := a.Repo.FindResult(c.Request.Context(), middleware.UserID(c), cf.RequiresCase); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			middleware.Error(c, http.StatusConflict, "case_locked", "That file is sealed. Close the previous case first.")
			return false
		}
		middleware.Error(c, http.StatusInternalServerError, "internal", "The precinct is having trouble. Try again.")
		return false
	}
	return true
}

// requireInvestigation loads the caller's investigation or writes 404/500.
func (a *App) requireInvestigation(c *gin.Context, caseID string) *repository.Investigation {
	iv, err := a.Repo.FindInvestigation(c.Request.Context(), middleware.UserID(c), caseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			middleware.Error(c, http.StatusNotFound, "not_found", "You haven't opened that case yet.")
			return nil
		}
		middleware.Error(c, http.StatusInternalServerError, "internal", "The precinct is having trouble. Try again.")
		return nil
	}
	return iv
}

// guardOpen rejects writes against a completed (closed) investigation.
func guardOpen(c *gin.Context, iv *repository.Investigation) bool {
	if iv.Status == "completed" {
		middleware.Error(c, http.StatusConflict, "case_closed", "That case is closed. The file is sealed.")
		return false
	}
	return true
}

func suspectByID(cf *models.CaseFile, id string) *models.Suspect {
	for i := range cf.Suspects {
		if cf.Suspects[i].ID == id {
			return &cf.Suspects[i]
		}
	}
	return nil
}

func topicByID(s *models.Suspect, id string) *models.Topic {
	for i := range s.Topics {
		if s.Topics[i].ID == id {
			return &s.Topics[i]
		}
	}
	return nil
}

func evidenceByID(cf *models.CaseFile, id string) *models.Evidence {
	for i := range cf.Evidence {
		if cf.Evidence[i].ID == id {
			return &cf.Evidence[i]
		}
	}
	return nil
}

func testimonyByID(cf *models.CaseFile, id string) *models.Testimony {
	for i := range cf.Testimonies {
		if cf.Testimonies[i].ID == id {
			return &cf.Testimonies[i]
		}
	}
	return nil
}

func investigationJSON(iv *repository.Investigation) gin.H {
	var completed any
	if iv.CompletedAt != nil {
		completed = iv.CompletedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return gin.H{
		"id": iv.ID, "case_id": iv.CaseID, "status": iv.Status,
		"started_at":   iv.StartedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"completed_at": completed,
	}
}

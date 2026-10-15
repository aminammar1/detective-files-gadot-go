// Package investigations holds server-authoritative progress math: which
// exhibits a player holds, which testimonies they unlocked, and whether an
// accusation is proven. The client never sends scores or flags.
package investigations

import "detectivefiles/internal/models"

// EvidenceMet reports whether every required evidence id is in hand.
func EvidenceMet(required []string, have map[string]bool) bool {
	for _, id := range required {
		if !have[id] {
			return false
		}
	}
	return true
}

// TestimoniesMet reports whether every required testimony id is unlocked.
func TestimoniesMet(required []string, have map[string]string) bool {
	for _, id := range required {
		if _, ok := have[id]; !ok {
			return false
		}
	}
	return true
}

// Proven: the accused is the culprit AND holds the full required evidence
// AND unlocked the full required testimony set. Correct person alone is
// "unproven", never guilty.
func Proven(c *models.CaseFile, suspectID string, discovered map[string]bool, unlocked map[string]string) bool {
	if suspectID != c.CulpritSuspectID {
		return false
	}
	return EvidenceMet(c.AccusationRules.RequiredEvidence, discovered) &&
		TestimoniesMet(c.AccusationRules.RequiredTestimonies, unlocked)
}

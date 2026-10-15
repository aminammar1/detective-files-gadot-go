package cases

import (
	"encoding/json"
	"fmt"
	"sort"

	"detectivefiles/content"
	"detectivefiles/internal/models"
)

// Registry holds all validated cases keyed by case id.
type Registry struct {
	cases map[string]*models.CaseFile
	order []string
}

// Load reads every *.json case from the embedded content dir and validates
// strictly. Any broken cross-reference fails startup.
func Load() (*Registry, error) {
	entries, err := content.FS.ReadDir("cases")
	if err != nil {
		return nil, fmt.Errorf("read embedded cases: %w", err)
	}
	r := &Registry{cases: map[string]*models.CaseFile{}}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		raw, err := content.FS.ReadFile("cases/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read case %s: %w", e.Name(), err)
		}
		var c models.CaseFile
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, fmt.Errorf("parse case %s: %w", e.Name(), err)
		}
		applyAliases(raw, &c)
		if err := Validate(&c); err != nil {
			return nil, fmt.Errorf("invalid case %s: %w", e.Name(), err)
		}
		if _, dup := r.cases[c.ID]; dup {
			return nil, fmt.Errorf("duplicate case id %q", c.ID)
		}
		r.cases[c.ID] = &c
		r.order = append(r.order, c.ID)
	}
	sort.Strings(r.order)
	if len(r.cases) == 0 {
		return nil, fmt.Errorf("no cases embedded")
	}
	return r, nil
}

// applyAliases fills canonical fields from accepted aliases found in the raw
// JSON (evidence "summary", suspect "voice_note") when canonical is empty.
func applyAliases(raw []byte, c *models.CaseFile) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return
	}
	strField := func(obj map[string]json.RawMessage, key string) string {
		var s string
		if r, ok := obj[key]; ok {
			_ = json.Unmarshal(r, &s)
		}
		return s
	}
	var evList []map[string]json.RawMessage
	if r, ok := m["evidence"]; ok {
		_ = json.Unmarshal(r, &evList)
	}
	for i := range c.Evidence {
		if c.Evidence[i].Description == "" && i < len(evList) {
			c.Evidence[i].Description = strField(evList[i], "summary")
		}
	}
	var susList []map[string]json.RawMessage
	if r, ok := m["suspects"]; ok {
		_ = json.Unmarshal(r, &susList)
	}
	for i := range c.Suspects {
		if c.Suspects[i].Voice == "" && i < len(susList) {
			c.Suspects[i].Voice = strField(susList[i], "voice_note")
			if c.Suspects[i].Voice == "" {
				c.Suspects[i].Voice = strField(susList[i], "voiceNote")
			}
		}
	}
}

// Get returns the case by id or nil.
func (r *Registry) Get(id string) *models.CaseFile {
	return r.cases[id]
}

// IDs returns sorted case ids.
func (r *Registry) IDs() []string {
	out := append([]string(nil), r.order...)
	return out
}

// Validate enforces every structural rule: unique IDs and resolvable
// cross-references, with the culprit present in suspects.
func Validate(c *models.CaseFile) error {
	if c.ID == "" || c.Title == "" || c.Difficulty == "" || c.Briefing == "" || c.AuditLetter == "" {
		return fmt.Errorf("case %q: id/title/difficulty/briefing/audit_letter must all be non-empty", c.ID)
	}
	locIDs := map[string]bool{}
	for _, l := range c.Locations {
		if l.ID == "" || l.Name == "" {
			return fmt.Errorf("case %q: location with empty id/name", c.ID)
		}
		if locIDs[l.ID] {
			return fmt.Errorf("case %q: duplicate location id %q", c.ID, l.ID)
		}
		locIDs[l.ID] = true
	}
	if len(locIDs) == 0 {
		return fmt.Errorf("case %q: no locations", c.ID)
	}
	susIDs := map[string]bool{}
	for _, s := range c.Suspects {
		if s.ID == "" || s.Name == "" {
			return fmt.Errorf("case %q: suspect with empty id/name", c.ID)
		}
		if susIDs[s.ID] {
			return fmt.Errorf("case %q: duplicate suspect id %q", c.ID, s.ID)
		}
		susIDs[s.ID] = true
		seen := map[string]bool{}
		for _, t := range s.Topics {
			if t.ID == "" || t.Prompt == "" || t.Response == "" {
				return fmt.Errorf("case %q suspect %q: topic with empty id/prompt/response", c.ID, s.ID)
			}
			if seen[t.ID] {
				return fmt.Errorf("case %q suspect %q: duplicate topic id %q", c.ID, s.ID, t.ID)
			}
			seen[t.ID] = true
		}
	}
	if len(susIDs) == 0 {
		return fmt.Errorf("case %q: no suspects", c.ID)
	}
	evIDs := map[string]bool{}
	for _, e := range c.Evidence {
		if e.ID == "" || e.Name == "" {
			return fmt.Errorf("case %q: evidence with empty id/name", c.ID)
		}
		if evIDs[e.ID] {
			return fmt.Errorf("case %q: duplicate evidence id %q", c.ID, e.ID)
		}
		evIDs[e.ID] = true
		if !locIDs[e.LocationID] {
			return fmt.Errorf("case %q: evidence %q references unknown location %q", c.ID, e.ID, e.LocationID)
		}
	}
	testIDs := map[string]*models.Testimony{}
	for i := range c.Testimonies {
		t := &c.Testimonies[i]
		if t.ID == "" || t.Title == "" || t.Text == "" {
			return fmt.Errorf("case %q: testimony with empty id/title/text", c.ID)
		}
		if _, dup := testIDs[t.ID]; dup {
			return fmt.Errorf("case %q: duplicate testimony id %q", c.ID, t.ID)
		}
		testIDs[t.ID] = t
	}
	for _, s := range c.Suspects {
		for _, con := range s.Contradictions {
			if !evIDs[con.EvidenceID] {
				return fmt.Errorf("case %q suspect %q: contradiction references unknown evidence %q", c.ID, s.ID, con.EvidenceID)
			}
			if con.Response == "" {
				return fmt.Errorf("case %q suspect %q: contradiction for %q has empty response", c.ID, s.ID, con.EvidenceID)
			}
			if con.UnlocksTestimony != nil && *con.UnlocksTestimony != "" {
				if _, ok := testIDs[*con.UnlocksTestimony]; !ok {
					return fmt.Errorf("case %q suspect %q: contradiction references unknown testimony %q", c.ID, s.ID, *con.UnlocksTestimony)
				}
			}
		}
	}
	if !susIDs[c.CulpritSuspectID] {
		return fmt.Errorf("case %q: culprit %q not in suspects", c.ID, c.CulpritSuspectID)
	}
	for _, id := range c.AccusationRules.RequiredEvidence {
		if !evIDs[id] {
			return fmt.Errorf("case %q: required_evidence references unknown %q", c.ID, id)
		}
	}
	if len(c.AccusationRules.RequiredEvidence) == 0 {
		return fmt.Errorf("case %q: required_evidence is empty", c.ID)
	}
	for _, id := range c.AccusationRules.RequiredTestimonies {
		if _, ok := testIDs[id]; !ok {
			return fmt.Errorf("case %q: required_testimonies references unknown %q", c.ID, id)
		}
	}
	if len(c.AccusationRules.RequiredTestimonies) == 0 {
		return fmt.Errorf("case %q: required_testimonies is empty", c.ID)
	}
	if c.Scoring.BaseCorrect <= 0 || c.Scoring.PerEvidence <= 0 || c.Scoring.PerTestimony <= 0 {
		return fmt.Errorf("case %q: scoring base/per_evidence/per_testimony must be positive", c.ID)
	}
	if c.Scoring.ConsolationUnproven < 0 || c.Scoring.ConsolationWrong < 0 {
		return fmt.Errorf("case %q: scoring consolations must be non-negative", c.ID)
	}
	return nil
}

// Sanitize builds the spoiler-free public view of a case.
func Sanitize(c *models.CaseFile) models.CaseDetail {
	d := models.CaseDetail{
		ID:          c.ID,
		Title:       c.Title,
		Difficulty:  c.Difficulty,
		Briefing:    c.Briefing,
		AuditLetter: c.AuditLetter,
		Locations:   append([]models.Location(nil), c.Locations...),
	}
	for _, s := range c.Suspects {
		sp := models.SuspectPublic{
			ID:          s.ID,
			Name:        s.Name,
			Role:        s.Role,
			Description: s.Description,
			Dossier:     s.Dossier,
			Alibi:       s.Alibi,
			Asset:       s.Asset,
		}
		for _, t := range s.Topics {
			sp.Topics = append(sp.Topics, models.TopicPublic{ID: t.ID, Prompt: t.Prompt})
		}
		d.Suspects = append(d.Suspects, sp)
	}
	return d
}

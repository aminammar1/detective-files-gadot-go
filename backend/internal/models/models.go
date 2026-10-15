package models

// CaseFile mirrors backend/content/cases/*.json. The "voice" field on
// suspects and "description" on evidence are canonical; "voice_note" and
// "summary" are accepted as aliases on unmarshal.
type CaseFile struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Difficulty       string          `json:"difficulty"`
	Briefing         string          `json:"briefing"`
	AuditLetter      string          `json:"audit_letter"`
	Locations        []Location      `json:"locations"`
	Suspects         []Suspect       `json:"suspects"`
	Evidence         []Evidence      `json:"evidence"`
	Testimonies      []Testimony     `json:"testimonies"`
	CulpritSuspectID string          `json:"culprit_suspect_id"`
	AccusationRules  AccusationRules `json:"accusation_rules"`
	Scoring          Scoring         `json:"scoring"`
}

type Location struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Asset       string `json:"asset"`
}

type Topic struct {
	ID       string `json:"id"`
	Prompt   string `json:"prompt"`
	Response string `json:"response"`
}

type Contradiction struct {
	EvidenceID       string  `json:"evidence_id"`
	Response         string  `json:"response"`
	UnlocksTestimony *string `json:"unlocks_testimony"`
}

type Suspect struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Role           string          `json:"role"`
	Description    string          `json:"description"`
	Dossier        string          `json:"dossier"`
	Voice          string          `json:"voice"`
	Alibi          string          `json:"alibi"`
	Asset          string          `json:"asset"`
	Topics         []Topic         `json:"topics"`
	Contradictions []Contradiction `json:"contradictions"`
}

type Evidence struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LocationID  string `json:"location_id"`
	Description string `json:"description"`
	Proves      string `json:"proves"`
	Asset       string `json:"asset"`
}

type Testimony struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type AccusationRules struct {
	RequiredEvidence    []string `json:"required_evidence"`
	RequiredTestimonies []string `json:"required_testimonies"`
}

type Scoring struct {
	BaseCorrect         int `json:"base_correct"`
	PerEvidence         int `json:"per_evidence"`
	PerTestimony        int `json:"per_testimony"`
	ConsolationUnproven int `json:"consolation_unproven"`
	ConsolationWrong    int `json:"consolation_wrong"`
}

// ---- Public (sanitized) DTOs: never include culprit, rules, scoring,
// topic responses, contradiction responses, or evidence content. ----

type TopicPublic struct {
	ID     string `json:"id"`
	Prompt string `json:"prompt"`
}

type SuspectPublic struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Role        string        `json:"role"`
	Description string        `json:"description"`
	Dossier     string        `json:"dossier"`
	Alibi       string        `json:"alibi"`
	Asset       string        `json:"asset"`
	Topics      []TopicPublic `json:"topics"`
}

type CaseDetail struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Difficulty  string          `json:"difficulty"`
	Briefing    string          `json:"briefing"`
	AuditLetter string          `json:"audit_letter"`
	Locations   []Location      `json:"locations"`
	Suspects    []SuspectPublic `json:"suspects"`
}

type EvidenceWithTime struct {
	Evidence
	DiscoveredAt string `json:"discovered_at"`
}

type TestimonyWithTime struct {
	Testimony
	UnlockedAt string `json:"unlocked_at"`
}

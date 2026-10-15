package tests

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"detectivefiles/internal/cases"
)

// TestCaseContentValidity mirrors the loader's strict checks explicitly:
// every location/suspect/evidence/testimony id is unique and every
// cross-reference resolves.
func TestCaseContentValidity(t *testing.T) {
	reg, err := cases.Load()
	require.NoError(t, err)
	require.NotEmpty(t, reg.IDs())
	for _, id := range reg.IDs() {
		t.Run(id, func(t *testing.T) {
			cf := reg.Get(id)
			require.NotNil(t, cf)

			locs := map[string]bool{}
			for _, l := range cf.Locations {
				require.NotEmpty(t, l.ID)
				require.False(t, locs[l.ID], "dup location %q", l.ID)
				locs[l.ID] = true
			}
			sus := map[string]bool{}
			for _, s := range cf.Suspects {
				require.NotEmpty(t, s.ID)
				require.False(t, sus[s.ID], "dup suspect %q", s.ID)
				sus[s.ID] = true
			}
			ev := map[string]bool{}
			for _, e := range cf.Evidence {
				require.False(t, ev[e.ID], "dup evidence %q", e.ID)
				ev[e.ID] = true
				require.True(t, locs[e.LocationID], "evidence %q bad location %q", e.ID, e.LocationID)
			}
			tests := map[string]bool{}
			for _, x := range cf.Testimonies {
				require.False(t, tests[x.ID], "dup testimony %q", x.ID)
				tests[x.ID] = true
			}
			for _, s := range cf.Suspects {
				for _, con := range s.Contradictions {
					require.True(t, ev[con.EvidenceID], "bad evidence ref %q", con.EvidenceID)
					if con.UnlocksTestimony != nil {
						require.True(t, tests[*con.UnlocksTestimony], "bad testimony ref %q", *con.UnlocksTestimony)
					}
				}
			}
			require.True(t, sus[cf.CulpritSuspectID])
			for _, id := range cf.AccusationRules.RequiredEvidence {
				require.True(t, ev[id], "required evidence %q missing", id)
			}
			for _, id := range cf.AccusationRules.RequiredTestimonies {
				require.True(t, tests[id], "required testimony %q missing", id)
			}
		})
	}
}

// TestDeducibility simulates the accusation rule: the full required set
// proves the culprit guilty, and any other suspect is not guilty.
func TestDeducibility(t *testing.T) {
	reg, err := cases.Load()
	require.NoError(t, err)
	require.NotEmpty(t, reg.IDs())
	for _, id := range reg.IDs() {
		t.Run(id, func(t *testing.T) {
			cf := reg.Get(id)
			require.NotNil(t, cf)

			judge := func(suspect string, haveEv, haveTest map[string]bool) (correct, proven bool) {
				correct = suspect == cf.CulpritSuspectID
				proven = correct
				for _, id := range cf.AccusationRules.RequiredEvidence {
					if !haveEv[id] {
						proven = false
					}
				}
				for _, id := range cf.AccusationRules.RequiredTestimonies {
					if !haveTest[id] {
						proven = false
					}
				}
				return correct, proven
			}
			fullEv := map[string]bool{}
			for _, id := range cf.AccusationRules.RequiredEvidence {
				fullEv[id] = true
			}
			fullTest := map[string]bool{}
			for _, id := range cf.AccusationRules.RequiredTestimonies {
				fullTest[id] = true
			}

			correct, proven := judge(cf.CulpritSuspectID, fullEv, fullTest)
			require.True(t, correct)
			require.True(t, proven, "full required set must prove the culprit")

			_, proven = judge(cf.CulpritSuspectID, map[string]bool{}, map[string]bool{})
			require.False(t, proven, "empty hands must not prove the culprit")

			for _, s := range cf.Suspects {
				if s.ID == cf.CulpritSuspectID {
					continue
				}
				correct, _ = judge(s.ID, fullEv, fullTest)
				require.False(t, correct, "suspect %q must never be guilty", s.ID)
			}

			max := cf.Scoring.BaseCorrect + cf.Scoring.PerEvidence*len(cf.Evidence) + cf.Scoring.PerTestimony*len(cf.Testimonies)
			require.Greater(t, max, cf.Scoring.BaseCorrect)
			require.Equal(t, 1600, max, "both cases are designed 8 exhibits + 2 testimonies = 1600 max")
		})
	}
}

// TestNoCulpritLeak asserts the sanitized case view carries no spoilers.
// NOTE on decoy-safety: suspect ids/names (e.g. "edwin_marsh") legitimately
// appear in the public suspect roster, so a naive whole-body substring check
// for the culprit id would false-positive. The decoy-safe check instead
// asserts: no spoiler KEYS anywhere, and none of the secret TEXTS
// (contradiction responses, topic responses, evidence content, testimony
// texts, rules, scoring) appear in the payload.
func TestNoCulpritLeak(t *testing.T) {
	reg, err := cases.Load()
	require.NoError(t, err)
	require.NotEmpty(t, reg.IDs())
	for _, id := range reg.IDs() {
		t.Run(id, func(t *testing.T) {
			cf := reg.Get(id)
			require.NotNil(t, cf)

			raw, err := json.Marshal(cases.Sanitize(cf))
			require.NoError(t, err)
			body := string(raw)

			var decoded map[string]any
			require.NoError(t, json.Unmarshal(raw, &decoded))
			for _, key := range []string{"culprit_suspect_id", "accusation_rules", "scoring", "contradictions", "voice"} {
				require.NotContains(t, decoded, key, "spoiler key %q leaked", key)
			}
			suspects, ok := decoded["suspects"].([]any)
			require.True(t, ok)
			require.NotEmpty(t, suspects)
			for _, s := range suspects {
				sm := s.(map[string]any)
				for _, key := range []string{"contradictions", "voice", "response"} {
					require.NotContains(t, sm, key)
				}
				topics, ok := sm["topics"].([]any)
				require.True(t, ok, "public topics must be present")
				require.NotEmpty(t, topics)
				for _, tp := range topics {
					tm := tp.(map[string]any)
					require.Len(t, tm, 2, "public topic carries only id+prompt")
					require.Contains(t, tm, "id")
					require.Contains(t, tm, "prompt")
				}
			}

			// Secret texts must never appear in the payload.
			for _, s := range cf.Suspects {
				for _, topic := range s.Topics {
					require.NotContains(t, body, topic.Response, "topic response leaked")
				}
				for _, con := range s.Contradictions {
					require.NotContains(t, body, con.Response, "contradiction response leaked")
				}
			}
			for _, e := range cf.Evidence {
				require.NotContains(t, body, e.Name, "evidence name leaked")
				require.NotContains(t, body, e.Description, "evidence content leaked")
				require.NotContains(t, body, e.Proves, "evidence proof leaked")
			}
			for _, x := range cf.Testimonies {
				require.NotContains(t, body, x.Title, "testimony title leaked")
				require.NotContains(t, body, x.Text, "testimony text leaked")
			}
		})
	}
}

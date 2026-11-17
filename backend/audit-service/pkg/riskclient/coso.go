package riskclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// CosoSignalQuery selects the RCM / risk data of one assessment scope
type CosoSignalQuery struct {
	DepartmentID string
	Year         int
	Quarter      int    // 0 = annual
	LocationID   string // optional branch
	AsOf         string // YYYY-MM-DD
}

// CosoSignal is one risk-side signal (ra.quarterly_coverage,
// ra.above_appetite_mitigated, ca.rcm_effectiveness). State is OK, NO_DATA or
// NOT_APPLICABLE; ValuePct is set only with OK.
type CosoSignal struct {
	SignalKey   string                 `json:"signal_key"`
	State       string                 `json:"state"`
	ValuePct    *float64               `json:"value_pct"`
	Numerator   *int                   `json:"numerator"`
	Denominator *int                   `json:"denominator"`
	ScopeNote   string                 `json:"scope_note"`
	Detail      map[string]interface{} `json:"detail"`
	Principles  []CosoPrincipleMean    `json:"principles,omitempty"`
}

// CosoPrincipleMean is the mean total_weighted_score of the scope's RCM rows
// tagged with one principle
type CosoPrincipleMean struct {
	PrincipleCode string   `json:"principle_code"`
	ControlCount  int      `json:"control_count"`
	AvgPct        *float64 `json:"avg_pct"`
}

// CosoSignals is the answer of GET /api/v1/rcm/coso-signals
type CosoSignals struct {
	Signals    []CosoSignal        `json:"signals"`
	Principles []CosoPrincipleMean `json:"principles"`
}

// Signal returns the signal with key, or nil
func (s *CosoSignals) Signal(key string) *CosoSignal {
	if s == nil {
		return nil
	}
	for i := range s.Signals {
		if strings.EqualFold(strings.TrimSpace(s.Signals[i].SignalKey), key) {
			return &s.Signals[i]
		}
	}
	return nil
}

// PrincipleMeans returns the per-principle means: the top-level list, else
// the one carried by ca.rcm_effectiveness
func (s *CosoSignals) PrincipleMeans() []CosoPrincipleMean {
	if s == nil {
		return nil
	}
	if len(s.Principles) > 0 {
		return s.Principles
	}
	if ca := s.Signal("ca.rcm_effectiveness"); ca != nil {
		return ca.Principles
	}
	return nil
}

// GetCosoSignals calls GET /api/v1/rcm/coso-signals with the caller's token.
// Any non-200 answer (including a 404 from a risk-service that does not have
// the route yet) is an error, which the assessment records as
// SOURCE_UNAVAILABLE.
func (c *RiskClient) GetCosoSignals(ctx context.Context, authorization string, q CosoSignalQuery) (*CosoSignals, error) {
	v := url.Values{}
	v.Set("department_id", q.DepartmentID)
	v.Set("year", strconv.Itoa(q.Year))
	v.Set("quarter", strconv.Itoa(q.Quarter))
	if strings.TrimSpace(q.LocationID) != "" {
		v.Set("location_id", q.LocationID)
	}
	if strings.TrimSpace(q.AsOf) != "" {
		v.Set("as_of", q.AsOf)
	}
	data, err := c.getData(ctx, authorization, "/api/v1/rcm/coso-signals?"+v.Encode())
	if err != nil {
		return nil, err
	}
	return DecodeCosoSignals(data)
}

// DecodeCosoSignals accepts {signals: [...], principles: [...]}, a bare
// array of signals, or an object keyed by signal_key
func DecodeCosoSignals(data json.RawMessage) (*CosoSignals, error) {
	var out CosoSignals
	if err := json.Unmarshal(data, &out); err == nil && len(out.Signals) > 0 {
		return &out, nil
	}
	var arr []CosoSignal
	if err := json.Unmarshal(data, &arr); err == nil {
		return &CosoSignals{Signals: arr}, nil
	}
	var byKey map[string]json.RawMessage
	if err := json.Unmarshal(data, &byKey); err != nil {
		return nil, fmt.Errorf("decode coso signals: %w", err)
	}
	res := &CosoSignals{}
	for k, raw := range byKey {
		if k == "principles" {
			_ = json.Unmarshal(raw, &res.Principles)
			continue
		}
		var s CosoSignal
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}
		if s.SignalKey == "" {
			s.SignalKey = k
		}
		res.Signals = append(res.Signals, s)
	}
	return res, nil
}

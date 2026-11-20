package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"risk-service/models"
	"risk-service/pkg/masterclient"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Risk-side signals of the COSO 2013 component assessment
// (docs/coso-2013-component-assessment-plan.md 5.2, 5.3). Read-only. The
// weights, the score roll-up and the NULL-below-50%-coverage rule live in
// audit-service; this only reports each signal's value with its numerator,
// denominator and the ids behind it.

// Signal states (plan 4.2.4). SOURCE_UNAVAILABLE is set by the caller when
// this endpoint cannot be reached, never here.
const (
	SignalStateOK            = "OK"
	SignalStateNoData        = "NO_DATA"
	SignalStateNotApplicable = "NOT_APPLICABLE"
)

// Signal keys this service computes.
const (
	SignalRAQuarterlyCoverage      = "ra.quarterly_coverage"
	SignalRAAboveAppetiteMitigated = "ra.above_appetite_mitigated"
	SignalCARCMEffectiveness       = "ca.rcm_effectiveness"
)

// maxDetailIDs caps the ids listed in a signal's detail (plan 4.2.4).
const maxDetailIDs = 200

// CosoSignalsParams is the scope of one signal read.
type CosoSignalsParams struct {
	DepartmentID uuid.UUID
	Year         int
	Quarter      int        // 0 = annual, 1..4
	LocationID   *uuid.UUID // optional branch scope
	AsOf         time.Time  // date (UTC midnight)
	// LegacyDepartmentName is the master name of DepartmentID when it is the
	// only active master department with that name ("" otherwise). Legacy RCM
	// rows without department_id whose department text matches it (the
	// backfill-rcm-department rule) are counted as the department's rows
	// until they are backfilled or edited.
	LegacyDepartmentName string
}

// CosoSignalsScope echoes the scope used, with the derived period.
type CosoSignalsScope struct {
	DepartmentID string  `json:"department_id"`
	Year         int     `json:"year"`
	Quarter      int     `json:"quarter"`
	LocationID   *string `json:"location_id"`
	PeriodStart  string  `json:"period_start"`
	PeriodEnd    string  `json:"period_end"`
	AsOf         string  `json:"as_of"`
	DueQuarters  []int   `json:"due_quarters"`
	// BranchScoped is true when location_id narrowed the RCM and risk sets.
	BranchScoped bool `json:"branch_scoped"`
	// LocationMasterAvailable is false when a branch-scoped read could not
	// read the Location master: stored location ids were still used, legacy
	// branch names could not be matched (those rows count as unassigned).
	LocationMasterAvailable bool `json:"location_master_available"`
}

// CosoSignal is one signal value (plan 4.2.4 without weight, which audit-service owns).
type CosoSignal struct {
	SignalKey      string         `json:"signal_key"`
	ComponentCode  string         `json:"component_code"`
	PrincipleHints []string       `json:"principle_hints"`
	State          string         `json:"state"`
	ValuePct       *float64       `json:"value_pct"`
	Numerator      *int           `json:"numerator"`
	Denominator    *int           `json:"denominator"`
	SourceService  string         `json:"source_service"`
	SourceRef      string         `json:"source_ref"`
	ScopeNote      string         `json:"scope_note"`
	Detail         map[string]any `json:"detail"`
}

// CosoPrincipleStat is the tagged-control evidence shown next to a principle.
type CosoPrincipleStat struct {
	PrincipleCode       string   `json:"principle_code"`
	ComponentCode       string   `json:"component_code"`
	TaggedControlCount  int      `json:"tagged_control_count"`
	TaggedControlAvgPct *float64 `json:"tagged_control_avg_pct"`
}

// CosoSignalsResult is the response body of GET /rcm/coso-signals.
type CosoSignalsResult struct {
	Scope CosoSignalsScope `json:"scope"`
	// RCMCount/RiskCount are the sizes of the sets C and R (plan 5.2).
	RCMCount  int `json:"rcm_count"`
	RiskCount int `json:"risk_count"`
	// UnassignedRCMCount is the number of live RCM rows of the year with no
	// department_id (data quality context, not scored).
	UnassignedRCMCount int `json:"unassigned_rcm_count"`
	// LegacyNameMatchedRCMCount is how many of the department's RCM rows had
	// no department_id and were matched by department name (included in
	// RCMCount). Running `./risk backfill-rcm-department --apply` makes it 0.
	LegacyNameMatchedRCMCount int          `json:"legacy_name_matched_rcm_count"`
	Signals                   []CosoSignal `json:"signals"`
	CollectedAt               time.Time    `json:"collected_at"`
}

// PeriodBounds returns the first and last day of the year or quarter.
func PeriodBounds(year, quarter int) (time.Time, time.Time) {
	if quarter < 1 || quarter > 4 {
		return time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	}
	start := time.Date(year, time.Month(3*(quarter-1)+1), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 3, -1)
}

// DueQuarters lists the quarters whose risk assessment is due at asOf
// (plan 5.3, ra.quarterly_coverage): quarters of year, up to quarter when the
// assessment is quarterly, that ended on or before asOf. When none has ended
// yet, the current quarter (clamped to the period).
func DueQuarters(year, quarter int, asOf time.Time) []int {
	maxQ := 4
	if quarter >= 1 && quarter <= 4 {
		maxQ = quarter
	}
	var due []int
	for q := 1; q <= maxQ; q++ {
		_, end := PeriodBounds(year, q)
		if !end.After(asOf) {
			due = append(due, q)
		}
	}
	if len(due) > 0 {
		return due
	}
	cur := 1
	if asOf.Year() == year {
		cur = (int(asOf.Month())-1)/3 + 1
	} else if asOf.Year() > year {
		cur = 4
	}
	if cur > maxQ {
		cur = maxQ
	}
	return []int{cur}
}

func quarterValues(a models.RiskAssessment, q int) (impact, likelihood int) {
	switch q {
	case 1:
		return a.ImpactQ1, a.LikelihoodQ1
	case 2:
		return a.ImpactQ2, a.LikelihoodQ2
	case 3:
		return a.ImpactQ3, a.LikelihoodQ3
	case 4:
		return a.ImpactQ4, a.LikelihoodQ4
	}
	return 0, 0
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func pct(num, den int) *float64 {
	v := round2(float64(num) / float64(den) * 100)
	return &v
}

func intPtr(v int) *int { return &v }

func capIDs(ids []string) []string {
	sort.Strings(ids)
	if len(ids) > maxDetailIDs {
		return ids[:maxDetailIDs]
	}
	return ids
}

// ComputeCosoSignals reads the risk-side signals for one scope. locs may be
// nil when no branch scope is requested.
func ComputeCosoSignals(ctx context.Context, db *gorm.DB, locs masterclient.LocationSource, p CosoSignalsParams) (*CosoSignalsResult, error) {
	db = db.WithContext(ctx)
	start, end := PeriodBounds(p.Year, p.Quarter)
	due := DueQuarters(p.Year, p.Quarter, p.AsOf)
	res := &CosoSignalsResult{
		Scope: CosoSignalsScope{
			DepartmentID:            p.DepartmentID.String(),
			Year:                    p.Year,
			Quarter:                 p.Quarter,
			PeriodStart:             start.Format("2006-01-02"),
			PeriodEnd:               end.Format("2006-01-02"),
			AsOf:                    p.AsOf.Format("2006-01-02"),
			DueQuarters:             due,
			LocationMasterAvailable: true,
		},
		CollectedAt: time.Now().UTC(),
	}

	var unassigned int64
	if err := db.Model(&models.RiskControlMatrix{}).
		Where("year = ? AND department_id IS NULL", p.Year).Count(&unassigned).Error; err != nil {
		return nil, fmt.Errorf("count unassigned rcm: %w", err)
	}
	res.UnassignedRCMCount = int(unassigned)

	// C: live RCM rows of the department and year (annual data, also for a
	// quarterly assessment).
	var rcms []models.RiskControlMatrix
	if err := db.Where("department_id = ? AND year = ?", p.DepartmentID, p.Year).
		Order("id").Find(&rcms).Error; err != nil {
		return nil, fmt.Errorf("load rcm: %w", err)
	}
	if key := NormalizeDepartmentName(p.LegacyDepartmentName); key != "" && unassigned > 0 {
		var legacy []models.RiskControlMatrix
		if err := db.Where("department_id IS NULL AND year = ?", p.Year).Find(&legacy).Error; err != nil {
			return nil, fmt.Errorf("load legacy rcm: %w", err)
		}
		for _, r := range legacy {
			if NormalizeDepartmentName(r.Department) == key {
				rcms = append(rcms, r)
				res.LegacyNameMatchedRCMCount++
			}
		}
		sort.Slice(rcms, func(i, j int) bool { return rcms[i].ID.String() < rcms[j].ID.String() })
	}

	// Live registers linked from C, with profile and assessments.
	riskIDSet := map[uuid.UUID]struct{}{}
	for _, r := range rcms {
		if r.RiskID != nil && *r.RiskID != uuid.Nil {
			riskIDSet[*r.RiskID] = struct{}{}
		}
	}
	registers := map[uuid.UUID]*models.RiskRegister{}
	if len(riskIDSet) > 0 {
		ids := make([]uuid.UUID, 0, len(riskIDSet))
		for id := range riskIDSet {
			ids = append(ids, id)
		}
		var regs []models.RiskRegister
		if err := db.Preload("Profile").Preload("Assessments").
			Where("id IN ?", ids).Find(&regs).Error; err != nil {
			return nil, fmt.Errorf("load risks: %w", err)
		}
		for i := range regs {
			registers[regs[i].ID] = &regs[i]
		}
	}

	// Branch scope: keep the rows whose own branch, else whose linked risk's
	// branch, is the location.
	if p.LocationID != nil {
		loc := p.LocationID.String()
		res.Scope.LocationID = &loc
		res.Scope.BranchScoped = true
		var list []masterclient.Location
		available := false
		if locs != nil {
			if l, err := locs.ListLocations(ctx); err == nil {
				list, available = l, true
			}
		}
		res.Scope.LocationMasterAvailable = available
		resolver := NewRCMBranchResolver(list, available)
		kept := rcms[:0]
		for _, r := range rcms {
			var reg *models.RiskRegister
			if r.RiskID != nil {
				reg = registers[*r.RiskID]
			}
			if id := resolver.RowLocationID(r, reg); id != nil && *id == *p.LocationID {
				kept = append(kept, r)
			}
		}
		rcms = kept
	}

	// R: distinct live risks linked from C.
	var riskIDs []uuid.UUID
	seen := map[uuid.UUID]bool{}
	danglingRefs := 0
	for _, r := range rcms {
		if r.RiskID == nil || *r.RiskID == uuid.Nil || seen[*r.RiskID] {
			continue
		}
		seen[*r.RiskID] = true
		if _, ok := registers[*r.RiskID]; ok {
			riskIDs = append(riskIDs, *r.RiskID)
		} else {
			danglingRefs++
		}
	}
	sort.Slice(riskIDs, func(i, j int) bool { return riskIDs[i].String() < riskIDs[j].String() })
	res.RCMCount = len(rcms)
	res.RiskCount = len(riskIDs)

	scopeNote := "department, annual RCM data"
	if p.Quarter > 0 {
		scopeNote = fmt.Sprintf("department, annual RCM data for quarter %d", p.Quarter)
	}
	if res.Scope.BranchScoped {
		scopeNote = "branch-scoped: " + scopeNote
	}

	rcmIDStrings := make([]string, 0, len(rcms))
	for _, r := range rcms {
		rcmIDStrings = append(rcmIDStrings, r.ID.String())
	}
	riskIDStrings := make([]string, 0, len(riskIDs))
	for _, id := range riskIDs {
		riskIDStrings = append(riskIDStrings, id.String())
	}

	// Assessment rows of R for the year (one per risk and year).
	assessments := map[uuid.UUID]models.RiskAssessment{}
	for _, id := range riskIDs {
		for _, a := range registers[id].Assessments {
			if a.Year == p.Year {
				if _, dup := assessments[id]; !dup {
					assessments[id] = a
				}
			}
		}
	}

	res.Signals = append(res.Signals,
		quarterlyCoverageSignal(riskIDs, assessments, due, riskIDStrings, danglingRefs, scopeNote),
	)

	appetite, err := aboveAppetiteSignal(db, riskIDs, registers, assessments, due, scopeNote)
	if err != nil {
		return nil, err
	}
	res.Signals = append(res.Signals, appetite)

	ca, err := rcmEffectivenessSignal(db, rcms, rcmIDStrings, scopeNote)
	if err != nil {
		return nil, err
	}
	res.Signals = append(res.Signals, ca)
	return res, nil
}

func quarterlyCoverageSignal(riskIDs []uuid.UUID, assessments map[uuid.UUID]models.RiskAssessment, due []int, riskIDStrings []string, dangling int, scopeNote string) CosoSignal {
	s := CosoSignal{
		SignalKey:      SignalRAQuarterlyCoverage,
		ComponentCode:  "RA",
		PrincipleHints: []string{"P07", "P09"},
		SourceService:  "risk-service",
		SourceRef:      "risk_assessment",
		ScopeNote:      scopeNote,
		Detail: map[string]any{
			"due_quarters":      due,
			"risk_ids":          capIDs(riskIDStrings),
			"dangling_risk_ref": dangling,
		},
	}
	if len(riskIDs) == 0 {
		s.State = SignalStateNoData
		return s
	}
	filled, missing := 0, []string{}
	for _, id := range riskIDs {
		a, ok := assessments[id]
		for _, q := range due {
			if !ok {
				missing = append(missing, fmt.Sprintf("%s:Q%d", id, q))
				continue
			}
			imp, lik := quarterValues(a, q)
			if imp > 0 && lik > 0 {
				filled++
			} else {
				missing = append(missing, fmt.Sprintf("%s:Q%d", id, q))
			}
		}
	}
	den := len(riskIDs) * len(due)
	s.State = SignalStateOK
	s.Numerator, s.Denominator, s.ValuePct = intPtr(filled), intPtr(den), pct(filled, den)
	s.Detail["missing"] = capIDs(missing)
	return s
}

func aboveAppetiteSignal(db *gorm.DB, riskIDs []uuid.UUID, registers map[uuid.UUID]*models.RiskRegister, assessments map[uuid.UUID]models.RiskAssessment, due []int, scopeNote string) (CosoSignal, error) {
	s := CosoSignal{
		SignalKey:      SignalRAAboveAppetiteMitigated,
		ComponentCode:  "RA",
		PrincipleHints: []string{"P07"},
		SourceService:  "risk-service",
		SourceRef:      "risk_assessment, risk_register, risk_mitigation",
		ScopeNote:      scopeNote,
		Detail:         map[string]any{},
	}
	if len(riskIDs) == 0 {
		s.State = SignalStateNoData
		return s, nil
	}
	lastDue := due[len(due)-1]
	var above []uuid.UUID
	levels := map[string]string{}
	for _, id := range riskIDs {
		reg := registers[id]
		lik, imp := reg.InherentLikelihood, reg.InherentImpact
		source := "inherent"
		if a, ok := assessments[id]; ok {
			for q := lastDue; q >= 1; q-- {
				i, l := quarterValues(a, q)
				if i > 0 && l > 0 {
					lik, imp, source = l, i, fmt.Sprintf("Q%d", q)
					break
				}
			}
		}
		lvl := RiskLevelFor(lik, imp)
		levels[id.String()] = lvl + " (" + source + ")"
		if !WithinAppetite(lvl) {
			above = append(above, id)
		}
	}
	s.Detail["levels"] = levels
	if len(above) == 0 {
		s.State = SignalStateNotApplicable
		return s, nil
	}

	type row struct {
		RiskID uuid.UUID
		N      int
	}
	var rows []row
	if err := db.Model(&models.RiskMitigation{}).Select("risk_id, COUNT(*) AS n").
		Where("risk_id IN ?", above).Group("risk_id").Scan(&rows).Error; err != nil {
		return s, fmt.Errorf("count mitigations: %w", err)
	}
	mitigated := map[uuid.UUID]bool{}
	for _, r := range rows {
		if r.N > 0 {
			mitigated[r.RiskID] = true
		}
	}
	var withIDs, withoutIDs []string
	for _, id := range above {
		if mitigated[id] {
			withIDs = append(withIDs, id.String())
		} else {
			withoutIDs = append(withoutIDs, id.String())
		}
	}
	s.State = SignalStateOK
	s.Numerator, s.Denominator, s.ValuePct = intPtr(len(withIDs)), intPtr(len(above)), pct(len(withIDs), len(above))
	s.Detail["mitigated_risk_ids"] = capIDs(withIDs)
	s.Detail["unmitigated_risk_ids"] = capIDs(withoutIDs)
	return s, nil
}

// allDefaultRatings reports rows whose five ratings all read 3, the value
// unset ratings default to (plan R3). A heuristic: a real all-3 row counts too.
func allDefaultRatings(r models.RiskControlMatrix) bool {
	return r.DesignEffectivenessRating == 3 && r.OperatingEffectivenessRating == 3 &&
		r.CoverageCompletenessRating == 3 && r.TimelinessRating == 3 && r.AutomationMonitoringRating == 3
}

func rcmEffectivenessSignal(db *gorm.DB, rcms []models.RiskControlMatrix, rcmIDStrings []string, scopeNote string) (CosoSignal, error) {
	s := CosoSignal{
		SignalKey:      SignalCARCMEffectiveness,
		ComponentCode:  "CA",
		PrincipleHints: []string{"P10", "P11", "P12"},
		SourceService:  "risk-service",
		SourceRef:      "risk_control_matrix, rcm_coso_principles",
		ScopeNote:      scopeNote,
		Detail:         map[string]any{"rcm_ids": capIDs(rcmIDStrings)},
	}

	// Per-principle evidence for every catalogue principle, also when C is empty.
	var principles []models.CosoPrinciple
	if err := db.Where("framework = ?", models.CosoFramework2013).Order("number").Find(&principles).Error; err != nil {
		return s, fmt.Errorf("load principles: %w", err)
	}
	scoreByRCM := map[uuid.UUID]float64{}
	ids := make([]uuid.UUID, 0, len(rcms))
	defaults := 0
	sum := 0.0
	for _, r := range rcms {
		scoreByRCM[r.ID] = r.TotalWeightedScore
		ids = append(ids, r.ID)
		sum += r.TotalWeightedScore
		if allDefaultRatings(r) {
			defaults++
		}
	}
	tags, err := LoadRCMPrincipleCodes(db, ids)
	if err != nil {
		return s, err
	}
	type acc struct {
		n   int
		sum float64
	}
	byCode := map[string]*acc{}
	for rcmID, codes := range tags {
		for _, code := range codes {
			a := byCode[code]
			if a == nil {
				a = &acc{}
				byCode[code] = a
			}
			a.n++
			a.sum += scoreByRCM[rcmID]
		}
	}
	stats := make([]CosoPrincipleStat, 0, len(principles))
	for _, p := range principles {
		st := CosoPrincipleStat{PrincipleCode: p.Code, ComponentCode: p.ComponentCode}
		if a := byCode[p.Code]; a != nil && a.n > 0 {
			st.TaggedControlCount = a.n
			v := round2(a.sum / float64(a.n))
			st.TaggedControlAvgPct = &v
		}
		stats = append(stats, st)
	}
	s.Detail["per_principle"] = stats
	s.Detail["all_default_ratings_count"] = defaults
	s.Detail["untagged_count"] = len(rcms) - len(tags)

	if len(rcms) == 0 {
		s.State = SignalStateNoData
		return s, nil
	}
	s.State = SignalStateOK
	// Mean of total_weighted_score (already 0-100). Numerator/denominator are
	// not counts here: denominator is |C|, numerator is left null.
	v := round2(sum / float64(len(rcms)))
	s.ValuePct = &v
	s.Denominator = intPtr(len(rcms))
	return s, nil
}

// LoadRCMPrincipleCodes returns the tagged principle codes per RCM row, sorted
// by principle number. Rows without tags are absent from the map.
func LoadRCMPrincipleCodes(db *gorm.DB, rcmIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	out := map[uuid.UUID][]string{}
	const chunk = 1000
	for i := 0; i < len(rcmIDs); i += chunk {
		j := i + chunk
		if j > len(rcmIDs) {
			j = len(rcmIDs)
		}
		type row struct {
			RcmID uuid.UUID
			Code  string
		}
		var rows []row
		err := db.Table("rcm_coso_principles AS t").
			Select("t.rcm_id AS rcm_id, p.code AS code").
			Joins("JOIN coso_principles p ON p.id = t.principle_id").
			Where("t.rcm_id IN ?", rcmIDs[i:j]).
			Order("t.rcm_id, p.number").
			Scan(&rows).Error
		if err != nil {
			return nil, fmt.Errorf("load rcm principle tags: %w", err)
		}
		for _, r := range rows {
			out[r.RcmID] = append(out[r.RcmID], r.Code)
		}
	}
	return out, nil
}

// NormalizeDepartmentName is the department-name key used to match legacy
// department text to the master: trimmed, whitespace collapsed, lower case.
func NormalizeDepartmentName(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

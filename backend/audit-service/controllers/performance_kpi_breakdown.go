package controllers

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

const (
	kpiBreakdownDefaultPageSize = 10
	kpiBreakdownMaxPageSize     = 100

	KpiSourceStrategicPlan  = "strategic_plan"
	KpiSourceKpiAchievement = "kpi_achievement"

	KpiStatusExceeded       = "Exceeded"
	KpiStatusOnTrack        = "On Track"
	KpiStatusNeedsAttention = "Needs Attention"
	KpiStatusNoTarget       = "No Target"
)

// KpiBreakdownItem is one row of the KPI Detailed Breakdown table.
type KpiBreakdownItem struct {
	ID              string   `json:"id"`
	Source          string   `json:"source"`
	Metric          string   `json:"metric"`
	Category        string   `json:"category"`
	Period          string   `json:"period"`
	Unit            string   `json:"unit"`
	Target          float64  `json:"target"`
	Actual          float64  `json:"actual"`
	Gap             float64  `json:"gap"`
	GapIsPositive   bool     `json:"gapIsPositive"`
	AchievementRate *float64 `json:"achievementRate"`
	Status          string   `json:"status"`
}

// KpiBreakdownQuery holds the parsed, defaulted query parameters.
type KpiBreakdownQuery struct {
	Year     int
	Page     int
	PageSize int
	Search   string
	Category string
	Status   string
	Period   string
}

// ParseKpiBreakdownQuery applies the endpoint's defaults and bounds: year defaults
// to the current year, page to 1, page_size to 10 (anything outside 1..100 → 10).
func ParseKpiBreakdownQuery(ctx *gin.Context, now time.Time) KpiBreakdownQuery {
	q := KpiBreakdownQuery{Year: now.Year(), Page: 1, PageSize: kpiBreakdownDefaultPageSize}
	if y, err := strconv.Atoi(strings.TrimSpace(ctx.Query("year"))); err == nil && y > 0 {
		q.Year = y
	}
	if p, err := strconv.Atoi(strings.TrimSpace(ctx.Query("page"))); err == nil && p >= 1 {
		q.Page = p
	}
	if ps, err := strconv.Atoi(strings.TrimSpace(ctx.Query("page_size"))); err == nil && ps >= 1 && ps <= kpiBreakdownMaxPageSize {
		q.PageSize = ps
	}
	q.Search = strings.TrimSpace(ctx.Query("search"))
	q.Category = ctx.Query("category")
	q.Status = ctx.Query("status")
	q.Period = ctx.Query("period")
	return q
}

// GetKpiBreakdown returns the paginated KPI Detailed Breakdown for one year: the
// strategic plans active in that year (target from kpiTargets[year], actual from
// the matching KPI achievement when there is one) plus that year's KPI
// achievements that match no plan.
// GET /api/v1/performance/kpi-breakdown?year=&page=&page_size=&search=&category=&status=&period=
func (c *PerformanceStatsController) GetKpiBreakdown(ctx *gin.Context) {
	q := ParseKpiBreakdownQuery(ctx, time.Now())

	// Soft-deleted rows are excluded by GORM's default scope. Plans are filtered by
	// year in Go because the year can also come from the JSON-serialized kpiTargets.
	var plans []models.StrategicPlan
	if err := c.db.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).Find(&plans).Error; err != nil {
		if logger.Log != nil {
			logger.Error("kpi breakdown: failed to load strategic plans", zap.Error(err))
		}
		response.InternalServerError(ctx, "Failed to fetch KPI breakdown")
		return
	}

	// Most recently recorded achievement first, so it wins when several match one plan.
	var achievements []models.KPIAchievement
	if err := c.db.Where("year = ?", q.Year).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).
		Find(&achievements).Error; err != nil {
		if logger.Log != nil {
			logger.Error("kpi breakdown: failed to load KPI achievements", zap.Error(err))
		}
		response.InternalServerError(ctx, "Failed to fetch KPI breakdown")
		return
	}

	rows := BuildKpiBreakdownRows(plans, achievements, q.Year)
	// Filter options come from every row of the year, before search/filters/paging.
	filters := KpiBreakdownFilterOptions(rows)
	rows = FilterKpiBreakdownRows(rows, q)
	SortKpiBreakdownRows(rows)
	items, pagination := PaginateKpiBreakdownRows(rows, q.Page, q.PageSize)

	response.OK(ctx, "KPI breakdown fetched successfully", gin.H{
		"items":      items,
		"pagination": pagination,
		"filters":    filters,
	})
}

// KpiBreakdownFilters lists the values the category and period filters can take
// for the requested year.
type KpiBreakdownFilters struct {
	Categories []string `json:"categories"`
	Periods    []string `json:"periods"`
}

// KpiBreakdownFilterOptions returns the distinct non-empty categories (sorted
// alphabetically) and periods (sorted by periodSortKey) of rows. Both slices are
// non-nil so they encode as [] rather than null.
func KpiBreakdownFilterOptions(rows []KpiBreakdownItem) KpiBreakdownFilters {
	categories := []string{}
	periods := []string{}
	seenCategory := map[string]bool{}
	seenPeriod := map[string]bool{}
	for _, r := range rows {
		if r.Category != "" && !seenCategory[r.Category] {
			seenCategory[r.Category] = true
			categories = append(categories, r.Category)
		}
		if r.Period != "" && !seenPeriod[r.Period] {
			seenPeriod[r.Period] = true
			periods = append(periods, r.Period)
		}
	}
	sort.Strings(categories)
	sort.SliceStable(periods, func(i, j int) bool {
		gi, ki := periodSortKey(periods[i])
		gj, kj := periodSortKey(periods[j])
		if gi != gj {
			return gi < gj
		}
		if ki != kj {
			return ki < kj
		}
		return periods[i] < periods[j]
	})
	return KpiBreakdownFilters{Categories: categories, Periods: periods}
}

// periodSortKey orders the period filter options naturally:
//
//	group 0: years             "2025" < "2026"           (by year)
//	group 1: bare quarters     "Q1" < "Q2" < "Q3" < "Q4"
//	group 2: quarter-year      "Q4-2025" < "Q1-2026" < "Q2-2026"   (by year, then quarter)
//	group 3: anything else     alphabetical (e.g. "Semester 1")
//	group 4: "Tahunan"         always last (case-insensitive)
//
// Quarter matching is case-insensitive ("q1" sorts with "Q1"); ties within a
// group fall back to plain string order so the result is deterministic.
func periodSortKey(p string) (group int, key int) {
	s := strings.TrimSpace(p)
	if strings.EqualFold(s, "Tahunan") {
		return 4, 0
	}
	if y, err := strconv.Atoi(s); err == nil && y > 0 && len(s) == 4 {
		return 0, y
	}
	if len(s) >= 2 && (s[0] == 'Q' || s[0] == 'q') && s[1] >= '1' && s[1] <= '4' {
		quarter := int(s[1] - '0')
		if len(s) == 2 {
			return 1, quarter
		}
		if s[2] == '-' {
			if y, err := strconv.Atoi(s[3:]); err == nil && y > 0 && len(s[3:]) == 4 {
				return 2, y*10 + quarter
			}
		}
	}
	return 3, 0
}

// BuildKpiBreakdownRows merges the plans active in year with the year's KPI
// achievements. achievements must already be limited to year; their order decides
// which one is used when several match the same plan (first wins).
func BuildKpiBreakdownRows(plans []models.StrategicPlan, achievements []models.KPIAchievement, year int) []KpiBreakdownItem {
	rows := make([]KpiBreakdownItem, 0, len(plans)+len(achievements))
	var activePlanNames []string

	for _, p := range plans {
		if !planActiveInYear(p, year) {
			continue
		}
		planName := planMatchName(p)
		activePlanNames = append(activePlanNames, planName)

		target := parseLeadingFloat(p.Target)
		if t, ok := yearTarget(p, year); ok {
			target = t
		}
		// actualKnown is false when the plan has no numeric actual and no achievement
		// matches it, i.e. nothing has been recorded yet (see newKpiBreakdownItem).
		actual, actualKnown := leadingFloat(p.Actual)
		for _, a := range achievements {
			if kpiNamesMatch(planName, a.KPIName) {
				actual, actualKnown = a.Actual, true
				break
			}
		}

		metric := strings.TrimSpace(p.KPI)
		if metric == "" {
			metric = strings.TrimSpace(p.StrategicObjective)
		}
		rows = append(rows, newKpiBreakdownItem(KpiBreakdownItem{
			ID:       p.ID.String(),
			Source:   KpiSourceStrategicPlan,
			Metric:   metric,
			Category: p.Category,
			Period:   p.SelectedPeriod,
			Unit:     p.Unit,
			Target:   target,
			Actual:   actual,
		}, actualKnown, higherIsGood(p.HibHig)))
	}

	for _, a := range achievements {
		matched := false
		for _, name := range activePlanNames {
			if kpiNamesMatch(name, a.KPIName) {
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		// KPI achievements carry no HIB/HIG flag and no category; they are treated as
		// higher-is-good (gap = actual - target, rate = actual/target) and have
		// category "".
		rows = append(rows, newKpiBreakdownItem(KpiBreakdownItem{
			ID:     a.ID.String(),
			Source: KpiSourceKpiAchievement,
			Metric: strings.TrimSpace(a.KPIName),
			Period: a.Period,
			Target: a.Target,
			Actual: a.Actual,
		}, true, true))
	}
	return rows
}

// newKpiBreakdownItem fills Gap, GapIsPositive, AchievementRate and Status of
// base from its Target and Actual.
//
// Gap: HIG actual - target, HIB target - actual (positive = better than target).
//
// Achievement rate, mirroring the plan view modal (StrategicObjectiveViewModal.vue):
//
//	target <= 0                     → rate null, "No Target"
//	HIG                             → actual / target * 100
//	HIB                             → target / actual * 100 (lower actual is better)
//	HIB, actual == 0                → rate null, "Exceeded" (the modal's +Infinity,
//	                                  which JSON cannot carry)
//	HIB, actual < 0                 → negative rate → "Needs Attention" (as the modal)
//	HIB, nothing recorded           → rate 0, "Needs Attention" (actualKnown false: the
//	                                  modal shows no rate; a missing actual must not
//	                                  read as a perfect 0 for HIB, and this matches what
//	                                  HIG gets for a missing actual)
//
// Status from the rate: >= 100 Exceeded, >= 80 On Track, else Needs Attention.
func newKpiBreakdownItem(base KpiBreakdownItem, actualKnown, hig bool) KpiBreakdownItem {
	item := base
	target, actual := item.Target, item.Actual

	gap := target - actual
	if hig {
		gap = actual - target
	}
	item.Gap = finite(gap)
	item.GapIsPositive = item.Gap >= 0
	item.AchievementRate = nil
	item.Status = KpiStatusNoTarget
	if target <= 0 {
		return item
	}

	var rate float64
	switch {
	case hig:
		rate = actual / target * 100
	case !actualKnown:
		rate = 0
	case actual == 0:
		item.Status = KpiStatusExceeded
		return item
	default:
		rate = target / actual * 100
	}
	rate = finite(rate)
	item.AchievementRate = &rate
	switch {
	case rate >= 100:
		item.Status = KpiStatusExceeded
	case rate >= 80:
		item.Status = KpiStatusOnTrack
	default:
		item.Status = KpiStatusNeedsAttention
	}
	return item
}

// FilterKpiBreakdownRows applies search (case-insensitive substring on the metric)
// and the exact-match category/status/period filters; empty means no filter.
func FilterKpiBreakdownRows(rows []KpiBreakdownItem, q KpiBreakdownQuery) []KpiBreakdownItem {
	search := strings.ToLower(q.Search)
	out := rows[:0:0]
	for _, r := range rows {
		if search != "" && !strings.Contains(strings.ToLower(r.Metric), search) {
			continue
		}
		if q.Category != "" && r.Category != q.Category {
			continue
		}
		if q.Status != "" && r.Status != q.Status {
			continue
		}
		if q.Period != "" && r.Period != q.Period {
			continue
		}
		out = append(out, r)
	}
	return out
}

// SortKpiBreakdownRows orders by metric (case-insensitive), then metric, then id.
func SortKpiBreakdownRows(rows []KpiBreakdownItem) {
	sort.SliceStable(rows, func(i, j int) bool {
		li, lj := strings.ToLower(rows[i].Metric), strings.ToLower(rows[j].Metric)
		if li != lj {
			return li < lj
		}
		if rows[i].Metric != rows[j].Metric {
			return rows[i].Metric < rows[j].Metric
		}
		return rows[i].ID < rows[j].ID
	})
}

// PaginateKpiBreakdownRows slices one page; a page past the end yields no items.
func PaginateKpiBreakdownRows(rows []KpiBreakdownItem, page, pageSize int) ([]KpiBreakdownItem, gin.H) {
	total := len(rows)
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	items := []KpiBreakdownItem{}
	if page <= totalPages {
		start := (page - 1) * pageSize
		end := start + pageSize
		if end > total {
			end = total
		}
		items = append(items, rows[start:end]...)
	}
	return items, gin.H{
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": totalPages,
	}
}

// planActiveInYear: YearStart <= year <= YearEnd, or kpiTargets has a non-empty
// target for that year (key "2026", or a quarterly key such as "Q1-2026").
func planActiveInYear(p models.StrategicPlan, year int) bool {
	if p.YearStart <= year && year <= p.YearEnd {
		return true
	}
	yr := strconv.Itoa(year)
	for key, val := range p.KPITargets {
		if strings.TrimSpace(val) == "" {
			continue
		}
		if key == yr {
			return true
		}
		if len(key) == len("Q1-")+len(yr) && (key[0] == 'Q' || key[0] == 'q') &&
			key[1] >= '1' && key[1] <= '4' && key[2] == '-' && key[3:] == yr {
			return true
		}
	}
	return false
}

// yearTarget returns kpiTargets[year] when it is present and numeric.
func yearTarget(p models.StrategicPlan, year int) (float64, bool) {
	raw, ok := p.KPITargets[strconv.Itoa(year)]
	if !ok {
		return 0, false
	}
	if v, ok := leadingFloat(raw); ok {
		return v, true
	}
	return 0, false
}

// planMatchName is the name used to match a plan to KPI achievements: the KPI,
// falling back to the strategic objective (as the frontend did).
func planMatchName(p models.StrategicPlan) string {
	name := strings.TrimSpace(p.KPI)
	if name == "" {
		name = strings.TrimSpace(p.StrategicObjective)
	}
	return name
}

// kpiNamesMatch: case-insensitive containment either way; empty names never match.
func kpiNamesMatch(a, b string) bool {
	a, b = strings.ToLower(strings.TrimSpace(a)), strings.ToLower(strings.TrimSpace(b))
	if a == "" || b == "" {
		return false
	}
	return strings.Contains(a, b) || strings.Contains(b, a)
}

// higherIsGood reports whether the plan's direction is HIG ("High is Good").
// The other stored value is "HIB" ("High is Bad").
func higherIsGood(hibHig string) bool {
	return strings.EqualFold(strings.TrimSpace(hibHig), "HIG")
}

// parseLeadingFloat mimics JavaScript parseFloat for the plan's free-text
// target/actual ("84.0", "90%", " 4.5 / 5"), returning 0 when nothing parses.
func parseLeadingFloat(s string) float64 {
	v, _ := leadingFloat(s)
	return v
}

func leadingFloat(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	end := 0
	seenDigit, seenDot, seenExp := false, false, false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= '0' && ch <= '9':
			seenDigit = true
			end = i + 1
		case (ch == '+' || ch == '-') && (i == 0 || s[i-1] == 'e' || s[i-1] == 'E'):
		case ch == '.' && !seenDot && !seenExp:
			seenDot = true
		case (ch == 'e' || ch == 'E') && seenDigit && !seenExp:
			seenExp = true
		default:
			i = len(s)
		}
	}
	if !seenDigit {
		return 0, false
	}
	v, err := strconv.ParseFloat(s[:end], 64)
	if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, false
	}
	return v, true
}

// finite keeps NaN/Inf (which encoding/json rejects) out of the response.
func finite(v float64) float64 {
	switch {
	case math.IsNaN(v):
		return 0
	case math.IsInf(v, 1):
		return math.MaxFloat64
	case math.IsInf(v, -1):
		return -math.MaxFloat64
	}
	return v
}

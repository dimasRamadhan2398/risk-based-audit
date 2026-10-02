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
	rows = FilterKpiBreakdownRows(rows, q)
	SortKpiBreakdownRows(rows)
	items, pagination := PaginateKpiBreakdownRows(rows, q.Page, q.PageSize)

	response.OK(ctx, "KPI breakdown fetched successfully", gin.H{
		"items":      items,
		"pagination": pagination,
	})
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
		actual := parseLeadingFloat(p.Actual)
		for _, a := range achievements {
			if kpiNamesMatch(planName, a.KPIName) {
				actual = a.Actual
				break
			}
		}

		metric := strings.TrimSpace(p.KPI)
		if metric == "" {
			metric = strings.TrimSpace(p.StrategicObjective)
		}
		rows = append(rows, newKpiBreakdownItem(
			p.ID.String(), KpiSourceStrategicPlan, metric, p.SelectedPeriod, p.Unit,
			target, actual, higherIsGood(p.HibHig),
		))
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
		// KPI achievements carry no HIB/HIG flag; they are treated as higher-is-good,
		// as the frontend did (gap = actual - target), which also agrees with the
		// actual/target achievement rate used for the status.
		rows = append(rows, newKpiBreakdownItem(
			a.ID.String(), KpiSourceKpiAchievement, strings.TrimSpace(a.KPIName), a.Period, "",
			a.Target, a.Actual, true,
		))
	}
	return rows
}

func newKpiBreakdownItem(id, source, metric, period, unit string, target, actual float64, hig bool) KpiBreakdownItem {
	gap := target - actual
	if hig {
		gap = actual - target
	}
	gap = finite(gap)

	item := KpiBreakdownItem{
		ID:            id,
		Source:        source,
		Metric:        metric,
		Category:      "", // neither StrategicPlan nor KPIAchievement has a category field
		Period:        period,
		Unit:          unit,
		Target:        target,
		Actual:        actual,
		Gap:           gap,
		GapIsPositive: gap >= 0,
		Status:        KpiStatusNoTarget,
	}
	if target > 0 {
		rate := finite(actual / target * 100)
		item.AchievementRate = &rate
		switch {
		case rate >= 100:
			item.Status = KpiStatusExceeded
		case rate >= 80:
			item.Status = KpiStatusOnTrack
		default:
			item.Status = KpiStatusNeedsAttention
		}
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

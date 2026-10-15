package executivesummary

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"audit-service/models"

	"gorm.io/gorm"
)

// SyncRiskColumnsBeforeUpdate is the crud.UpdateHook of PUT /executive-summaries/:id.
// It keeps the five-level risk counts and the legacy three-level counts in step:
//   - payload sets any five-level key: those values are written as integers
//     (null -> 0) and the legacy counts are derived from the resulting five
//     (models.LegacyRiskCounts);
//   - payload sets only legacy keys (old client): they are copied into the five
//     levels (Tinggi -> High, Sedang -> Moderate, Rendah -> Low, the in-between
//     levels become 0).
func SyncRiskColumnsBeforeUpdate(_ *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	e, ok := existing.(*models.ExecutiveSummary)
	if !ok {
		return nil
	}
	value := func(col string, current int) int {
		if raw, ok := columns[col]; ok {
			return toInt(raw)
		}
		return current
	}

	if hasAnyColumn(columns, models.ExecutiveSummaryRiskColumns) {
		high := value("risiko_high", e.RisikoHigh)
		moderateHigh := value("risiko_moderate_high", e.RisikoModerateHigh)
		moderate := value("risiko_moderate", e.RisikoModerate)
		lowModerate := value("risiko_low_moderate", e.RisikoLowModerate)
		low := value("risiko_low", e.RisikoLow)
		columns["risiko_high"] = high
		columns["risiko_moderate_high"] = moderateHigh
		columns["risiko_moderate"] = moderate
		columns["risiko_low_moderate"] = lowModerate
		columns["risiko_low"] = low
		columns["risiko_tinggi"], columns["risiko_sedang"], columns["risiko_rendah"] = models.LegacyRiskCounts(high, moderateHigh, moderate, lowModerate, low)
		return nil
	}

	if hasAnyColumn(columns, models.ExecutiveSummaryLegacyRiskColumns) {
		tinggi := value("risiko_tinggi", e.RisikoTinggi)
		sedang := value("risiko_sedang", e.RisikoSedang)
		rendah := value("risiko_rendah", e.RisikoRendah)
		columns["risiko_tinggi"], columns["risiko_sedang"], columns["risiko_rendah"] = tinggi, sedang, rendah
		columns["risiko_high"] = tinggi
		columns["risiko_moderate_high"] = 0
		columns["risiko_moderate"] = sedang
		columns["risiko_low_moderate"] = 0
		columns["risiko_low"] = rendah
	}
	return nil
}

func hasAnyColumn(columns map[string]interface{}, names []string) bool {
	for _, n := range names {
		if _, ok := columns[n]; ok {
			return true
		}
	}
	return false
}

// toInt reads a JSON-decoded count; null, empty or non-numeric values are 0
func toInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(math.Round(n))
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		if f, err := n.Float64(); err == nil {
			return int(math.Round(f))
		}
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err == nil {
			return int(math.Round(f))
		}
	}
	return 0
}

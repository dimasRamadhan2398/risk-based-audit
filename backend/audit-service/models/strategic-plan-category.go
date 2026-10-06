package models

import "strings"

// Strategic plan categories. A plan may also have no category ("").
const (
	StrategicPlanCategoryOperational = "Operational"
	StrategicPlanCategoryFinancial   = "Financial"
	StrategicPlanCategoryQuality     = "Quality"
	StrategicPlanCategoryIssue       = "Issue"
	StrategicPlanCategoryEfficiency  = "Efficiency"
)

// StrategicPlanCategories is the set of non-empty categories a plan may be saved with.
var StrategicPlanCategories = []string{
	StrategicPlanCategoryOperational,
	StrategicPlanCategoryFinancial,
	StrategicPlanCategoryQuality,
	StrategicPlanCategoryIssue,
	StrategicPlanCategoryEfficiency,
}

// NormalizeStrategicPlanCategory maps a category to its stored form: surrounding
// whitespace is ignored and the match is case-insensitive ("financial" ->
// "Financial"). A blank value is "" (no category). ok is false for unknown values.
func NormalizeStrategicPlanCategory(raw string) (category string, ok bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", true
	}
	for _, known := range StrategicPlanCategories {
		if strings.EqualFold(s, known) {
			return known, true
		}
	}
	return "", false
}

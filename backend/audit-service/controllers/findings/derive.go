// Package findings derives audit findings from the data auditors record
// (saved audit result reports, Digital Working Paper / KKA tabs, fieldwork
// test controls) for the auto-findings and recent-findings endpoints.
package findings

import (
	"fmt"
	"strings"
	"time"

	"audit-service/models"
)

// Source labels shown to the user
const (
	SourceWorkingPaper      = "Digital Working Paper (KKA - AOI & RCA)"
	SourceFieldwork         = "Audit Fieldwork (Test Controls)"
	SourceAuditResultReport = "Audit Result Report (LHA)"
)

// Finding is one derived finding. Date is the updated_at (falling back to
// created_at) of the row it was derived from.
type Finding struct {
	Title    string
	Category string
	Action   string
	Source   string
	Impact   string
	Criteria string
	Date     time.Time
}

// LiveSources holds the rows of ONE assignment letter that findings are
// derived from. Rows should be in creation order: KKA causes are paired with
// action plans by position.
type LiveSources struct {
	Causes       []models.WorkingPaperCause    // KKA F04 AOI & RCA
	Plans        []models.WorkingPaperPlan     // KKA F05 action plan
	Risks        []models.WorkingPaperRisk     // KKA F02 risk profile
	TestControls []models.FieldworkTestControl // Audit Fieldwork
}

func rowDate(updated, created time.Time) time.Time {
	if !updated.IsZero() {
		return updated
	}
	return created
}

func planAction(p models.WorkingPaperPlan) string {
	if p.ActionDescription != "" {
		return p.ActionDescription
	}
	return p.Recommendation
}

// Derive builds the live findings of one assignment letter, deduplicated by
// lower-cased title (first occurrence wins: KKA, then fieldwork).
// This is the logic behind GET /audit-result-reports/auto-findings.
func Derive(src LiveSources) []Finding {
	var out []Finding
	seen := make(map[string]bool)

	// Default category from the KKA risk profile (F02)
	defaultCategory := "Significant"
	for _, r := range src.Risks {
		rLevel := strings.ToUpper(strings.TrimSpace(r.RiskLevel))
		if rLevel == "HIGH" || rLevel == "CRITICAL" || rLevel == "VERY HIGH" {
			defaultCategory = "Very Significant"
			break
		}
	}

	// 1. KKA F04 causes (condition = finding), paired with F05 plans by position
	for idx, cause := range src.Causes {
		cond := strings.TrimSpace(cause.Condition)
		if cond == "" {
			continue
		}
		key := strings.ToLower(cond)
		if seen[key] {
			continue
		}
		seen[key] = true

		action := ""
		if idx < len(src.Plans) {
			action = planAction(src.Plans[idx])
		} else if len(src.Plans) > 0 {
			action = planAction(src.Plans[0])
		}

		cat := defaultCategory
		if strings.Contains(key, "kritis") || strings.Contains(key, "critical") || strings.Contains(key, "tidak sesuai") || strings.Contains(key, "override") || strings.Contains(key, "mfa") {
			cat = "Very Significant"
		}

		out = append(out, Finding{
			Title:    cond,
			Category: cat,
			Action:   action,
			Source:   SourceWorkingPaper,
			Impact:   cause.Impact,
			Criteria: cause.Criteria,
			Date:     rowDate(cause.UpdatedAt, cause.CreatedAt),
		})
	}

	// 2. Fieldwork test controls that are ineffective / partially effective
	// or carry a finding
	for _, tc := range src.TestControls {
		findingText := strings.TrimSpace(tc.Finding)
		resultUpper := strings.ToUpper(strings.TrimSpace(tc.TestResult))
		if findingText == "" && resultUpper != "INEFFECTIVE" && resultUpper != "PARTIALLY EFFECTIVE" {
			continue
		}
		title := findingText
		if title == "" {
			title = fmt.Sprintf("Kelemahan Kontrol: %s", tc.ControlName)
		}
		key := strings.ToLower(title)
		if seen[key] {
			continue
		}
		seen[key] = true

		action := strings.TrimSpace(tc.MitigationPlan)
		if action == "" {
			action = strings.TrimSpace(tc.Recommendation)
		}

		cat := "Significant"
		if resultUpper == "INEFFECTIVE" {
			cat = "Very Significant"
		}

		out = append(out, Finding{
			Title:    title,
			Category: cat,
			Action:   action,
			Source:   SourceFieldwork,
			Date:     rowDate(tc.UpdatedAt, tc.CreatedAt),
		})
	}

	return out
}

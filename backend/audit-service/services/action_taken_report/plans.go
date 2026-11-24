package action_taken_report

import (
	"encoding/json"
	"strings"
	"time"

	"audit-service/models"
)

// The LHA finding's actionPlans (Rencana Aksi) are the source of truth for its
// ATR. A finding without actionPlans keeps working through its legacy single
// Action string, which becomes the ATR recommendation.

// cleanPlans drops fully empty plans and trims the text
func cleanPlans(in []models.FindingActionPlan) []models.FindingActionPlan {
	out := make([]models.FindingActionPlan, 0, len(in))
	for _, p := range in {
		p.Action = strings.TrimSpace(p.Action)
		p.TargetDate = strings.TrimSpace(p.TargetDate)
		p.PIC = strings.TrimSpace(p.PIC)
		p.FinalApprover = strings.TrimSpace(p.FinalApprover)
		p.Evidence = strings.TrimSpace(p.Evidence)
		if p == (models.FindingActionPlan{}) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// findingRecommendation is the ATR recommendation: the action plans' actions
// (one per line) or the legacy Action
func findingRecommendation(f models.AuditReportFinding) string {
	plans := cleanPlans(f.ActionPlans)
	var lines []string
	for _, p := range plans {
		if p.Action != "" {
			lines = append(lines, p.Action)
		}
	}
	if len(lines) == 0 {
		return strings.TrimSpace(f.Action)
	}
	return strings.Join(lines, "\n")
}

func firstFinalApprover(in []models.FindingActionPlan) string {
	for _, p := range cleanPlans(in) {
		if p.FinalApprover != "" {
			return p.FinalApprover
		}
	}
	return ""
}

func firstPIC(in []models.FindingActionPlan) string {
	for _, p := range cleanPlans(in) {
		if p.PIC != "" {
			return p.PIC
		}
	}
	return ""
}

var targetLayouts = []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"}

// earliestTarget is the earliest parseable target date of the plans (free text
// such as "Q4 2026" is ignored), or nil
func earliestTarget(in []models.FindingActionPlan) *time.Time {
	var best *time.Time
	for _, p := range cleanPlans(in) {
		for _, layout := range targetLayouts {
			if t, err := time.Parse(layout, p.TargetDate); err == nil {
				t := t
				if best == nil || t.Before(*best) {
					best = &t
				}
				break
			}
		}
	}
	return best
}

// planColumns are the ATR columns refreshed from a finding's action plans
func planColumns(f models.AuditReportFinding) map[string]interface{} {
	encoded, _ := json.Marshal(cleanPlans(f.ActionPlans))
	cols := map[string]interface{}{
		"planned_actions": string(encoded),
		"final_approver":  firstFinalApprover(f.ActionPlans),
	}
	// pic_name and due_date are only overwritten when the plans supply them,
	// so a legacy finding never blanks values set elsewhere
	if pic := firstPIC(f.ActionPlans); pic != "" {
		cols["pic_name"] = pic
	}
	if t := earliestTarget(f.ActionPlans); t != nil {
		cols["due_date"] = *t
	}
	return cols
}

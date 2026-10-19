package main

import (
	"fmt"
	"testing"

	"audit-service/models"
)

func TestAdditionalAnnualPlanSeeds(t *testing.T) {
	seeds := getAdditionalAnnualPlanSeeds()
	if len(seeds) != 10 {
		t.Fatalf("expected 10 additional annual plans, got %d", len(seeds))
	}

	seen := make(map[string]bool, len(seeds))
	validStatuses := map[string]bool{
		models.AnnualPlanStatusDraft:           true,
		models.AnnualPlanStatusPendingReview:   true,
		models.AnnualPlanStatusPendingApproval: true,
		models.AnnualPlanStatusApproved:        true,
		models.AnnualPlanStatusRevised:         true,
	}

	for i, plan := range seeds {
		wantCode := fmt.Sprintf("PKAT-2026-%03d", i+2)
		if plan.Code != wantCode {
			t.Errorf("seed %d: code = %q, want %q", i, plan.Code, wantCode)
		}
		if seen[plan.Code] {
			t.Errorf("duplicate plan code %q", plan.Code)
		}
		seen[plan.Code] = true

		if plan.Title == "" || plan.ActivityType == "" || plan.Year != 2026 {
			t.Errorf("%s: missing realistic plan header data", plan.Code)
		}
		if !validStatuses[plan.Status] {
			t.Errorf("%s: invalid status %q", plan.Code, plan.Status)
		}
		if len(plan.SelectedMonths) == 0 || len(plan.Quarters) == 0 {
			t.Errorf("%s: schedule is incomplete", plan.Code)
		}
		if len(plan.Activities) == 0 {
			t.Errorf("%s: expected at least one activity", plan.Code)
			continue
		}

		activity := plan.Activities[0]
		if activity.Name == "" || activity.Department == "" || activity.RiskName == "" || activity.RiskLevel == "" {
			t.Errorf("%s: activity identity/risk data is incomplete", plan.Code)
		}
		if activity.Justification == "" || activity.AuditObjective == "" || activity.AuditFocus == "" || activity.TimelineText == "" {
			t.Errorf("%s: activity audit detail is incomplete", plan.Code)
		}
		if activity.AuditorCount < 1 || activity.TotalMandays < 1 || activity.SupervisorName == "" {
			t.Errorf("%s: activity resource allocation is incomplete", plan.Code)
		}
		if plan.AuditorCount != activity.AuditorCount || plan.TotalMandays != activity.TotalMandays {
			t.Errorf("%s: plan totals do not match its activity", plan.Code)
		}
	}
}

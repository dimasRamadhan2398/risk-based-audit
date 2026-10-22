package main

import (
	"strings"
	"testing"

	"risk-service/models"
	"risk-service/pkg/sqlitetest"
)

// `seed --catalog-only` on an empty database fills the reference tables only,
// and a second run changes nothing.
func TestSeedCatalogOnlySeedsReferenceDataOnceAndNothingElse(t *testing.T) {
	catalog := map[string]any{
		"risk_level":              &models.RiskLevel{},
		"standard_risk_factors":   &models.StandardRiskFactor{},
		"standard_audit_universe": &models.StandardAuditUniverse{},
	}
	other := map[string]any{
		"corporate_risk_factors":     &models.CorporateRiskFactor{},
		"corporate_audit_universe":   &models.CorporateAuditUniverse{},
		"audit_universe_years":       &models.AuditUniverseYear{},
		"audit_universe_risk_scores": &models.AuditUniverseRiskScore{},
		"risk_profile":               &models.RiskProfile{},
		"risk_register":              &models.RiskRegister{},
		"risk_assessment":            &models.RiskAssessment{},
		"risk_control_matrix":        &models.RiskControlMatrix{},
	}
	var all []any
	for _, m := range catalog {
		all = append(all, m)
	}
	for _, m := range other {
		all = append(all, m)
	}
	db := sqlitetest.Open(t, all...)

	counts := func() map[string]int64 {
		out := map[string]int64{}
		for name, m := range catalog {
			var n int64
			if err := db.Model(m).Count(&n).Error; err != nil {
				t.Fatalf("count %s: %v", name, err)
			}
			out[name] = n
		}
		return out
	}

	if err := seedReferenceCatalogs(db); err != nil {
		t.Fatal(err)
	}
	first := counts()
	for name, n := range first {
		if n == 0 {
			t.Errorf("%s has no rows after --catalog-only", name)
		}
	}
	if first["risk_level"] != 5 {
		t.Errorf("risk_level rows = %d, want 5", first["risk_level"])
	}
	for name, m := range other {
		var n int64
		if err := db.Model(m).Count(&n).Error; err != nil {
			t.Fatalf("count %s: %v", name, err)
		}
		if n != 0 {
			t.Errorf("%s has %d rows; --catalog-only must not seed it", name, n)
		}
	}

	if err := seedReferenceCatalogs(db); err != nil {
		t.Fatalf("second run: %v", err)
	}
	for name, n := range counts() {
		if n != first[name] {
			t.Errorf("%s: %d rows after the second run, %d after the first", name, n, first[name])
		}
	}
}

func TestSeedHelpListsCatalogOnly(t *testing.T) {
	// scripts/onboard-tenant.sh detects support by grepping `./risk seed --help`.
	if !strings.Contains(seedCmd.Flags().FlagUsages(), "--catalog-only") {
		t.Fatal("seed --help does not list --catalog-only")
	}
}

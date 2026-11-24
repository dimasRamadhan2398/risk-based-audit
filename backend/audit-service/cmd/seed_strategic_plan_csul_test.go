package main

import (
	"regexp"
	"testing"

	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"
)

var csulNumber = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

func TestSeedStrategicPlansCSUL(t *testing.T) {
	db := sqlitetest.Open(t, &models.StrategicPlan{})

	// Generic demo row that collides with a CSUL code but has another KPI.
	if err := db.Create(&models.StrategicPlan{Code: "SO-IA01", KPI: "Generic KPI"}).Error; err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ { // second run proves idempotency
		if err := seedStrategicPlansCSUL(db, true); err != nil {
			t.Fatal(err)
		}
	}

	var rows []models.StrategicPlan
	if err := db.Order("code, kpi").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 11 {
		t.Fatalf("want 11 KPI rows, got %d", len(rows))
	}
	so := map[string]bool{}
	for _, r := range rows {
		so[r.Code] = true
		if r.Initiative == "" || len(r.InitiativeYears) != 5 || len(r.KPITargets) != 5 {
			t.Errorf("%s %q: incomplete row", r.Code, r.KPI)
		}
		for y, v := range r.KPITargets {
			if v != "" && !csulNumber.MatchString(v) {
				t.Errorf("target %s=%q is not a plain number", y, v)
			}
		}
	}
	if len(so) != 9 {
		t.Fatalf("want 9 SO, got %d", len(so))
	}

	var gic models.StrategicPlan
	if err := db.Where("kpi LIKE ?", "Tingkat kesesuaian (conformance)%").First(&gic).Error; err != nil {
		t.Fatal(err)
	}
	if gic.KPITargets["2029"] != "75.0" || gic.KPITargets["2026"] != "" || gic.TargetNotes["2029"] != "PC" {
		t.Errorf("KPI 9 targets/notes wrong: %v %v", gic.KPITargets, gic.TargetNotes)
	}
	if gic.InitiativeYears["2029"] != "√" || gic.InitiativeYears["2028"] != "-" {
		t.Errorf("KPI 9 initiative years wrong: %v", gic.InitiativeYears)
	}
	var hib models.StrategicPlan
	db.Where("kpi = ?", "Tingkat efisiensi biaya audit internal").First(&hib)
	if hib.HibHig != "HIB" || hib.KPITargets["2030"] != "62.0" {
		t.Errorf("KPI 4 wrong: %+v", hib)
	}
}

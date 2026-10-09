package services

import (
	"context"
	"testing"

	"risk-service/models"

	"github.com/google/uuid"
)

// seedYear gives a fake risk an assessment row for one year.
func seedYear(repo *fakeRepo, id uuid.UUID, year int, name string, impact int) {
	repo.assessments[id] = append(repo.assessments[id], models.RiskAssessment{
		ID: uuid.New(), RiskRegisterID: id, Year: year, RiskEvent: name,
		ImpactQ1: impact, ImpactQ2: impact, ImpactQ3: impact, ImpactQ4: impact,
	})
}

func assessmentFor(t *testing.T, repo *fakeRepo, id uuid.UUID, year int) models.RiskAssessment {
	t.Helper()
	for _, a := range repo.assessments[id] {
		if a.Year == year {
			return a
		}
	}
	t.Fatalf("no %d assessment", year)
	return models.RiskAssessment{}
}

func TestUpdatePastYearKeepsOtherYearsAndLatestVersion(t *testing.T) {
	repo := newFakeRepo()
	id := repo.addRisk("Fraud", models.RiskProfile{Category: "Financial"})
	seedYear(repo, id, 2025, "Fraud", 2)
	seedYear(repo, id, 2026, "Fraud", 4)
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

	// A stale client also sends a 2026 row; it must be ignored.
	res, err := svc.Update(context.Background(), id, &RiskRequest{
		Year: 2025, Name: "Fraud (2025 wording)", Category: "Compliance", Impact: 5, Likelihood: 5,
		Assessments: []RiskAssessmentReq{flatQuarters(2025, 3, 3), flatQuarters(2026, 1, 1)},
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	y25 := assessmentFor(t, repo, id, 2025)
	if y25.RiskEvent != "Fraud (2025 wording)" || y25.Category != "Compliance" || y25.ImpactQ1 != 3 || y25.SnapshotAt == nil {
		t.Errorf("2025 snapshot = %+v", y25)
	}
	y26 := assessmentFor(t, repo, id, 2026)
	if y26.RiskEvent != "Fraud" || y26.ImpactQ1 != 4 {
		t.Errorf("editing 2025 rewrote 2026: %+v", y26)
	}
	if reg := repo.registers[id]; reg.RiskEvent != "Fraud" || reg.Profile.Category != "Financial" || repo.profileSaves != 0 {
		t.Errorf("editing a past year moved the latest version: %q / %q", reg.RiskEvent, reg.Profile.Category)
	}
	if res.Name != "Fraud" || res.Category != "Financial" {
		t.Errorf("response reports %q / %q, want the latest version", res.Name, res.Category)
	}
}

func TestUpdateLatestYearMovesLatestVersion(t *testing.T) {
	repo := newFakeRepo()
	id := repo.addRisk("Fraud", models.RiskProfile{Category: "Financial"})
	seedYear(repo, id, 2025, "Fraud", 2)
	seedYear(repo, id, 2026, "Fraud", 4)
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

	if _, err := svc.Update(context.Background(), id, &RiskRequest{Year: 2026, Name: "Fraud v2", Category: "Compliance", LocationID: bali.ID.String()}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if reg := repo.registers[id]; reg.RiskEvent != "Fraud v2" || reg.Profile.Category != "Compliance" {
		t.Errorf("latest version = %q / %q", reg.RiskEvent, reg.Profile.Category)
	}
	y26 := assessmentFor(t, repo, id, 2026)
	if y26.LocationID == nil || *y26.LocationID != bali.ID || y26.LocationName != "Bali Branch" {
		t.Errorf("2026 location snapshot = %v %q", y26.LocationID, y26.LocationName)
	}
	if y25 := assessmentFor(t, repo, id, 2025); y25.RiskEvent != "Fraud" || y25.LocationID != nil {
		t.Errorf("editing 2026 rewrote 2025: %+v", y25)
	}
}

func TestUpdatePastYearLocationMatchingLatestIsStillApplied(t *testing.T) {
	repo := newFakeRepo()
	baliID, jakartaID := bali.ID, jakarta.ID
	id := repo.addRisk("Fraud", models.RiskProfile{LocationID: &baliID, LocationName: "Bali Branch"})
	seedYear(repo, id, 2025, "Fraud", 2)
	repo.assessments[id][0].LocationID, repo.assessments[id][0].LocationName = &jakartaID, "Jakarta Branch"
	seedYear(repo, id, 2026, "Fraud", 4)
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

	// Bali is the latest location, but 2025 read Jakarta: this is a change.
	if _, err := svc.Update(context.Background(), id, &RiskRequest{Year: 2025, Name: "Fraud", LocationID: bali.ID.String()}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if y25 := assessmentFor(t, repo, id, 2025); y25.LocationID == nil || *y25.LocationID != bali.ID {
		t.Errorf("2025 location = %v, want Bali", y25.LocationID)
	}
}

func TestUpdateNewYearStartsFromLatestVersion(t *testing.T) {
	repo := newFakeRepo()
	baliID := bali.ID
	id := repo.addRisk("Fraud", models.RiskProfile{LocationID: &baliID, LocationName: "Bali Branch"})
	seedYear(repo, id, 2026, "Fraud", 4)
	svc := NewRiskService(repo, &fakeLocations{err: errDown})

	if _, err := svc.Update(context.Background(), id, &RiskRequest{Year: 2027, Name: "Fraud", Impact: 2, Likelihood: 3}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	y27 := assessmentFor(t, repo, id, 2027)
	if y27.ImpactQ4 != 2 || y27.LikelihoodQ4 != 3 || y27.LocationID == nil || *y27.LocationID != bali.ID {
		t.Errorf("2027 row = %+v", y27)
	}
}

func TestCreateSnapshotsEachAssessment(t *testing.T) {
	repo := newFakeRepo()
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

	res, err := svc.Create(context.Background(), &RiskRequest{
		Name: "Fraud", Category: "Financial", Description: "d", LocationID: jakarta.ID.String(),
		Assessments: []RiskAssessmentReq{flatQuarters(2026, 3, 3)},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	a := res.Assessments[0]
	if a.Name != "Fraud" || a.Category != "Financial" || deref(a.LocationID) != jakarta.ID.String() || deref(a.Branch) != "Jakarta Branch" {
		t.Errorf("assessment snapshot = %+v", a)
	}
}

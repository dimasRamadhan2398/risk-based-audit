package fieldwork

import (
	"testing"

	"audit-service/models"

	"github.com/google/uuid"
)

// Two LHAs of one assignment letter (and an LHA that references the same
// letter by UUID) must not repeat its test controls in the total. Each LHA
// item keeps the letter's count.
func TestLoadRecommendationSummarySharedLetter(t *testing.T) {
	db := newDB(t)
	letter := models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-050/A/2026", AuditYear: "2026"}
	mustCreate(t, db, &letter)

	r1 := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letter.LetterNumber}
	r2 := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letter.LetterNumber}
	r3 := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letter.ID.String()}
	for _, r := range []*models.AuditResultReport{r1, r2, r3} {
		mustCreate(t, db, r)
	}

	tc := func(key, rec string) {
		mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: key, Recommendation: rec})
	}
	tc(letter.LetterNumber, "A")
	tc(letter.LetterNumber, "B")
	tc(letter.ID.String(), "C")

	ids := []string{r1.ID.String(), r2.ID.String(), r3.ID.String()}
	s, err := LoadRecommendationSummary(db, ids)
	if err != nil {
		t.Fatal(err)
	}
	for i, it := range s.Items {
		if it.Count != 3 {
			t.Errorf("item %d count = %d, want 3 (the letter's count per LHA)", i, it.Count)
		}
	}
	if s.Total != 3 {
		t.Errorf("total = %d, want 3: one letter, three recommendations, counted once", s.Total)
	}

	// Asking for the same letter's LHA twice in one request changes nothing
	s, err = LoadRecommendationSummary(db, []string{r1.ID.String(), r1.ID.String()})
	if err != nil {
		t.Fatal(err)
	}
	if s.Total != 3 {
		t.Errorf("repeated LHA total = %d, want 3", s.Total)
	}
}

// Two letters: each letter's recommendations are counted once, and the total is their sum.
func TestLoadRecommendationSummaryTwoLettersSharedLHAs(t *testing.T) {
	db := newDB(t)
	la := models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-061/A/2026", AuditYear: "2026"}
	lb := models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-062/A/2026", AuditYear: "2026"}
	mustCreate(t, db, &la)
	mustCreate(t, db, &lb)

	a1 := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: la.LetterNumber}
	a2 := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: la.LetterNumber}
	b1 := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: lb.ID.String()}
	for _, r := range []*models.AuditResultReport{a1, a2, b1} {
		mustCreate(t, db, r)
	}
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: la.LetterNumber, Recommendation: "A"})
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: lb.LetterNumber, Recommendation: "B"})
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: lb.ID.String(), Recommendation: "C"})

	s, err := LoadRecommendationSummary(db, []string{a1.ID.String(), a2.ID.String(), b1.ID.String()})
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{1, 1, 2}
	for i, it := range s.Items {
		if it.Count != want[i] {
			t.Errorf("item %d count = %d, want %d", i, it.Count, want[i])
		}
	}
	if s.Total != 3 {
		t.Errorf("total = %d, want 3 (letter A once, letter B once)", s.Total)
	}
}

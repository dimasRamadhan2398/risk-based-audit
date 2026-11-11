package controllers

import (
	"math"
	"net/http"
	"strings"
	"testing"
	"time"

	"risk-service/models"

	"github.com/google/uuid"
)

func weightReqs(ws ...float64) []CorporateRiskFactorReq {
	reqs := make([]CorporateRiskFactorReq, len(ws))
	for i, w := range ws {
		reqs[i] = CorporateRiskFactorReq{StandardRiskFactorID: uuid.NewString(), Weight: w}
	}
	return reqs
}

func TestCorporateWeightsTotalIsExact(t *testing.T) {
	cases := []struct {
		name    string
		weights []float64
		want    int64
		ok      bool
	}{
		{"exactly 100", []float64{20, 80}, 10000, true},
		{"thirds summing to 100", []float64{33.33, 33.33, 33.34}, 10000, true},
		{"float drift tenths", []float64{0.1, 0.2, 99.7}, 10000, true},
		{"single 100", []float64{100}, 10000, true},
		{"99.99 rejected", []float64{33.33, 33.33, 33.33}, 9999, false},
		{"100.01 rejected", []float64{50, 50.01}, 10001, false},
		{"99.9 rejected", []float64{49.9, 50}, 9990, false},
		{"100.1 rejected", []float64{50.1, 50}, 10010, false},
		{"empty rejected", nil, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := corporateWeightsTotalIsExact(weightReqs(tc.weights...))
			if got != tc.want || ok != tc.ok {
				t.Fatalf("corporateWeightsTotalIsExact(%v) = (%d, %v), want (%d, %v)", tc.weights, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestUpdateCorporateRiskFactorsRejectsInexactTotal(t *testing.T) {
	for _, tc := range []struct {
		name   string
		a, b   float64
		suffix string
	}{
		{"99.99", 49.99, 50, "(currently: 99.99%)"},
		{"100.01", 50.01, 50, "(currently: 100.01%)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newScoringFixture(t)
			w := f.do(t, http.MethodPost, "/risk-factors/corporate", []map[string]any{
				{"standard_risk_factor_id": f.stdA.ID.String(), "weight": tc.a},
				{"standard_risk_factor_id": f.stdB.ID.String(), "weight": tc.b},
			})
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
			}
			body := w.Body.String()
			if !strings.Contains(body, `"success":false`) || !strings.Contains(body, "Total weight must equal exactly 100% "+tc.suffix) {
				t.Fatalf("unexpected error body: %s", body)
			}
			// Weights must be untouched (still 50/50).
			var cfs []models.CorporateRiskFactor
			f.db.Find(&cfs)
			for _, cf := range cfs {
				if math.Abs(cf.Weight-0.5) > 1e-9 {
					t.Fatalf("weight changed on rejected request: %+v", cfs)
				}
			}
		})
	}
}

func TestUpdateCorporateRiskFactorsAcceptsThirds(t *testing.T) {
	f := newScoringFixture(t)
	stdC := models.StandardRiskFactor{ID: uuid.New(), Name: "HSE", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mustCreate(t, f.db, &stdC)

	w := f.do(t, http.MethodPost, "/risk-factors/corporate", []map[string]any{
		{"standard_risk_factor_id": f.stdA.ID.String(), "weight": 33.33},
		{"standard_risk_factor_id": f.stdB.ID.String(), "weight": 33.33},
		{"standard_risk_factor_id": stdC.ID.String(), "weight": 33.34},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
}

package annualplan_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ctrlAnnualPlan "audit-service/controllers/annual_plan"
	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func statusRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, &models.ActivityCodeSequence{}, &models.AuditAnnual{})
	db.Logger = logger.Discard
	r := gin.New()
	g := r.Group("/annual-audit-plans")
	g.POST("", crud.CreateWithHook(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }, ctrlAnnualPlan.AssignCodeAndValidateStatusOnCreate))
	g.PUT("/:id", crud.UpdateWithHook(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }, ctrlAnnualPlan.ValidateStatusOnUpdate))
	return r, db
}

// seedPlan inserts a plan with a given stored status, bypassing the hooks.
func seedPlan(t *testing.T, db *gorm.DB, status string) models.AuditAnnual {
	t.Helper()
	plan := models.AuditAnnual{Year: 2026, Status: status}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("seed plan with status %q: %v", status, err)
	}
	return plan
}

// putStatus sends PUT with {"status": <raw JSON value>} and returns the HTTP
// code and the stored status afterwards.
func putStatus(t *testing.T, r *gin.Engine, db *gorm.DB, id, rawJSON string) (int, string) {
	t.Helper()
	body := fmt.Sprintf(`{"status":%s}`, rawJSON)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/annual-audit-plans/"+id, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var stored models.AuditAnnual
	if err := db.Unscoped().First(&stored, "id = ?", id).Error; err != nil {
		t.Fatalf("reload plan: %v", err)
	}
	return w.Code, stored.Status
}

var canonicalStatuses = []string{
	models.AnnualPlanStatusDraft,
	models.AnnualPlanStatusPendingReview,
	models.AnnualPlanStatusPendingApproval,
	models.AnnualPlanStatusApproved,
	models.AnnualPlanStatusRevised,
}

// allowed is the transition table from the agreed workflow. Same-status
// updates are always allowed and are not listed.
var allowed = map[string][]string{
	models.AnnualPlanStatusDraft:           {models.AnnualPlanStatusPendingReview},
	models.AnnualPlanStatusPendingReview:   {models.AnnualPlanStatusPendingApproval, models.AnnualPlanStatusRevised},
	models.AnnualPlanStatusPendingApproval: {models.AnnualPlanStatusApproved, models.AnnualPlanStatusRevised},
	models.AnnualPlanStatusRevised:         {models.AnnualPlanStatusPendingReview},
	models.AnnualPlanStatusApproved:        {models.AnnualPlanStatusRevised},
}

func TestStatusTransitionsEnforcedOnUpdate(t *testing.T) {
	r, db := statusRouter(t)
	for _, from := range canonicalStatuses {
		for _, to := range canonicalStatuses {
			if from == to {
				continue
			}
			want := http.StatusConflict
			for _, ok := range allowed[from] {
				if ok == to {
					want = http.StatusOK
				}
			}
			plan := seedPlan(t, db, from)
			code, stored := putStatus(t, r, db, plan.ID.String(), fmt.Sprintf("%q", to))
			if code != want {
				t.Errorf("%s -> %s: HTTP %d, want %d", from, to, code, want)
				continue
			}
			if want == http.StatusOK && stored != to {
				t.Errorf("%s -> %s: stored %q", from, to, stored)
			}
			if want == http.StatusConflict && stored != from {
				t.Errorf("%s -> %s rejected but stored %q", from, to, stored)
			}
		}
	}
}

func TestSameStatusUpdateAccepted(t *testing.T) {
	r, db := statusRouter(t)
	plan := seedPlan(t, db, models.AnnualPlanStatusApproved)
	if code, stored := putStatus(t, r, db, plan.ID.String(), `"Approved"`); code != http.StatusOK || stored != "Approved" {
		t.Fatalf("same-status update: HTTP %d, stored %q", code, stored)
	}
}

func TestLegacyStatusInUpdateIsNormalizedBeforeCheck(t *testing.T) {
	r, db := statusRouter(t)
	plan := seedPlan(t, db, models.AnnualPlanStatusPendingApproval)
	// "Done" is the legacy form of Approved: a valid transition from Pending Approval.
	if code, stored := putStatus(t, r, db, plan.ID.String(), `"Done"`); code != http.StatusOK || stored != "Approved" {
		t.Fatalf("legacy Done: HTTP %d, stored %q", code, stored)
	}

	plan = seedPlan(t, db, models.AnnualPlanStatusApproved)
	// "Work In Progress" normalizes to Pending Review, which Approved cannot reach.
	if code, stored := putStatus(t, r, db, plan.ID.String(), `"Work In Progress"`); code != http.StatusConflict || stored != "Approved" {
		t.Fatalf("legacy WIP from Approved: HTTP %d, stored %q", code, stored)
	}
}

func TestUnknownStatusRejectedOnUpdate(t *testing.T) {
	r, db := statusRouter(t)
	plan := seedPlan(t, db, models.AnnualPlanStatusDraft)
	for _, raw := range []string{`"CLOSED"`, `"approved"`, `42`} {
		code, stored := putStatus(t, r, db, plan.ID.String(), raw)
		if code != http.StatusBadRequest || stored != "Draft" {
			t.Errorf("status %s: HTTP %d, stored %q; want 400 and Draft", raw, code, stored)
		}
	}
}

func TestUpdateWithoutStatusIsNotAffected(t *testing.T) {
	r, db := statusRouter(t)
	plan := seedPlan(t, db, models.AnnualPlanStatusApproved)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/annual-audit-plans/"+plan.ID.String(), strings.NewReader(`{"notes":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("notes-only update: HTTP %d", w.Code)
	}
}

func TestCreateStartsInDraft(t *testing.T) {
	r, _ := statusRouter(t)
	for _, status := range []string{`""`, `"Not Available"`, `"DRAFT"`, `"Draft"`, `null`} {
		body := fmt.Sprintf(`{"year":2026,"status":%s}`, status)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/annual-audit-plans", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var env struct {
			Data models.AuditAnnual `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &env)
		if w.Code != http.StatusCreated || env.Data.Status != "Draft" {
			t.Errorf("create with status %s: HTTP %d, status %q", status, w.Code, env.Data.Status)
		}
	}
}

func TestCreateRejectsNonDraftStatus(t *testing.T) {
	r, _ := statusRouter(t)
	for _, status := range []string{`"Approved"`, `"Pending Review"`, `"Done"`} {
		body := fmt.Sprintf(`{"year":2026,"status":%s}`, status)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/annual-audit-plans", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusConflict {
			t.Errorf("create with status %s: HTTP %d, want 409", status, w.Code)
		}
	}
}

func TestNormalizeStoredStatuses(t *testing.T) {
	_, db := statusRouter(t)
	legacy := map[string]string{
		"":                 models.AnnualPlanStatusDraft,
		"Not Available":    models.AnnualPlanStatusDraft,
		"DRAFT":            models.AnnualPlanStatusDraft,
		"Work In Progress": models.AnnualPlanStatusPendingReview,
		"SUBMITTED":        models.AnnualPlanStatusPendingReview,
		"Done":             models.AnnualPlanStatusApproved,
		"APPROVED":         models.AnnualPlanStatusApproved,
		"REJECTED":         models.AnnualPlanStatusRevised,
		"REVISED":          models.AnnualPlanStatusRevised,
		"PENDING_APPROVAL": models.AnnualPlanStatusPendingApproval,
	}
	ids := map[string]string{}
	for raw, want := range legacy {
		plan := seedPlan(t, db, raw)
		if raw == "" {
			// GORM and the column default turn an empty status into "Draft" on
			// insert, so write the empty string explicitly.
			if err := db.Exec("UPDATE audit_annuals SET status = '' WHERE id = ?", plan.ID).Error; err != nil {
				t.Fatalf("set empty status: %v", err)
			}
		}
		ids[plan.ID.String()] = want
	}
	// A soft-deleted legacy row is normalized too.
	deleted := seedPlan(t, db, "Done")
	if err := db.Delete(&deleted).Error; err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	ids[deleted.ID.String()] = models.AnnualPlanStatusApproved
	// A canonical row is untouched.
	kept := seedPlan(t, db, models.AnnualPlanStatusPendingApproval)
	ids[kept.ID.String()] = models.AnnualPlanStatusPendingApproval

	n, unknown, err := ctrlAnnualPlan.NormalizeStoredStatuses(db)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if len(unknown) != 0 {
		t.Errorf("unknown statuses: %v", unknown)
	}
	if n != int64(len(legacy)+1) {
		t.Errorf("rewritten %d rows, want %d", n, len(legacy)+1)
	}
	for id, want := range ids {
		var got models.AuditAnnual
		if err := db.Unscoped().First(&got, "id = ?", id).Error; err != nil {
			t.Fatalf("reload %s: %v", id, err)
		}
		if got.Status != want {
			t.Errorf("plan %s: status %q, want %q", id, got.Status, want)
		}
	}

	// Second run is a no-op.
	if n, _, err := ctrlAnnualPlan.NormalizeStoredStatuses(db); err != nil || n != 0 {
		t.Errorf("second run: rewritten %d, err %v; want 0, nil", n, err)
	}
}

func TestNormalizeStoredStatusesReportsUnknownValues(t *testing.T) {
	_, db := statusRouter(t)
	seedPlan(t, db, "CLOSED")
	_, unknown, err := ctrlAnnualPlan.NormalizeStoredStatuses(db)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if len(unknown) != 1 || unknown[0] != "CLOSED" {
		t.Errorf("unknown = %v, want [CLOSED]", unknown)
	}
}

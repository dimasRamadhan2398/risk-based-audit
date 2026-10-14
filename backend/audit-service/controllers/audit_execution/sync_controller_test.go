package audit_execution

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestSyncFromSources_ReportsUnmatchedRows(t *testing.T) {
	logger.Log = zap.NewNop()
	gin.SetMode(gin.TestMode)

	db := testdb.New(t,
		&models.AuditExecution{},
		&models.AuditActivity{},
		&models.AssignmentLetter{},
		&models.FieldworkInterview{},
		&models.FieldworkObservation{},
		&models.FieldworkDocument{},
		&models.FieldworkSample{},
		&models.FieldworkTestControl{},
		&models.WorkingPaperHeader{},
		&models.WorkingPaperRisk{},
		&models.WorkingPaperSample{},
		&models.WorkingPaperCause{},
		&models.WorkingPaperPlan{},
		&models.AuditResultReport{},
	)
	if err := db.Create(&models.AuditExecution{ID: uuid.New(), Ref: "ST-001/SKAI/2026", Name: "Audit A"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: "ST-404/SKAI/2026", Interviewee: "X"}).Error; err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.POST("/audit-executions/sync", NewSyncController(db).SyncFromSources)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/audit-executions/sync", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Data struct {
			Synced              int              `json:"synced"`
			UnmatchedSourceRows map[string]int64 `json:"unmatched_source_rows"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v; body = %s", err, w.Body.String())
	}
	if body.Data.Synced != 1 {
		t.Errorf("synced = %d, want 1", body.Data.Synced)
	}
	if body.Data.UnmatchedSourceRows["fieldwork_interviews"] != 1 {
		t.Errorf("unmatched_source_rows = %v, want fieldwork_interviews: 1", body.Data.UnmatchedSourceRows)
	}
}

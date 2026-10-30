package department

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"master-service/pkg/validations"
	departmentSvc "master-service/services/department"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func legacyRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	ctrl := NewDepartmentController(departmentSvc.NewDepartmentService(&fakeRepo{}, nil), validations.New())
	r := gin.New()
	r.GET("/departments/:id", ctrl.FindById)
	r.POST("/departments", ctrl.Create)
	r.PUT("/departments/:id", ctrl.Update)
	return r
}

// Legacy tenants refuse a parent instead of silently ignoring it.
func TestLegacyRejectsParentDepartmentID(t *testing.T) {
	r := legacyRouter()
	payload := fmt.Sprintf(`{"department_code":"FIN","department_name":"Finance","level":1,"company_id":"%s","parent_department_id":"%s"}`,
		uuid.NewString(), uuid.NewString())
	status, b, raw := call(t, r, http.MethodPost, "/departments", payload)
	if status != http.StatusUnprocessableEntity || b.Code != "ORG_STRUCTURE_NOT_ENABLED" {
		t.Fatalf("create: %d %s", status, raw)
	}
	status, b, raw = call(t, r, http.MethodPut, "/departments/"+uuid.NewString(), fmt.Sprintf(`{"parent_department_id":"%s"}`, uuid.NewString()))
	if status != http.StatusUnprocessableEntity || b.Code != "ORG_STRUCTURE_NOT_ENABLED" {
		t.Fatalf("update: %d %s", status, raw)
	}
	// null / absent are fine (no parent is the legacy meaning).
	status, _, raw = call(t, r, http.MethodPut, "/departments/"+uuid.NewString(), `{"parent_department_id":null}`)
	if status != http.StatusOK {
		t.Fatalf("update with null parent: %d %s", status, raw)
	}
}

// The legacy department JSON carries none of the flexible keys.
func TestLegacyResponseHasNoOrgKeys(t *testing.T) {
	_, _, raw := call(t, legacyRouter(), http.MethodGet, "/departments/"+uuid.NewString(), "")
	for _, k := range []string{"parent_department_id", "level_name", "name_key", `"path"`} {
		if strings.Contains(raw, k) {
			t.Fatalf("legacy response contains %s: %s", k, raw)
		}
	}
}

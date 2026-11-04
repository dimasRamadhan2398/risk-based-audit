package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/models"
	"auth-service/pkg/utils"
	authService "auth-service/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// fakeAuthService overrides only the methods /auth/me uses.
type fakeAuthService struct {
	authService.AuthServiceInterface
	link models.EmployeeLink
	err  error
}

func (f *fakeAuthService) ValidateToken(string) (*utils.Claims, error) {
	return &utils.Claims{UserID: uuid.New().String(), Username: "auditor1", Roles: []string{"AUDITOR"}}, nil
}

func (f *fakeAuthService) ResolveEmployeeLink(context.Context, uuid.UUID, string) (models.EmployeeLink, error) {
	return f.link, f.err
}

func meResponse(t *testing.T, svc authService.AuthServiceInterface) map[string]interface{} {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/auth/me", NewAuthController(nil, svc).Me)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer tok")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body.Data
}

func TestMe_IncludesEmployeeLinkFields(t *testing.T) {
	code, empID, deptID := "EMP001", "6f1c1b1e-0000-4000-8000-000000000001", "6f1c1b1e-0000-4000-8000-000000000002"
	data := meResponse(t, &fakeAuthService{link: models.EmployeeLink{EmployeeCode: &code, EmployeeID: &empID, DepartmentID: &deptID}})

	if data["employee_code"] != code || data["employee_id"] != empID || data["department_id"] != deptID {
		t.Fatalf("link fields = %v / %v / %v", data["employee_code"], data["employee_id"], data["department_id"])
	}
	if data["username"] != "auditor1" {
		t.Fatalf("existing field username changed: %v", data["username"])
	}
}

func TestMe_UnlinkedOrMasterDownReturnsNullsNotError(t *testing.T) {
	code := "EMP002"
	data := meResponse(t, &fakeAuthService{link: models.EmployeeLink{EmployeeCode: &code}})
	if data["employee_code"] != code || data["employee_id"] != nil || data["department_id"] != nil {
		t.Fatalf("partial link = %v", data)
	}

	data = meResponse(t, &fakeAuthService{err: context.DeadlineExceeded})
	for _, k := range []string{"employee_code", "employee_id", "department_id"} {
		if v, ok := data[k]; !ok || v != nil {
			t.Fatalf("%s = %v, want null", k, v)
		}
	}
}

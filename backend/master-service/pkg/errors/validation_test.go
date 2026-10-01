package errors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type bindReq struct {
	DepartmentCode string `json:"department_code" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Level          int    `json:"level" binding:"required,min=1"`
	Status         string `json:"status" binding:"omitempty,oneof=A B"`
}

func bind(t *testing.T, body string) *AppError {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	var req bindReq
	err := c.ShouldBindJSON(&req)
	if err == nil {
		return nil
	}
	return FromBinding(err, &req)
}

func TestFromBinding_ValidatorErrorsUseJSONNames(t *testing.T) {
	got := bind(t, `{"email":"not-an-email","level":0,"status":"Z"}`)
	if got == nil {
		t.Fatal("expected error")
	}
	if got.StatusCode != 400 || got.Code != CodeValidationFailed || got.Message != MsgValidationFailed {
		t.Fatalf("got %d %s %q", got.StatusCode, got.Code, got.Message)
	}
	want := map[string]string{
		"department_code": FieldRequired,
		"email":           FieldInvalidFormat,
		"level":           FieldRequired, // 0 fails `required` first
		"status":          FieldInvalidOption,
	}
	for k, v := range want {
		if got.Fields[k] != v {
			t.Fatalf("fields[%s] = %q, want %q (all: %v)", k, got.Fields[k], v, got.Fields)
		}
	}
	if strings.Contains(got.Message, "Key:") || strings.Contains(got.Message, "bindReq") {
		t.Fatalf("message leaks validator text: %q", got.Message)
	}
}

func TestFromBinding_BodyErrors(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantCode  string
		wantField string
	}{
		{"empty body", ``, CodeInvalidRequestBody, ""},
		{"malformed json", `{"department_code":`, CodeInvalidRequestBody, ""},
		{"syntax error", `{department_code}`, CodeInvalidRequestBody, ""},
		{"wrong type", `{"department_code":"D1","email":"a@b.co","level":"high"}`, CodeValidationFailed, "level"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bind(t, tt.body)
			if got == nil || got.StatusCode != 400 || got.Code != tt.wantCode {
				t.Fatalf("got %+v, want 400 %s", got, tt.wantCode)
			}
			if tt.wantField != "" && got.Fields[tt.wantField] != FieldInvalidType {
				t.Fatalf("fields = %v", got.Fields)
			}
			if strings.Contains(got.Message, "json:") || strings.Contains(got.Message, "unexpected") {
				t.Fatalf("message leaks parser text: %q", got.Message)
			}
		})
	}
}

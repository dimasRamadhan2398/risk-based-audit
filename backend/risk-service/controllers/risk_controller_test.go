package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"risk-service/repositories"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// stubService returns a fixed error from writes and fixed data from GetAll.
type stubService struct {
	err  error
	list []services.RiskResponse
}

func (s stubService) GetAll(context.Context) ([]services.RiskResponse, error) { return s.list, s.err }
func (s stubService) List(context.Context, repositories.RiskFilter) (*services.RiskPage, error) {
	return &services.RiskPage{Data: s.list, Page: 1, PageSize: len(s.list), Total: int64(len(s.list)), TotalPages: 1}, s.err
}
func (s stubService) Categories(context.Context, int) ([]string, error) { return nil, s.err }
func (s stubService) Create(context.Context, *services.RiskRequest) (*services.RiskResponse, error) {
	return nil, s.err
}
func (s stubService) Update(context.Context, uuid.UUID, *services.RiskRequest) (*services.RiskResponse, error) {
	return nil, s.err
}
func (s stubService) Delete(uuid.UUID) error { return s.err }

func router(svc services.IRiskService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := NewRiskController(svc)
	r.GET("/api/v1/risks", ctrl.ListRisks)
	r.POST("/api/v1/risks", ctrl.CreateRisk)
	r.PUT("/api/v1/risks/:id", ctrl.UpdateRisk)
	return r
}

func TestWriteErrorMapping(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantErr  string
	}{
		{"unregistered location is 400", &services.LocationError{Code: "UNKNOWN_LOCATION", Message: `branch "Medan Branch" is not a registered location`}, http.StatusBadRequest, "UNKNOWN_LOCATION"},
		{"master down is 503", fmt.Errorf("%w: dial tcp", services.ErrLocationsUnavailable), http.StatusServiceUnavailable, "LOCATION_MASTER_UNAVAILABLE"},
		{"other errors stay 500", errors.New("db gone"), http.StatusInternalServerError, "DB_ERROR"},
	}
	for _, tt := range tests {
		for _, req := range []struct{ method, path string }{
			{http.MethodPost, "/api/v1/risks"},
			{http.MethodPut, "/api/v1/risks/" + uuid.NewString()},
		} {
			t.Run(tt.name+" "+req.method, func(t *testing.T) {
				w := httptest.NewRecorder()
				r := httptest.NewRequest(req.method, req.path, strings.NewReader(`{"name":"R","branch":"Medan Branch"}`))
				r.Header.Set("Content-Type", "application/json")
				router(stubService{err: tt.err}).ServeHTTP(w, r)

				if w.Code != tt.wantCode {
					t.Fatalf("status = %d, want %d; body %s", w.Code, tt.wantCode, w.Body)
				}
				var body struct {
					Error struct{ Code string } `json:"error"`
				}
				_ = json.Unmarshal(w.Body.Bytes(), &body)
				if body.Error.Code != tt.wantErr {
					t.Errorf("error.code = %q, want %q", body.Error.Code, tt.wantErr)
				}
			})
		}
	}
}

func TestListRisksEmitsNullBranch(t *testing.T) {
	id, name := "7d1000c0-0b66-468b-978d-09e32c8d9df6", "Head Office"
	svc := stubService{list: []services.RiskResponse{
		{ID: "a", LocationID: &id, Branch: &name, Assessments: []services.RiskAssessmentRes{}},
		{ID: "b", Assessments: []services.RiskAssessmentRes{}},
	}}
	w := httptest.NewRecorder()
	router(svc).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/risks", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{
		`"location_id":"7d1000c0-0b66-468b-978d-09e32c8d9df6"`,
		`"branch":"Head Office"`,
		`"location_id":null`,
		`"branch":null`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response missing %s: %s", want, body)
		}
	}
}

func listRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := NewRiskController(stubService{list: []services.RiskResponse{{ID: "1"}}})
	r.GET("/api/v1/risks", ctrl.ListRisks)
	r.GET("/api/v1/risks/categories", ctrl.ListCategories)
	return r
}

func TestListRisksPagination(t *testing.T) {
	get := func(url string) (int, map[string]any) {
		w := httptest.NewRecorder()
		listRouter().ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
		var body map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		return w.Code, body
	}

	code, body := get("/api/v1/risks")
	if code != 200 || body["pagination"] != nil || len(body["data"].([]any)) != 1 {
		t.Errorf("no params must keep the unpaged shape: %d %v", code, body)
	}
	code, body = get("/api/v1/risks?page=1")
	if code != 200 || body["pagination"] == nil {
		t.Errorf("page=1: %d %v", code, body)
	}
	code, body = get("/api/v1/risks?page_size=500")
	if code != 200 {
		t.Errorf("page_size is clamped, not rejected: %d", code)
	}
	for _, bad := range []string{"page=abc", "page=0", "page=-1", "page_size=x", "fiscal_year=abc", "branch_id=nope", "period=q9", "sort_by=id%3Bdrop", "sort_dir=up"} {
		code, body = get("/api/v1/risks?" + bad)
		errObj, _ := body["error"].(map[string]any)
		if code != 400 || errObj["code"] != "INVALID_QUERY" || body["success"] != false {
			t.Errorf("%s: %d %v", bad, code, body)
		}
	}
	code, body = get("/api/v1/risks/categories")
	if code != 200 || body["data"] == nil {
		t.Errorf("categories: %d %v", code, body)
	}
	if code, _ = get("/api/v1/risks/categories?fiscal_year=x"); code != 400 {
		t.Errorf("categories bad year: %d", code)
	}
}

func TestParseRiskFilterClampsPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := map[string]int{"page=2": 10, "page_size=500": 100, "page_size=0": 1, "page_size=-5": 1, "page_size=25": 25}
	for q, want := range cases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/?"+q, nil)
		f, msg := parseRiskFilter(c)
		if msg != "" || !f.Paged || f.PageSize != want {
			t.Errorf("%s: %+v %q", q, f, msg)
		}
	}
}

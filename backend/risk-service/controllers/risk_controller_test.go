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

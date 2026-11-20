package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"risk-service/models"
	"risk-service/pkg/sqlitetest"
	"risk-service/services"

	"github.com/gin-gonic/gin"
)

func pofCatalogueEngine(t *testing.T, seed bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, models.CosoPofModels()...)
	if seed {
		if err := services.SeedCosoPofCatalogue(db); err != nil {
			t.Fatal(err)
		}
	}
	ctrl := &RCMController{db: db}
	e := gin.New()
	e.GET("/api/v1/rcm/coso-pof-catalogue", ctrl.GetCosoPofCatalogue)
	return e
}

func TestGetCosoPofCatalogue(t *testing.T) {
	e := pofCatalogueEngine(t, true)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/rcm/coso-pof-catalogue", nil))
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var out struct {
		Success bool `json:"success"`
		Data    struct {
			Catalogue services.PofCatalogue `json:"catalogue"`
			Counts    map[string]int        `json:"counts"`
			Owners    []map[string]string   `json:"owner_roles"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !out.Success || out.Data.Catalogue.Version != services.CosoPofCatalogVersion || !out.Data.Catalogue.IsCurrent {
		t.Fatalf("unexpected body: %s", w.Body)
	}
	if len(out.Data.Catalogue.Components) != 5 || out.Data.Counts["pofs"] < 70 || len(out.Data.Owners) == 0 {
		t.Errorf("components=%d counts=%v", len(out.Data.Catalogue.Components), out.Data.Counts)
	}
}

func TestGetCosoPofCatalogueUnknownVersionAndBadFramework(t *testing.T) {
	e := pofCatalogueEngine(t, true)
	for path, want := range map[string]int{
		"/api/v1/rcm/coso-pof-catalogue?version=0.0":     http.StatusNotFound,
		"/api/v1/rcm/coso-pof-catalogue?framework=a%20b": http.StatusUnprocessableEntity,
	} {
		w := httptest.NewRecorder()
		e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != want {
			t.Errorf("%s: status %d, want %d (%s)", path, w.Code, want, w.Body)
		}
	}
	// Not seeded yet: 404 with a hint, not a 500
	e = pofCatalogueEngine(t, false)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/rcm/coso-pof-catalogue", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("unseeded: status %d", w.Code)
	}
}

package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"master-service/models"
	"master-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// fakeHub answers /sources/register with the given body (raw JSON; "" = empty object),
// computed from the request ssl_mode when bodyFn is set.
func echoEnv(t *testing.T, bodyFn func(sent string) string) (*gin.Engine, *gorm.DB) {
	t.Helper()
	r, db, _ := newDSEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var m map[string]interface{}
		b, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(b, &m)
		sent, _ := m["ssl_mode"].(string)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(bodyFn(sent)))
	}))
	t.Cleanup(srv.Close)
	dh := services.NewDataHubClient()
	dh.BaseURL = srv.URL
	r = gin.New()
	r.POST("/data-sources", dataSourceCreate(db, dh))
	r.PUT("/data-sources/:id", dataSourceUpdate(db, dh))
	return r, db
}

func createWithMode(t *testing.T, r *gin.Engine, db *gorm.DB, mode string) (map[string]interface{}, models.DataSourceConnection) {
	t.Helper()
	extra := map[string]interface{}{"sslMode": mode}
	if mode == "verify-ca" || mode == "verify-full" {
		extra["sslRootCert"] = testRootPEM
	}
	w := doJSON(r, "POST", "/data-sources", baseBody(extra))
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	resp := decode(t, w)
	id := resp["data"].(map[string]interface{})["id"].(string)
	var conn models.DataSourceConnection
	if err := db.First(&conn, "id = ?", uuid.MustParse(id)).Error; err != nil {
		t.Fatal(err)
	}
	return resp, conn
}

func TestHubEcho_Matches_StaysConnected(t *testing.T) {
	r, db := echoEnv(t, func(sent string) string { return fmt.Sprintf(`{"status":"success","ssl_mode":%q}`, sent) })
	resp, conn := createWithMode(t, r, db, "verify-full")
	if resp["dataHubRegistered"] != true || conn.Status != "Connected" || conn.LastError != "" {
		t.Fatalf("want Connected, got registered=%v status=%q err=%q", resp["dataHubRegistered"], conn.Status, conn.LastError)
	}
}

func TestHubEcho_Missing_MarksError(t *testing.T) {
	r, db := echoEnv(t, func(string) string { return `{"status":"success","message":"registered"}` })
	resp, conn := createWithMode(t, r, db, "verify-full")
	want := "Data Hub does not support or did not apply ssl_mode=verify-full; update the Data Hub first"
	if resp["dataHubRegistered"] != false || resp["dataHubError"] != want {
		t.Fatalf("registered=%v dataHubError=%v", resp["dataHubRegistered"], resp["dataHubError"])
	}
	if conn.Status != "Error" || conn.LastError != want {
		t.Fatalf("status=%q lastError=%q", conn.Status, conn.LastError)
	}
}

func TestHubEcho_Different_MarksError(t *testing.T) {
	r, db := echoEnv(t, func(string) string { return `{"status":"success","ssl_mode":"require"}` })
	resp, conn := createWithMode(t, r, db, "verify-full")
	if conn.Status != "Error" || !strings.Contains(conn.LastError, "ssl_mode=verify-full") || !strings.Contains(conn.LastError, "stored ssl_mode=require") {
		t.Fatalf("status=%q lastError=%q", conn.Status, conn.LastError)
	}
	if resp["dataHubRegistered"] != false {
		t.Fatalf("dataHubRegistered=%v", resp["dataHubRegistered"])
	}
}

func TestHubEcho_Disable_NoEchoRequired(t *testing.T) {
	r, db := echoEnv(t, func(string) string { return `{"status":"success"}` })
	resp, conn := createWithMode(t, r, db, "disable")
	if resp["dataHubRegistered"] != true || conn.Status != "Connected" || conn.LastError != "" {
		t.Fatalf("want Connected, got registered=%v status=%q err=%q", resp["dataHubRegistered"], conn.Status, conn.LastError)
	}
}

func TestHubEcho_Update_MissingMarksError(t *testing.T) {
	// Created OK against a new hub, then the hub "downgrades" (no echo) on update.
	mode := "match"
	r, db := echoEnv(t, func(sent string) string {
		if mode == "match" {
			return fmt.Sprintf(`{"ssl_mode":%q}`, sent)
		}
		return `{}`
	})
	_, conn := createWithMode(t, r, db, "require")
	mode = "old"
	w := doJSON(r, "PUT", "/data-sources/"+conn.ID.String(), baseBody(map[string]interface{}{"sslMode": "require"}))
	if w.Code != http.StatusOK {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	var after models.DataSourceConnection
	db.First(&after, "id = ?", conn.ID)
	if after.Status != "Error" || !strings.Contains(after.LastError, "ssl_mode=require") {
		t.Fatalf("status=%q lastError=%q", after.Status, after.LastError)
	}
}

func TestSyncSourceToHubSSLStrict(t *testing.T) {
	r, db := echoEnv(t, func(string) string { return `{}` })
	_ = r
	conn := models.DataSourceConnection{ID: uuid.New(), Name: "GL", Host: "h", Status: "Connected", SSLMode: "verify-ca", SSL: true, SSLRootCert: testRootPEM}
	if err := db.Create(&conn).Error; err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{}`)) }))
	defer srv.Close()
	dh := services.NewDataHubClient()
	dh.BaseURL = srv.URL
	if err := syncSourceToHubSSLStrict(db, dh, &conn); !services.IsSSLModeNotApplied(err) {
		t.Fatalf("want SSLModeNotAppliedError, got %v", err)
	}
	var after models.DataSourceConnection
	db.First(&after, "id = ?", conn.ID)
	if after.Status != "Error" || after.LastError == "" {
		t.Fatalf("status=%q lastError=%q", after.Status, after.LastError)
	}
}

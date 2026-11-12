package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"master-service/models"
	"master-service/services"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type hubCapture struct {
	mu     sync.Mutex
	bodies []map[string]interface{}
	status int
}

func (h *hubCapture) last(t *testing.T) map[string]interface{} {
	t.Helper()
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.bodies) == 0 {
		t.Fatal("Data Hub received no registration")
	}
	return h.bodies[len(h.bodies)-1]
}

func newDSEnv(t *testing.T) (*gin.Engine, *gorm.DB, *hubCapture) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	// Hand-written DDL: AutoMigrate emits gen_random_uuid() defaults that SQLite cannot parse.
	for _, ddl := range []string{
		`CREATE TABLE data_source_connections (id TEXT PRIMARY KEY, name TEXT NOT NULL, type TEXT, host TEXT NOT NULL,
			port INTEGER, database TEXT, environment TEXT, status TEXT, sync_schedule TEXT, last_sync TEXT, last_error TEXT,
			ssl BOOLEAN, ssl_mode TEXT, ssl_root_cert TEXT, username TEXT, password TEXT, scopes TEXT, data_mappings TEXT,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE data_source_activity_logs (id TEXT PRIMARY KEY, connection_id TEXT, connection_name TEXT, type TEXT,
			event TEXT, status TEXT, records INTEGER, duration TEXT, timestamp TEXT, created_at DATETIME, deleted_at DATETIME)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	hub := &hubCapture{status: http.StatusOK}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]interface{}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &m)
		hub.mu.Lock()
		hub.bodies = append(hub.bodies, m)
		st := hub.status
		hub.mu.Unlock()
		w.WriteHeader(st)
		// Behave like an up-to-date Data Hub: echo the ssl_mode it stored.
		echo, _ := json.Marshal(map[string]interface{}{"status": "success", "ssl_mode": m["ssl_mode"]})
		_, _ = w.Write(echo)
	}))
	t.Cleanup(srv.Close)
	dh := services.NewDataHubClient()
	dh.BaseURL = srv.URL

	r := gin.New()
	r.POST("/data-sources", dataSourceCreate(db, dh))
	r.PUT("/data-sources/:id", dataSourceUpdate(db, dh))
	r.POST("/data-sources/test", dataSourceTestAdhoc())
	r.POST("/data-sources/:id/test", dataSourceTestByID(db))
	return r, db, hub
}

func doJSON(r http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("bad JSON %q: %v", w.Body.String(), err)
	}
	return m
}

func baseBody(extra map[string]interface{}) map[string]interface{} {
	m := map[string]interface{}{"name": "GL", "type": "postgres", "host": "db.example", "port": 5432, "database": "gl", "username": "u", "password": "p"}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestDataSourceCreate_SSL(t *testing.T) {
	r, db, hub := newDSEnv(t)

	// Legacy boolean only -> require; hub receives all three keys with the right types.
	w := doJSON(r, "POST", "/data-sources", baseBody(map[string]interface{}{"ssl": true}))
	if w.Code != 201 {
		t.Fatalf("legacy create: %d %s", w.Code, w.Body)
	}
	p := hub.last(t)
	if p["ssl_mode"] != "require" || p["ssl_enabled"] != true || p["ssl_root_cert"] != "" {
		t.Fatalf("hub payload ssl fields: %v", p)
	}

	// Explicit verify-full with cert is stored and forwarded.
	w = doJSON(r, "POST", "/data-sources", baseBody(map[string]interface{}{"sslMode": "verify-full", "sslRootCert": testRootPEM}))
	if w.Code != 201 {
		t.Fatalf("verify-full create: %d %s", w.Code, w.Body)
	}
	p = hub.last(t)
	if p["ssl_mode"] != "verify-full" || p["ssl_enabled"] != true || p["ssl_root_cert"] != testRootPEM {
		t.Fatalf("hub payload: %v", p)
	}
	id := decode(t, w)["data"].(map[string]interface{})["id"].(string)
	var row models.DataSourceConnection
	if err := db.First(&row, "id = ?", id).Error; err != nil || row.SSLRootCert != testRootPEM || row.SSLMode != "verify-full" {
		t.Fatalf("stored row: %+v err=%v", row, err)
	}

	// disable -> ssl_enabled=false.
	doJSON(r, "POST", "/data-sources", baseBody(map[string]interface{}{"sslMode": "disable"}))
	if p = hub.last(t); p["ssl_mode"] != "disable" || p["ssl_enabled"] != false {
		t.Fatalf("disable payload: %v", p)
	}

	rejects := []struct {
		name string
		body map[string]interface{}
		msg  string
	}{
		{"unknown mode", map[string]interface{}{"sslMode": "prefer"}, "unknown ssl_mode"},
		{"cert with require", map[string]interface{}{"sslMode": "require", "sslRootCert": testRootPEM}, "only applies to verify-ca and verify-full"},
		{"cert with disable", map[string]interface{}{"sslMode": "disable", "sslRootCert": testRootPEM}, "only applies to verify-ca and verify-full"},
		{"bad PEM", map[string]interface{}{"sslMode": "verify-ca", "sslRootCert": "junk"}, ""},
	}
	var n int64
	db.Model(&models.DataSourceConnection{}).Count(&n)
	for _, tc := range rejects {
		w := doJSON(r, "POST", "/data-sources", baseBody(tc.body))
		if w.Code != 400 || !strings.Contains(decode(t, w)["message"].(string), tc.msg) {
			t.Errorf("%s: %d %s", tc.name, w.Code, w.Body)
		}
	}
	var after int64
	db.Model(&models.DataSourceConnection{}).Count(&after)
	if after != n {
		t.Fatalf("rejected requests created rows: %d -> %d", n, after)
	}
}

func TestDataSourceCreate_HubFailureMarksError(t *testing.T) {
	r, db, hub := newDSEnv(t)
	hub.status = 500
	w := doJSON(r, "POST", "/data-sources", baseBody(map[string]interface{}{"sslMode": "require"}))
	m := decode(t, w)
	if w.Code != 201 || m["dataHubRegistered"] != false || m["dataHubError"] == nil {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var row models.DataSourceConnection
	db.First(&row)
	if row.Status != "Error" {
		t.Fatalf("status = %q", row.Status)
	}
}

func TestDataSourceUpdate_SSL(t *testing.T) {
	r, db, hub := newDSEnv(t)
	w := doJSON(r, "POST", "/data-sources", baseBody(map[string]interface{}{"sslMode": "verify-ca", "sslRootCert": testRootPEM}))
	id := decode(t, w)["data"].(map[string]interface{})["id"].(string)

	// Settings UI saves without resending the cert: it is retained.
	w = doJSON(r, "PUT", "/data-sources/"+id, baseBody(map[string]interface{}{"sslMode": "verify-full", "name": "GL2"}))
	if w.Code != 200 {
		t.Fatalf("update: %d %s", w.Code, w.Body)
	}
	var row models.DataSourceConnection
	db.First(&row, "id = ?", id)
	if row.SSLRootCert != testRootPEM || row.SSLMode != "verify-full" || row.Name != "GL2" || row.Password != "p" {
		t.Fatalf("after update: mode=%q certKept=%v name=%q pw=%q", row.SSLMode, row.SSLRootCert == testRootPEM, row.Name, row.Password)
	}
	if p := hub.last(t); p["ssl_root_cert"] != testRootPEM || p["ssl_mode"] != "verify-full" {
		t.Fatalf("hub payload: %v", p)
	}

	// Cert sent explicitly with require is rejected and nothing changes.
	w = doJSON(r, "PUT", "/data-sources/"+id, baseBody(map[string]interface{}{"sslMode": "require", "sslRootCert": testRootPEM}))
	if w.Code != 400 || !strings.Contains(w.Body.String(), "only applies to verify-ca and verify-full") {
		t.Fatalf("require+cert: %d %s", w.Code, w.Body)
	}
	db.First(&row, "id = ?", id)
	if row.SSLMode != "verify-full" {
		t.Fatalf("rejected update changed mode to %q", row.SSLMode)
	}

	// Switching to require without a cert drops the stale CA.
	w = doJSON(r, "PUT", "/data-sources/"+id, baseBody(map[string]interface{}{"sslMode": "require"}))
	db.First(&row, "id = ?", id)
	if w.Code != 200 || row.SSLRootCert != "" || row.SSLMode != "require" || !row.SSL {
		t.Fatalf("to require: %d cert=%q mode=%q", w.Code, row.SSLRootCert, row.SSLMode)
	}

	// Unknown mode and unknown id.
	if w = doJSON(r, "PUT", "/data-sources/"+id, baseBody(map[string]interface{}{"sslMode": "bogus"})); w.Code != 400 {
		t.Fatalf("bogus: %d", w.Code)
	}
	if w = doJSON(r, "PUT", "/data-sources/"+uuid.NewString(), baseBody(nil)); w.Code != 404 {
		t.Fatalf("missing: %d", w.Code)
	}
}

// fakePG listens on loopback and answers the PostgreSQL SSLRequest with 'N'.
func fakePGRefusingSSL(t *testing.T) (string, int) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				buf := make([]byte, 8)
				if _, err := io.ReadFull(c, buf); err == nil {
					_, _ = c.Write([]byte("N"))
				}
			}()
		}
	}()
	a := ln.Addr().(*net.TCPAddr)
	return a.IP.String(), a.Port
}

func TestDataSourceTest_Adhoc(t *testing.T) {
	r, _, _ := newDSEnv(t)
	host, port := fakePGRefusingSSL(t)

	w := doJSON(r, "POST", "/data-sources/test", map[string]interface{}{"host": host, "port": port, "sslMode": "require"})
	m := decode(t, w)
	if w.Code != 200 || m["success"] != false || !strings.Contains(m["message"].(string), "does not accept SSL") {
		t.Fatalf("require vs plaintext server: %d %s", w.Code, w.Body)
	}
	w = doJSON(r, "POST", "/data-sources/test", map[string]interface{}{"host": host, "port": port, "sslMode": "nope"})
	if w.Code != 400 || !strings.Contains(w.Body.String(), "unknown ssl_mode") {
		t.Fatalf("bad mode: %d %s", w.Code, w.Body)
	}
	w = doJSON(r, "POST", "/data-sources/test", map[string]interface{}{"host": host, "port": port, "sslMode": "require", "sslRootCert": testRootPEM})
	if w.Code != 400 {
		t.Fatalf("cert with require: %d %s", w.Code, w.Body)
	}
	w = doJSON(r, "POST", "/data-sources/test", map[string]interface{}{"port": port})
	if w.Code != 400 {
		t.Fatalf("no host: %d", w.Code)
	}
	// Legacy ssl=false: plain TCP reachability only.
	w = doJSON(r, "POST", "/data-sources/test", map[string]interface{}{"host": host, "port": port, "ssl": false})
	if m = decode(t, w); w.Code != 200 || m["success"] != true {
		t.Fatalf("plain reachability: %d %s", w.Code, w.Body)
	}
}

func TestDataSourceTestByID_MarksErrorOnTLSFailure(t *testing.T) {
	r, db, _ := newDSEnv(t)
	host, port := fakePGRefusingSSL(t)
	conn := models.DataSourceConnection{ID: uuid.New(), Name: "GL", Type: "postgres", Host: host, Port: port, Database: "gl",
		Status: "Connected", SSL: true, SSLMode: "require"}
	if err := db.Create(&conn).Error; err != nil {
		t.Fatal(err)
	}

	w := doJSON(r, "POST", "/data-sources/"+conn.ID.String()+"/test", nil)
	m := decode(t, w)
	if w.Code != 200 || m["success"] != false {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var row models.DataSourceConnection
	db.First(&row, "id = ?", conn.ID)
	if row.Status != "Error" || !strings.Contains(row.LastError, "does not accept SSL") {
		t.Fatalf("status=%q lastError=%q", row.Status, row.LastError)
	}

	// Recovery: ssl disabled -> Connected and LastError cleared.
	db.Model(&row).Updates(map[string]interface{}{"ssl": false, "ssl_mode": "disable"})
	w = doJSON(r, "POST", "/data-sources/"+conn.ID.String()+"/test", nil)
	db.First(&row, "id = ?", conn.ID)
	if w.Code != 200 || row.Status != "Connected" || row.LastError != "" {
		t.Fatalf("recovery: %d status=%q err=%q", w.Code, row.Status, row.LastError)
	}
	if w = doJSON(r, "POST", "/data-sources/"+uuid.NewString()+"/test", nil); w.Code != 404 {
		t.Fatalf("missing: %d", w.Code)
	}
}

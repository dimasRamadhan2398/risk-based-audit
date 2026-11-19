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
	"time"

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
			port INTEGER, database TEXT, environment TEXT, status TEXT, sync_schedule TEXT, last_sync TEXT,
			last_synced_at DATETIME, next_sync_at DATETIME, last_error TEXT,
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
	if p["data_mappings"] == nil || p["scopes"] == nil {
		t.Fatalf("hub payload must send empty list [] instead of null for scopes/data_mappings: %v", p)
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

func TestDataSourceSync_CooldownAndHubFailure(t *testing.T) {
	_, db, hub := newDSEnv(t)
	now := time.Unix(1700000000, 0)
	tracker := newSyncCooldownTracker(60 * time.Second)
	tracker.nowFunc = func() time.Time { return now }

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]interface{}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &m)
		hub.mu.Lock()
		hub.bodies = append(hub.bodies, m)
		st := hub.status
		hub.mu.Unlock()
		w.WriteHeader(st)
		resp := map[string]interface{}{
			"status":        "scheduled",
			"job_id":        "job-sync-1",
			"ssl_mode":      m["ssl_mode"],
			"sources_count": 1,
		}
		out, _ := json.Marshal(resp)
		_, _ = w.Write(out)
	}))
	defer srv.Close()

	dh := services.NewDataHubClient()
	dh.BaseURL = srv.URL

	r := gin.New()
	r.POST("/data-sources/:id/sync", dataSourceSync(db, dh, tracker))
	r.POST("/data-sources/sync-all", dataSourceSyncAll(db, dh, tracker, nil))

	conn1 := models.DataSourceConnection{ID: uuid.New(), Name: "CBS-1", Type: "postgres", Host: "127.0.0.1", Port: 5434, Database: "cbs", Status: "Connected", SSL: false, SSLMode: "disable"}
	conn2 := models.DataSourceConnection{ID: uuid.New(), Name: "CBS-2", Type: "postgres", Host: "127.0.0.1", Port: 5435, Database: "erp", Status: "Connected", SSL: false, SSLMode: "disable"}
	db.Create(&conn1)
	db.Create(&conn2)

	// 1st sync succeeds
	w := doJSON(r, "POST", "/data-sources/"+conn1.ID.String()+"/sync", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("first sync: %d %s", w.Code, w.Body.String())
	}
	m := decode(t, w)
	if m["job_id"] != "job-sync-1" || int(m["cooldown_seconds"].(float64)) != 60 {
		t.Fatalf("unexpected sync payload: %v", m)
	}
	// Must record 0 queued records instead of random rand.Intn
	var logEntry models.DataSourceActivityLog
	db.Order("created_at DESC").First(&logEntry, "connection_id = ?", conn1.ID)
	if logEntry.Records != 0 || logEntry.Status != "SUCCESS" {
		t.Fatalf("expected honest 0 queued records, got %d status=%q", logEntry.Records, logEntry.Status)
	}

	// Immediate repeat on conn1 returns 429 Too Many Requests with retry_after
	now = now.Add(15 * time.Second)
	w = doJSON(r, "POST", "/data-sources/"+conn1.ID.String()+"/sync", nil)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 during cooldown, got %d %s", w.Code, w.Body.String())
	}
	m = decode(t, w)
	if int(m["retry_after"].(float64)) != 45 {
		t.Fatalf("expected retry_after=45, got %v", m["retry_after"])
	}

	// Immediate sync on DIFFERENT connection still succeeds
	w = doJSON(r, "POST", "/data-sources/"+conn2.ID.String()+"/sync", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("different source should not be blocked: %d %s", w.Code, w.Body.String())
	}

	// After 60s elapses, conn1 can sync again, and if Hub returns 500, conn1 is marked Error
	now = now.Add(50 * time.Second)
	hub.mu.Lock()
	hub.status = http.StatusInternalServerError
	hub.mu.Unlock()
	w = doJSON(r, "POST", "/data-sources/"+conn1.ID.String()+"/sync", nil)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 when hub fails, got %d %s", w.Code, w.Body.String())
	}
	var afterErr models.DataSourceConnection
	db.First(&afterErr, "id = ?", conn1.ID)
	if afterErr.Status != "Error" || afterErr.LastError == "" {
		t.Fatalf("expected Error status on failed sync, got status=%q err=%q", afterErr.Status, afterErr.LastError)
	}

	// sync-all cooldown test
	hub.mu.Lock()
	hub.status = http.StatusOK
	hub.mu.Unlock()
	w = doJSON(r, "POST", "/data-sources/sync-all", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("sync-all first: %d %s", w.Code, w.Body.String())
	}
	w = doJSON(r, "POST", "/data-sources/sync-all", nil)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("sync-all immediate repeat should return 429, got %d", w.Code)
	}
}

func TestDataSourceScheduleAndMappingsAndReconcile(t *testing.T) {
	r, db, hub := newDSEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, _ := io.ReadAll(req.Body)
		var body map[string]interface{}
		_ = json.Unmarshal(b, &body)
		if strings.HasSuffix(req.URL.Path, "/reconcile") {
			w.WriteHeader(hub.status)
			if hub.status >= 400 {
				_, _ = w.Write([]byte(`{"detail":"hub offline"}`))
				return
			}
			_, _ = w.Write([]byte(`{"success":true,"total_deletes_detected":4}`))
			return
		}
		hub.mu.Lock()
		hub.bodies = append(hub.bodies, body)
		st := hub.status
		hub.mu.Unlock()
		w.WriteHeader(st)
		out, _ := json.Marshal(map[string]interface{}{"status": "success", "ssl_mode": body["ssl_mode"]})
		_, _ = w.Write(out)
	}))
	defer srv.Close()

	dh := services.NewDataHubClient()
	dh.BaseURL = srv.URL

	r = gin.New()
	r.POST("/data-sources", dataSourceCreate(db, dh))
	r.PUT("/data-sources/:id", dataSourceUpdate(db, dh))
	r.PUT("/data-sources/:id/mappings", dataSourceUpdateMappings(db, dh))
	r.POST("/data-sources/:id/reconcile", dataSourceReconcile(db, dh))

	// Legacy "Real-time (CDC)" on create is normalized to "Every 15 mins" with NextSyncAt populated.
	w := doJSON(r, "POST", "/data-sources", baseBody(map[string]interface{}{"syncSchedule": "Real-time (CDC)", "ssl": false}))
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	id := decode(t, w)["data"].(map[string]interface{})["id"].(string)
	var row models.DataSourceConnection
	db.First(&row, "id = ?", id)
	if row.SyncSchedule != models.SyncScheduleEvery15Mins || row.NextSyncAt == nil {
		t.Fatalf("expected Every 15 mins with NextSyncAt, got schedule=%q next=%v", row.SyncSchedule, row.NextSyncAt)
	}

	// Updating to Manual Only clears NextSyncAt.
	w = doJSON(r, "PUT", "/data-sources/"+id, baseBody(map[string]interface{}{"syncSchedule": "Manual Only", "ssl": false}))
	if w.Code != http.StatusOK {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	var afterManual models.DataSourceConnection
	db.First(&afterManual, "id = ?", id)
	if afterManual.SyncSchedule != models.SyncScheduleManualOnly || afterManual.NextSyncAt != nil {
		t.Fatalf("expected Manual Only with nil NextSyncAt, got schedule=%q next=%v", afterManual.SyncSchedule, afterManual.NextSyncAt)
	}

	// Saving mappings auto-resolves canonicalEntity and forwards it to Data Hub.
	w = doJSON(r, "PUT", "/data-sources/"+id+"/mappings", map[string]interface{}{
		"dataMappings": []map[string]interface{}{
			{
				"tableName":    "gl_transactions",
				"targetScope":  "data_analytics",
				"targetModule": "AI: Anomaly Detection",
				"auditField":   "id",
				"isActive":     true,
			},
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("mappings: %d %s", w.Code, w.Body.String())
	}
	db.First(&row, "id = ?", id)
	if len(row.DataMappings) != 1 || row.DataMappings[0].CanonicalEntity != "transactions" {
		t.Fatalf("expected canonicalEntity=transactions, got %+v", row.DataMappings)
	}

	// Reconcile success writes SUCCESS log and returns count.
	w = doJSON(r, "POST", "/data-sources/"+id+"/reconcile", nil)
	if w.Code != http.StatusOK || int(decode(t, w)["total_deletes_detected"].(float64)) != 4 {
		t.Fatalf("reconcile ok: %d %s", w.Code, w.Body.String())
	}

	// Reconcile failure returns 502 instead of fake rand count.
	hub.mu.Lock()
	hub.status = http.StatusBadGateway
	hub.mu.Unlock()
	w = doJSON(r, "POST", "/data-sources/"+id+"/reconcile", nil)
	if w.Code != http.StatusBadGateway || decode(t, w)["success"] != false {
		t.Fatalf("reconcile fail should return 502: %d %s", w.Code, w.Body.String())
	}
}

func TestDataSourceMappedRecords_RealProfileAndNoFakeFallback(t *testing.T) {
	_, db, _ := newDSEnv(t)
	hubOnline := true
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !hubOnline {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"detail":"offline"}`))
			return
		}
		if strings.Contains(req.URL.Path, "/preview/corporate_transactions") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"success": true,
				"data": {
					"tableName": "corporate_transactions",
					"total_rows": 3244,
					"rows": [
						{"trx_id": 2801, "trx_ref_number": "NSA-AFT8-00001", "branch_name": "KCU Sudirman Plaza", "amount": 850000000, "is_suspicious": true}
					],
					"profile": {
						"total_rows": 3244,
						"flagged_rows": 318,
						"has_numeric_measure": true,
						"dimension_column": "branch_name",
						"amount_column": "amount"
					}
				}
			}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dh := services.NewDataHubClient()
	dh.BaseURL = srv.URL

	conn := models.DataSourceConnection{
		ID:       uuid.New(),
		Name:     "CBS Local",
		Type:     "postgres",
		Host:     "127.0.0.1",
		Port:     5432,
		Database: "cbs",
		Status:   "Connected",
		DataMappings: []models.TableDataMapping{
			{
				TableName:    "corporate_transactions",
				TargetScope:  "audit_features",
				TargetModule: "Strategic Audit Plan",
				AuditField:   "trx_id",
				IsActive:     true,
			},
			{
				TableName:    "corporate_transactions",
				TargetScope:  "audit_features",
				TargetModule: "Annual Audit Plan",
				AuditField:   "trx_id",
				IsActive:     true,
			},
		},
	}
	if err := db.Create(&conn).Error; err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.GET("/data-sources/mapped-records", dataSourceMappedRecords(db, dh))

	// 1. Strategic Audit Plan returns real total_rows, profile, and sample_records
	req := httptest.NewRequest(http.MethodGet, "/data-sources/mapped-records?target_scope=audit_features&target_module=Strategic+Audit+Plan", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("mapped-records strategic: %d %s", w.Code, w.Body.String())
	}
	m := decode(t, w)
	dataList, _ := m["data"].([]interface{})
	if len(dataList) != 1 {
		t.Fatalf("expected 1 mapped record, got %d", len(dataList))
	}
	rec := dataList[0].(map[string]interface{})
	if int(rec["total_rows"].(float64)) != 3244 || rec["data_hub_unavailable"] == true {
		t.Fatalf("expected total_rows=3244 and hub online, got %+v", rec)
	}
	samples := rec["sample_records"].([]interface{})
	if len(samples) != 1 {
		t.Fatalf("expected 1 real sample row, got %d", len(samples))
	}

	// 2. When Data Hub is offline, returns empty sample_records and data_hub_unavailable=true (no fake REC-01)
	hubOnline = false
	req2 := httptest.NewRequest(http.MethodGet, "/data-sources/mapped-records?target_scope=audit_features&target_module=Annual+Audit+Plan", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	m2 := decode(t, w2)
	rec2 := m2["data"].([]interface{})[0].(map[string]interface{})
	samples2 := rec2["sample_records"].([]interface{})
	if len(samples2) != 0 || rec2["data_hub_unavailable"] != true || int(rec2["total_rows"].(float64)) != 0 {
		t.Fatalf("expected empty sample_records and data_hub_unavailable=true when offline, got %+v", rec2)
	}
}

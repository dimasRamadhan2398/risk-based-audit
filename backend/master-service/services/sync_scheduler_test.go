package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"master-service/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func openSchedulerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: gormlogger.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
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
	return db
}

type mockHubRecorder struct {
	mu          sync.Mutex
	ingestedIDs []string
	failIngest  bool
}

func (m *mockHubRecorder) server(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var body map[string]interface{}
		_ = json.Unmarshal(b, &body)

		switch {
		case strings.HasSuffix(r.URL.Path, "/sources/register"):
			w.WriteHeader(http.StatusOK)
			echo, _ := json.Marshal(map[string]interface{}{"status": "success", "ssl_mode": body["ssl_mode"]})
			_, _ = w.Write(echo)
		case strings.HasSuffix(r.URL.Path, "/sources/ingest"):
			m.mu.Lock()
			fail := m.failIngest
			if !fail {
				if rawIDs, ok := body["source_ids"].([]interface{}); ok {
					for _, id := range rawIDs {
						if s, ok := id.(string); ok {
							m.ingestedIDs = append(m.ingestedIDs, s)
						}
					}
				}
			}
			m.mu.Unlock()
			if fail {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"detail":"hub ingest error"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"scheduled","job_id":"job-sched-1"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func (m *mockHubRecorder) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.ingestedIDs)
}

func TestDataSourceSyncScheduler_IntervalsAndNormalization(t *testing.T) {
	db := openSchedulerTestDB(t)
	rec := &mockHubRecorder{}
	srv := rec.server(t)

	dh := NewDataHubClient()
	dh.BaseURL = srv.URL

	now := time.Date(2026, 11, 17, 10, 0, 0, 0, time.UTC)
	duePast := now.Add(-1 * time.Minute)
	future := now.Add(30 * time.Minute)

	// 1) Legacy "Real-time (CDC)" with due next_sync_at -> normalized to "Every 15 mins" and synced.
	connDue := models.DataSourceConnection{
		ID:           uuid.New(),
		Name:         "CBS Due",
		Type:         "postgres",
		Host:         "127.0.0.1",
		Port:         5432,
		Database:     "cbs",
		Status:       "Connected",
		SyncSchedule: "Real-time (CDC)",
		NextSyncAt:   &duePast,
		SSL:          false,
		SSLMode:      "disable",
	}
	// 2) Hourly not yet due -> skipped until 30 mins later.
	connFuture := models.DataSourceConnection{
		ID:           uuid.New(),
		Name:         "ERP Future",
		Type:         "postgres",
		Host:         "127.0.0.1",
		Port:         5432,
		Database:     "erp",
		Status:       "Connected",
		SyncSchedule: "Hourly",
		NextSyncAt:   &future,
		SSL:          false,
		SSLMode:      "disable",
	}
	// 3) Manual Only -> never synced automatically, next_sync_at cleared.
	connManual := models.DataSourceConnection{
		ID:           uuid.New(),
		Name:         "Manual DB",
		Type:         "postgres",
		Host:         "127.0.0.1",
		Port:         5432,
		Database:     "manual_db",
		Status:       "Connected",
		SyncSchedule: "Manual Only",
		NextSyncAt:   &duePast,
		SSL:          false,
		SSLMode:      "disable",
	}
	for _, c := range []models.DataSourceConnection{connDue, connFuture, connManual} {
		if err := db.Create(&c).Error; err != nil {
			t.Fatal(err)
		}
	}

	var hookCalls int
	sched := NewDataSourceSyncScheduler(
		db,
		dh,
		WithNowFunc(func() time.Time { return now }),
		WithJobPolling(0, 0),
		WithOnSyncSuccess(func(string) { hookCalls++ }),
	)

	// Pass 1: only connDue should trigger.
	if n := sched.RunDueSyncs(context.Background()); n != 1 {
		t.Fatalf("expected 1 triggered sync, got %d", n)
	}
	if rec.count() != 1 || hookCalls != 1 {
		t.Fatalf("expected 1 ingest & 1 hook call, got ingest=%d hook=%d", rec.count(), hookCalls)
	}

	var updatedDue models.DataSourceConnection
	db.First(&updatedDue, "id = ?", connDue.ID)
	if updatedDue.SyncSchedule != models.SyncScheduleEvery15Mins {
		t.Fatalf("expected normalized schedule %q, got %q", models.SyncScheduleEvery15Mins, updatedDue.SyncSchedule)
	}
	if updatedDue.LastSyncedAt == nil || !updatedDue.LastSyncedAt.Equal(now) {
		t.Fatalf("expected LastSyncedAt=%v, got %v", now, updatedDue.LastSyncedAt)
	}
	wantNext := now.Add(15 * time.Minute)
	if updatedDue.NextSyncAt == nil || !updatedDue.NextSyncAt.Equal(wantNext) {
		t.Fatalf("expected NextSyncAt=%v, got %v", wantNext, updatedDue.NextSyncAt)
	}

	var updatedManual models.DataSourceConnection
	db.First(&updatedManual, "id = ?", connManual.ID)
	if updatedManual.NextSyncAt != nil {
		t.Fatalf("expected Manual Only NextSyncAt to be cleared, got %v", updatedManual.NextSyncAt)
	}

	// Pass 2 immediately after: nothing due.
	if n := sched.RunDueSyncs(context.Background()); n != 0 {
		t.Fatalf("expected 0 triggered syncs before interval elapses, got %d", n)
	}

	// Advance clock by 15 minutes: connDue should trigger again!
	now = now.Add(15 * time.Minute)
	if n := sched.RunDueSyncs(context.Background()); n != 1 {
		t.Fatalf("expected 1 triggered sync after 15m, got %d", n)
	}

	// Advance clock another 20 minutes (t+35m): both connDue (next was t+30m) and connFuture (t+30m) should trigger!
	now = now.Add(20 * time.Minute)
	if n := sched.RunDueSyncs(context.Background()); n != 2 {
		t.Fatalf("expected connDue and connFuture to trigger at t+35m, got %d", n)
	}

	var logs []models.DataSourceActivityLog
	db.Find(&logs)
	if len(logs) != 4 {
		t.Fatalf("expected 4 scheduled activity logs, got %d", len(logs))
	}
	if !strings.HasPrefix(logs[0].Event, "Scheduled Sync (") || logs[0].Status != "SUCCESS" {
		t.Fatalf("unexpected log entry: %+v", logs[0])
	}
}

func TestDataSourceSyncScheduler_FailureAdvancesNextSync(t *testing.T) {
	db := openSchedulerTestDB(t)
	rec := &mockHubRecorder{failIngest: true}
	srv := rec.server(t)

	dh := NewDataHubClient()
	dh.BaseURL = srv.URL

	now := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	duePast := now.Add(-5 * time.Minute)
	conn := models.DataSourceConnection{
		ID:           uuid.New(),
		Name:         "Failing DB",
		Type:         "postgres",
		Host:         "127.0.0.1",
		Port:         5432,
		Database:     "fail_db",
		Status:       "Connected",
		SyncSchedule: "Hourly",
		NextSyncAt:   &duePast,
		SSL:          false,
		SSLMode:      "disable",
	}
	if err := db.Create(&conn).Error; err != nil {
		t.Fatal(err)
	}

	sched := NewDataSourceSyncScheduler(
		db,
		dh,
		WithNowFunc(func() time.Time { return now }),
		WithJobPolling(0, 0),
	)

	if n := sched.RunDueSyncs(context.Background()); n != 1 {
		t.Fatalf("expected 1 processed sync, got %d", n)
	}

	var after models.DataSourceConnection
	db.First(&after, "id = ?", conn.ID)
	if after.Status != "Error" || !strings.Contains(after.LastError, "Scheduled ingest failed") {
		t.Fatalf("expected Error status, got status=%q err=%q", after.Status, after.LastError)
	}
	wantNext := now.Add(1 * time.Hour)
	if after.NextSyncAt == nil || !after.NextSyncAt.Equal(wantNext) {
		t.Fatalf("expected NextSyncAt to advance to %v even on failure, got %v", wantNext, after.NextSyncAt)
	}

	var failLog models.DataSourceActivityLog
	db.First(&failLog, "connection_id = ?", conn.ID)
	if failLog.Status != "FAILED" {
		t.Fatalf("expected FAILED log, got %+v", failLog)
	}
}

func TestAutoMappingDefaultsAndCanonicalResolution(t *testing.T) {
	mappings := []models.TableDataMapping{
		{TableName: "gl_transactions", TargetScope: "data_analytics", TargetModule: "AI: Anomaly Detection", AuditField: "id", IsActive: true},
		{TableName: "loan_accounts", TargetScope: "risk_management", TargetModule: "Corporate Risk Profile", AuditField: "loan_id", IsActive: true},
		{TableName: "compliance_incident_records", TargetScope: "data_analytics", TargetModule: "AI: Anomaly Detection", AuditField: "incident_id", IsActive: true},
	}
	out := models.ApplyAutoMappingDefaults(mappings, "Real-time (CDC)")
	if out[0].CanonicalEntity != "transactions" {
		t.Fatalf("expected transactions, got %q", out[0].CanonicalEntity)
	}
	if out[1].CanonicalEntity != "loans" {
		t.Fatalf("expected loans, got %q", out[1].CanonicalEntity)
	}
	if out[2].CanonicalEntity != "ai_training_pool" {
		t.Fatalf("expected ai_training_pool, got %q", out[2].CanonicalEntity)
	}
	if out[0].SyncInterval != models.SyncScheduleEvery15Mins {
		t.Fatalf("expected normalized Every 15 mins, got %q", out[0].SyncInterval)
	}
}

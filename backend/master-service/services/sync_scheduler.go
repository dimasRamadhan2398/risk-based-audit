package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"master-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BuildDataHubPayload constructs the canonical registration payload sent to Data Hub
// for a given data source connection.
func BuildDataHubPayload(conn *models.DataSourceConnection) map[string]interface{} {
	scopes := conn.Scopes
	if scopes == nil {
		scopes = []string{}
	}
	schedule := models.NormalizeSyncSchedule(conn.SyncSchedule)
	var mappings []models.TableDataMapping
	if len(conn.DataMappings) > 0 {
		mappings = make([]models.TableDataMapping, len(conn.DataMappings))
		copy(mappings, conn.DataMappings)
		mappings = models.ApplyAutoMappingDefaults(mappings, schedule)
	} else {
		mappings = []models.TableDataMapping{}
	}
	return map[string]interface{}{
		"source_id":     conn.ID.String(),
		"name":          conn.Name,
		"source_type":   conn.Type,
		"host":          conn.Host,
		"port":          conn.Port,
		"database_name": conn.Database,
		"username":      conn.Username,
		"password":      conn.Password,
		"ssl_enabled":   conn.SSL,
		"ssl_mode":      conn.SSLMode,
		"ssl_root_cert": conn.SSLRootCert,
		"sync_schedule": schedule,
		"scopes":        scopes,
		"data_mappings": mappings,
	}
}

// RegisterDataSourceInHub pushes a connection to Data Hub synchronously. On failure it
// records the error on the connection (status "Error", lastError) so callers and the UI
// can see it, instead of the failure being swallowed.
func RegisterDataSourceInHub(db *gorm.DB, dh *DataHubClient, conn *models.DataSourceConnection) error {
	if err := dh.RegisterSource(BuildDataHubPayload(conn)); err != nil {
		fmt.Printf("[MasterService] Warning: failed to register source %s in Data Hub: %v\n", conn.ID, err)
		conn.Status = "Error"
		if IsSSLModeNotApplied(err) {
			conn.LastError = err.Error()
		} else {
			conn.LastError = "Data Hub registration failed: " + err.Error()
		}
		db.Save(conn)
		return err
	}
	return nil
}

// SyncSourceToHubSSLStrict re-pushes a connection to the Data Hub (mapping save, schema, test).
// Other Hub errors are only logged (legacy behaviour), but when the Hub did not apply the
// requested ssl_mode the source is marked Error with lastError and the error is returned,
// so it is never left Connected on a silently downgraded TLS mode.
func SyncSourceToHubSSLStrict(db *gorm.DB, dh *DataHubClient, conn *models.DataSourceConnection) error {
	err := dh.RegisterSource(BuildDataHubPayload(conn))
	if err == nil {
		return nil
	}
	if IsSSLModeNotApplied(err) {
		conn.Status = "Error"
		conn.LastError = err.Error()
		db.Save(conn)
		return err
	}
	fmt.Printf("[MasterService] Warning: Data Hub registration of source %s failed (non-fatal): %v\n", conn.ID, err)
	return nil
}

// EnsureDataSourceSyncColumns safely adds last_synced_at and next_sync_at columns when
// running against an existing database where migrate wasn't run separately, and
// normalizes legacy "Real-time (CDC)" schedules to "Every 15 mins".
func EnsureDataSourceSyncColumns(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if !db.Migrator().HasColumn(&models.DataSourceConnection{}, "last_synced_at") {
		if err := db.Migrator().AddColumn(&models.DataSourceConnection{}, "LastSyncedAt"); err != nil {
			return fmt.Errorf("add last_synced_at column: %w", err)
		}
	}
	if !db.Migrator().HasColumn(&models.DataSourceConnection{}, "next_sync_at") {
		if err := db.Migrator().AddColumn(&models.DataSourceConnection{}, "NextSyncAt"); err != nil {
			return fmt.Errorf("add next_sync_at column: %w", err)
		}
	}

	var conns []models.DataSourceConnection
	if err := db.Find(&conns).Error; err != nil {
		return nil
	}
	now := time.Now().UTC()
	for i := range conns {
		conn := &conns[i]
		normalized := models.NormalizeSyncSchedule(conn.SyncSchedule)
		changed := conn.SyncSchedule != normalized
		conn.SyncSchedule = normalized

		if _, scheduled := models.SyncScheduleInterval(normalized); scheduled {
			if conn.NextSyncAt == nil {
				base := now
				if conn.LastSyncedAt != nil {
					base = *conn.LastSyncedAt
				}
				conn.NextSyncAt = models.ComputeNextSyncAt(normalized, base)
				changed = true
			}
		} else if conn.NextSyncAt != nil {
			conn.NextSyncAt = nil
			changed = true
		}

		if changed {
			_ = db.Model(conn).Updates(map[string]interface{}{
				"sync_schedule": conn.SyncSchedule,
				"next_sync_at":  conn.NextSyncAt,
			}).Error
		}
	}
	return nil
}

// DataSourceSyncScheduler periodically checks data_source_connections whose
// SyncSchedule interval has elapsed and triggers Data Hub ingestion.
type DataSourceSyncScheduler struct {
	db            *gorm.DB
	dh            *DataHubClient
	tickInterval  time.Duration
	pollInterval  time.Duration
	pollAttempts  int
	nowFunc       func() time.Time
	onSyncSuccess func(connName string)

	mu       sync.Mutex
	inFlight map[string]bool
}

// SchedulerOption configures a DataSourceSyncScheduler.
type SchedulerOption func(*DataSourceSyncScheduler)

// WithTickInterval overrides the scheduler check frequency.
func WithTickInterval(d time.Duration) SchedulerOption {
	return func(s *DataSourceSyncScheduler) {
		if d > 0 {
			s.tickInterval = d
		}
	}
}

// WithNowFunc overrides the clock (useful in tests).
func WithNowFunc(fn func() time.Time) SchedulerOption {
	return func(s *DataSourceSyncScheduler) {
		if fn != nil {
			s.nowFunc = fn
		}
	}
}

// WithOnSyncSuccess registers a callback fired after a scheduled ingest job is queued.
func WithOnSyncSuccess(fn func(connName string)) SchedulerOption {
	return func(s *DataSourceSyncScheduler) {
		s.onSyncSuccess = fn
	}
}

// WithJobPolling configures post-trigger status polling (set attempts=0 in tests to skip async polling).
func WithJobPolling(interval time.Duration, attempts int) SchedulerOption {
	return func(s *DataSourceSyncScheduler) {
		s.pollInterval = interval
		s.pollAttempts = attempts
	}
}

// NewDataSourceSyncScheduler constructs a scheduler with sensible defaults (30s check interval).
func NewDataSourceSyncScheduler(db *gorm.DB, dh *DataHubClient, opts ...SchedulerOption) *DataSourceSyncScheduler {
	s := &DataSourceSyncScheduler{
		db:           db,
		dh:           dh,
		tickInterval: 30 * time.Second,
		pollInterval: 2 * time.Second,
		pollAttempts: 30,
		nowFunc:      time.Now,
		inFlight:     make(map[string]bool),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start runs the scheduler loop until ctx is cancelled.
func (s *DataSourceSyncScheduler) Start(ctx context.Context) {
	if s.db == nil || s.dh == nil {
		return
	}
	// Run an initial check on startup so overdue connections catch up promptly.
	s.RunDueSyncs(ctx)

	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.RunDueSyncs(ctx)
		}
	}
}

// RunDueSyncs inspects all active connections and triggers ingestion for any source
// whose scheduled next_sync_at is at or before now. Returns the count of triggered sources.
func (s *DataSourceSyncScheduler) RunDueSyncs(ctx context.Context) int {
	if s.db == nil || s.dh == nil {
		return 0
	}
	var conns []models.DataSourceConnection
	if err := s.db.WithContext(ctx).Find(&conns).Error; err != nil {
		fmt.Printf("[DataSourceScheduler] Failed to query connections: %v\n", err)
		return 0
	}

	now := s.nowFunc().UTC()
	triggered := 0

	for i := range conns {
		if ctx.Err() != nil {
			break
		}
		conn := conns[i]
		normalized := models.NormalizeSyncSchedule(conn.SyncSchedule)
		if conn.SyncSchedule != normalized {
			conn.SyncSchedule = normalized
			_ = s.db.Model(&conn).Update("sync_schedule", normalized).Error
		}

		_, scheduled := models.SyncScheduleInterval(normalized)
		if !scheduled {
			if conn.NextSyncAt != nil {
				_ = s.db.Model(&conn).Update("next_sync_at", nil).Error
			}
			continue
		}

		if conn.NextSyncAt == nil {
			base := now
			if conn.LastSyncedAt != nil {
				base = *conn.LastSyncedAt
			}
			next := models.ComputeNextSyncAt(normalized, base)
			if next != nil && next.After(now) {
				_ = s.db.Model(&conn).Update("next_sync_at", next).Error
				continue
			}
		} else if conn.NextSyncAt.After(now) {
			continue
		}

		if !s.tryAcquire(conn.ID.String()) {
			continue
		}
		s.syncConnection(ctx, &conn, now)
		s.release(conn.ID.String())
		triggered++
	}

	return triggered
}

func (s *DataSourceSyncScheduler) tryAcquire(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inFlight[id] {
		return false
	}
	s.inFlight[id] = true
	return true
}

func (s *DataSourceSyncScheduler) release(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.inFlight, id)
}

func (s *DataSourceSyncScheduler) syncConnection(ctx context.Context, conn *models.DataSourceConnection, now time.Time) {
	startedAt := time.Now()
	nextSync := models.ComputeNextSyncAt(conn.SyncSchedule, now)

	if hubErr := RegisterDataSourceInHub(s.db, s.dh, conn); hubErr != nil {
		_ = s.db.Model(conn).Update("next_sync_at", nextSync).Error
		failLog := models.DataSourceActivityLog{
			ID:             uuid.New(),
			ConnectionID:   conn.ID,
			ConnectionName: conn.Name,
			Type:           conn.Type,
			Event:          fmt.Sprintf("Scheduled Sync (%s) — Data Hub Registration Failed", conn.SyncSchedule),
			Status:         "FAILED",
			Records:        0,
			Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
			Timestamp:      now.Format("15:04:05"),
		}
		s.db.Create(&failLog)
		return
	}

	ingestRes, err := s.dh.TriggerIngest([]string{conn.ID.String()}, "selected")
	if err != nil {
		conn.Status = "Error"
		conn.LastError = "Scheduled ingest failed: " + err.Error()
		conn.NextSyncAt = nextSync
		_ = s.db.Save(conn).Error

		failLog := models.DataSourceActivityLog{
			ID:             uuid.New(),
			ConnectionID:   conn.ID,
			ConnectionName: conn.Name,
			Type:           conn.Type,
			Event:          fmt.Sprintf("Scheduled Sync (%s) — Ingest Failed", conn.SyncSchedule),
			Status:         "FAILED",
			Records:        0,
			Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
			Timestamp:      now.Format("15:04:05"),
		}
		s.db.Create(&failLog)
		return
	}

	var jobID string
	if ingestRes != nil {
		if jid, ok := ingestRes["job_id"].(string); ok {
			jobID = jid
		}
	}

	conn.Status = "Connected"
	conn.LastError = ""
	conn.LastSync = "Just now"
	conn.LastSyncedAt = &now
	conn.NextSyncAt = nextSync
	_ = s.db.Save(conn).Error

	logEntry := models.DataSourceActivityLog{
		ID:             uuid.New(),
		ConnectionID:   conn.ID,
		ConnectionName: conn.Name,
		Type:           conn.Type,
		Event:          fmt.Sprintf("Scheduled Sync (%s)", conn.SyncSchedule),
		Status:         "SUCCESS",
		Records:        0,
		Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
		Timestamp:      now.Format("15:04:05"),
	}
	s.db.Create(&logEntry)

	if s.onSyncSuccess != nil {
		s.onSyncSuccess(conn.Name)
	}

	if jobID != "" && s.pollAttempts > 0 && s.pollInterval > 0 {
		go s.pollScheduledJob(ctx, conn.ID, logEntry.ID, conn.SyncSchedule, jobID, startedAt)
	}
}

func (s *DataSourceSyncScheduler) pollScheduledJob(
	ctx context.Context,
	connID, logID uuid.UUID,
	schedule, jobID string,
	startedAt time.Time,
) {
	for attempt := 0; attempt < s.pollAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.pollInterval):
		}

		status, err := s.dh.GetIngestStatus(jobID)
		if err != nil || status == nil {
			continue
		}

		st, _ := status["status"].(string)
		switch st {
		case "completed":
			processed := 0
			if rp, ok := status["records_processed"].(float64); ok {
				processed = int(rp)
			}
			_ = s.db.Model(&models.DataSourceActivityLog{}).Where("id = ?", logID).Updates(map[string]interface{}{
				"records":  processed,
				"event":    fmt.Sprintf("Scheduled Sync (%s) — Completed", schedule),
				"duration": fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
			}).Error
			return
		case "failed":
			errMsg, _ := status["error_message"].(string)
			if errMsg == "" {
				errMsg = "Background ingestion failed in Data Hub worker"
			}
			_ = s.db.Model(&models.DataSourceActivityLog{}).Where("id = ?", logID).Updates(map[string]interface{}{
				"status":   "FAILED",
				"event":    fmt.Sprintf("Scheduled Sync (%s) — Failed", schedule),
				"duration": fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
			}).Error
			_ = s.db.Model(&models.DataSourceConnection{}).Where("id = ?", connID).Updates(map[string]interface{}{
				"status":     "Error",
				"last_error": errMsg,
			}).Error
			return
		}
	}
}

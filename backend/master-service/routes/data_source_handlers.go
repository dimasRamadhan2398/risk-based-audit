package routes

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"master-service/models"
	"master-service/pkg/datasource"
	"master-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Data source handlers, extracted from RegisterRoutes so they can be tested with httptest.

func dataSourceCreate(db *gorm.DB, dh *services.DataHubClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var conn models.DataSourceConnection
		if err := c.ShouldBindJSON(&conn); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}
		if conn.ID == uuid.Nil {
			conn.ID = uuid.New()
		}
		if conn.Status == "" {
			conn.Status = "Connected"
		}
		if conn.LastSync == "" {
			conn.LastSync = "Just configured"
		}
		conn.SyncSchedule = models.NormalizeSyncSchedule(conn.SyncSchedule)
		conn.NextSyncAt = models.ComputeNextSyncAt(conn.SyncSchedule, time.Now().UTC())
		if err := applyDataSourceSSL(&conn, conn.SSLRootCert); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}

		if err := db.Create(&conn).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create data source: " + err.Error()})
			return
		}

		// Register in Data Hub synchronously so a failure reaches the caller.
		hubErr := registerDataSourceInHub(db, dh, &conn)

		log := models.DataSourceActivityLog{
			ID:             uuid.New(),
			ConnectionID:   conn.ID,
			ConnectionName: conn.Name,
			Type:           conn.Type,
			Event:          "Connection Created",
			Status:         "SUCCESS",
			Records:        0,
			Duration:       "0.1s",
			Timestamp:      time.Now().Format("15:04:05"),
		}
		if hubErr != nil {
			log.Status = "FAILED"
			log.Event = "Connection Created (Data Hub registration failed)"
		}
		db.Create(&log)

		resp := gin.H{"success": true, "data": conn, "dataHubRegistered": hubErr == nil}
		if hubErr != nil {
			resp["dataHubError"] = hubErr.Error()
		}
		c.JSON(http.StatusCreated, resp)
	}
}

func dataSourceUpdate(db *gorm.DB, dh *services.DataHubClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
			return
		}
		var existing models.DataSourceConnection
		if err := db.First(&existing, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Data source connection not found"})
			return
		}

		var req models.DataSourceConnection
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}

		existing.Name = req.Name
		existing.Type = req.Type
		existing.Host = req.Host
		existing.Port = req.Port
		existing.Database = req.Database
		existing.Environment = req.Environment
		if req.Status != "" {
			existing.Status = req.Status
		}
		existing.SyncSchedule = models.NormalizeSyncSchedule(req.SyncSchedule)
		existing.NextSyncAt = models.ComputeNextSyncAt(existing.SyncSchedule, time.Now().UTC())
		existing.SSLMode = req.SSLMode
		existing.SSL = req.SSL
		if err := applyDataSourceSSL(&existing, req.SSLRootCert); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}
		existing.Username = req.Username
		if req.Password != "" {
			existing.Password = req.Password
		}
		if req.Scopes != nil {
			existing.Scopes = req.Scopes
		}
		if req.DataMappings != nil {
			existing.DataMappings = req.DataMappings
		}

		if err := db.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update data source: " + err.Error()})
			return
		}

		// Update registration in Data Hub synchronously so a failure reaches the caller.
		hubErr := registerDataSourceInHub(db, dh, &existing)
		resp := gin.H{"success": true, "data": existing, "dataHubRegistered": hubErr == nil}
		if hubErr != nil {
			resp["dataHubError"] = hubErr.Error()
		}
		c.JSON(http.StatusOK, resp)
	}
}

func dataSourceTestAdhoc() gin.HandlerFunc {
	return func(c *gin.Context) {
		type TestRequest struct {
			Host        string `json:"host"`
			Port        int    `json:"port"`
			SSL         bool   `json:"ssl"`
			SSLMode     string `json:"sslMode"`
			SSLRootCert string `json:"sslRootCert"`
		}
		var req TestRequest
		c.ShouldBindJSON(&req)

		if req.Host == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Please enter a valid Host / Server IP address."})
			return
		}

		if req.Port == 0 {
			req.Port = 5432
		}

		target := net.JoinHostPort(req.Host, strconv.Itoa(req.Port))

		// SSL requested: run the same SSLRequest + TLS handshake the data source would use.
		// No plaintext fallback: a server that refuses SSL fails the test.
		sslMode, err := datasource.NormalizeSSLMode(req.SSL, req.SSLMode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}
		if err := datasource.ValidateSSLConfig(sslMode, req.SSLRootCert); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}
		if sslMode != datasource.SSLModeDisable {
			if err := datasource.ProbePostgresTLS(c.Request.Context(), req.Host, req.Port, sslMode, req.SSLRootCert, 5*time.Second); err != nil {
				c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": fmt.Sprintf("TLS handshake to %s succeeded (ssl_mode=%s).", target, sslMode)})
			return
		}

		start := time.Now()
		err = datasource.DialPostgresTCP(c.Request.Context(), req.Host, req.Port, 2*time.Second)
		durationMs := time.Since(start).Milliseconds()

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": fmt.Sprintf("Connection test failed: could not reach %s (after %dms): %v", target, durationMs, err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Connection handshake to %s succeeded (latency %dms).", target, durationMs),
		})
	}
}

func dataSourceTestByID(db *gorm.DB, dh ...*services.DataHubClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
			return
		}
		var conn models.DataSourceConnection
		if err := db.First(&conn, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Data source connection not found"})
			return
		}

		if mode, _ := datasource.NormalizeSSLMode(conn.SSL, conn.SSLMode); mode != datasource.SSLModeDisable {
			if err := datasource.ProbePostgresTLS(c.Request.Context(), conn.Host, conn.Port, mode, conn.SSLRootCert, 5*time.Second); err != nil {
				conn.Status = "Error"
				conn.LastError = err.Error()
				db.Save(&conn)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error(), "data": conn})
				return
			}
		} else {
			if err := datasource.DialPostgresTCP(c.Request.Context(), conn.Host, conn.Port, 3*time.Second); err != nil {
				conn.Status = "Error"
				conn.LastError = err.Error()
				db.Save(&conn)
				c.JSON(http.StatusOK, gin.H{"success": false, "message": "Connection test failed: " + err.Error(), "data": conn})
				return
			}
		}

		conn.Status = "Connected"
		conn.LastError = ""
		conn.LastSync = "Just now"
		db.Save(&conn)

		if len(dh) > 0 && dh[0] != nil {
			_ = syncSourceToHubSSLStrict(db, dh[0], &conn)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Successfully connected to %s at %s:%d (Database: %s).", conn.Name, conn.Host, conn.Port, conn.Database),
			"data":    conn,
		})
	}
}

const dataSourceSyncCooldown = 60 * time.Second

type syncCooldownTracker struct {
	mu       sync.Mutex
	window   time.Duration
	lastRun  map[string]time.Time
	inFlight map[string]bool
	nowFunc  func() time.Time
}

func newSyncCooldownTracker(window time.Duration) *syncCooldownTracker {
	return &syncCooldownTracker{
		window:   window,
		lastRun:  make(map[string]time.Time),
		inFlight: make(map[string]bool),
		nowFunc:  time.Now,
	}
}

func (t *syncCooldownTracker) tryStart(key string) (int, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.inFlight[key] {
		rem := int(math.Ceil(t.window.Seconds()))
		if rem < 1 {
			rem = 1
		}
		return rem, false
	}
	if last, ok := t.lastRun[key]; ok {
		elapsed := t.nowFunc().Sub(last)
		if elapsed < t.window {
			rem := int(math.Ceil((t.window - elapsed).Seconds()))
			if rem < 1 {
				rem = 1
			}
			return rem, false
		}
	}
	t.inFlight[key] = true
	return 0, true
}

func (t *syncCooldownTracker) finish(key string, applyCooldown bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.inFlight, key)
	if applyCooldown {
		t.lastRun[key] = t.nowFunc()
	}
}

func dataSourceSync(db *gorm.DB, dh *services.DataHubClient, tracker *syncCooldownTracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
			return
		}
		var conn models.DataSourceConnection
		if err := db.First(&conn, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Data source connection not found"})
			return
		}

		if tracker != nil {
			if rem, ok := tracker.tryStart(id.String()); !ok {
				c.Header("Retry-After", strconv.Itoa(rem))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"success":          false,
					"message":          fmt.Sprintf("Sync sedang dalam jeda anti-spam. Silakan tunggu %d detik sebelum melakukan sinkronisasi lagi.", rem),
					"retry_after":      rem,
					"cooldown_seconds": int(tracker.window.Seconds()),
				})
				return
			}
		}

		// Ensure source is registered in Data Hub (any Hub error aborts sync honestly)
		if hubErr := registerDataSourceInHub(db, dh, &conn); hubErr != nil {
			if tracker != nil {
				tracker.finish(id.String(), false)
			}
			failLog := models.DataSourceActivityLog{
				ID:             uuid.New(),
				ConnectionID:   conn.ID,
				ConnectionName: conn.Name,
				Type:           conn.Type,
				Event:          "Manual Triggered Sync (Data Hub Unreachable / Registration Failed)",
				Status:         "FAILED",
				Records:        0,
				Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
				Timestamp:      time.Now().Format("15:04:05"),
			}
			db.Create(&failLog)
			c.JSON(http.StatusBadGateway, gin.H{
				"success":      false,
				"message":      "Data Hub registration failed: " + hubErr.Error(),
				"dataHubError": hubErr.Error(),
				"data":         conn,
			})
			return
		}

		// Trigger real Data Hub ingest
		ingestRes, err := dh.TriggerIngest([]string{id.String()}, "selected")
		if err != nil {
			if tracker != nil {
				tracker.finish(id.String(), false)
			}
			conn.Status = "Error"
			conn.LastError = "Data Hub ingest failed: " + err.Error()
			db.Save(&conn)

			failLog := models.DataSourceActivityLog{
				ID:             uuid.New(),
				ConnectionID:   conn.ID,
				ConnectionName: conn.Name,
				Type:           conn.Type,
				Event:          "Manual Triggered Sync (Ingest Failed)",
				Status:         "FAILED",
				Records:        0,
				Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
				Timestamp:      time.Now().Format("15:04:05"),
			}
			db.Create(&failLog)

			c.JSON(http.StatusBadGateway, gin.H{
				"success":      false,
				"message":      "Data Hub ingest failed: " + err.Error(),
				"dataHubError": err.Error(),
				"data":         conn,
			})
			return
		}

		if tracker != nil {
			tracker.finish(id.String(), true)
		}

		var jobID string
		if ingestRes != nil {
			if jid, ok := ingestRes["job_id"].(string); ok {
				jobID = jid
			}
		}

		now := time.Now().UTC()
		conn.SyncSchedule = models.NormalizeSyncSchedule(conn.SyncSchedule)
		conn.Status = "Connected"
		conn.LastSync = "Just now"
		conn.LastSyncedAt = &now
		conn.NextSyncAt = models.ComputeNextSyncAt(conn.SyncSchedule, now)
		conn.LastError = ""
		db.Save(&conn)

		logEntry := models.DataSourceActivityLog{
			ID:             uuid.New(),
			ConnectionID:   conn.ID,
			ConnectionName: conn.Name,
			Type:           conn.Type,
			Event:          "Manual Triggered Sync (Data Hub Job Queued)",
			Status:         "SUCCESS",
			Records:        0,
			Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
			Timestamp:      time.Now().Format("15:04:05"),
		}
		db.Create(&logEntry)

		cooldownSec := int(dataSourceSyncCooldown.Seconds())
		if tracker != nil {
			cooldownSec = int(tracker.window.Seconds())
		}
		c.JSON(http.StatusOK, gin.H{
			"success":          true,
			"message":          fmt.Sprintf("Real ingestion job queued in Data Hub for %s.", conn.Name),
			"job_id":           jobID,
			"log_id":           logEntry.ID,
			"cooldown_seconds": cooldownSec,
			"retry_after":      cooldownSec,
			"data":             conn,
		})
	}
}

func dataSourceSyncAll(db *gorm.DB, dh *services.DataHubClient, tracker *syncCooldownTracker, afterHook func()) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		const batchKey = "__all__"
		if tracker != nil {
			if rem, ok := tracker.tryStart(batchKey); !ok {
				c.Header("Retry-After", strconv.Itoa(rem))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"success":          false,
					"message":          fmt.Sprintf("Sync All sedang dalam jeda anti-spam. Silakan tunggu %d detik lagi.", rem),
					"retry_after":      rem,
					"cooldown_seconds": int(tracker.window.Seconds()),
				})
				return
			}
		}

		res, err := dh.TriggerBatchIngest()
		if err != nil {
			if tracker != nil {
				tracker.finish(batchKey, false)
			}
			failLog := models.DataSourceActivityLog{
				ID:             uuid.New(),
				ConnectionName: "All Data Sources",
				Type:           "postgres",
				Event:          "Sync All Triggered (Data Hub Failed)",
				Status:         "FAILED",
				Records:        0,
				Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
				Timestamp:      time.Now().Format("15:04:05"),
			}
			db.Create(&failLog)
			c.JSON(http.StatusBadGateway, gin.H{
				"success":      false,
				"message":      "Failed to trigger sync all: " + err.Error(),
				"dataHubError": err.Error(),
			})
			return
		}

		if tracker != nil {
			tracker.finish(batchKey, true)
		}
		if afterHook != nil {
			afterHook()
		}

		now := time.Now().UTC()
		var allConns []models.DataSourceConnection
		if err := db.Find(&allConns).Error; err == nil {
			for i := range allConns {
				sched := models.NormalizeSyncSchedule(allConns[i].SyncSchedule)
				nextSync := models.ComputeNextSyncAt(sched, now)
				_ = db.Model(&allConns[i]).Updates(map[string]interface{}{
					"sync_schedule":  sched,
					"last_sync":      "Just now",
					"last_synced_at": &now,
					"next_sync_at":   nextSync,
				}).Error
			}
		}

		sourcesCount := 0
		if sc, ok := res["sources_count"].(float64); ok {
			sourcesCount = int(sc)
		}

		logEntry := models.DataSourceActivityLog{
			ID:             uuid.New(),
			ConnectionName: fmt.Sprintf("All Data Sources (%d)", sourcesCount),
			Type:           "postgres",
			Event:          "Sync All Triggered",
			Status:         "SUCCESS",
			Records:        sourcesCount,
			Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
			Timestamp:      time.Now().Format("15:04:05"),
		}
		db.Create(&logEntry)

		cooldownSec := int(dataSourceSyncCooldown.Seconds())
		if tracker != nil {
			cooldownSec = int(tracker.window.Seconds())
		}
		c.JSON(http.StatusOK, gin.H{
			"success":          true,
			"message":          "Full synchronization initiated for all registered data sources.",
			"job_id":           res["job_id"],
			"log_id":           logEntry.ID,
			"cooldown_seconds": cooldownSec,
			"retry_after":      cooldownSec,
			"data":             res,
		})
	}
}

func dataSourceSyncStatus(db *gorm.DB, dh *services.DataHubClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("jobId")
		status, err := dh.GetIngestStatus(jobID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": err.Error()})
			return
		}

		if st, _ := status["status"].(string); st == "completed" {
			var processed int
			if rp, ok := status["records_processed"].(float64); ok {
				processed = int(rp)
			}
			if logIDStr := c.Query("log_id"); logIDStr != "" {
				if logUUID, parseErr := uuid.Parse(logIDStr); parseErr == nil {
					var logRow models.DataSourceActivityLog
					if db.First(&logRow, "id = ?", logUUID).Error == nil {
						db.Model(&models.DataSourceActivityLog{}).Where("id = ?", logUUID).Updates(map[string]interface{}{
							"records": processed,
							"event":   "Manual Triggered Sync (Completed)",
						})
						if logRow.ConnectionID != uuid.Nil {
							now := time.Now().UTC()
							db.Model(&models.DataSourceConnection{}).Where("id = ?", logRow.ConnectionID).Updates(map[string]interface{}{
								"last_sync":      "Just now",
								"last_synced_at": &now,
							})
						}
					}
				}
			}
		} else if st == "failed" {
			errMsg, _ := status["error_message"].(string)
			if errMsg == "" {
				errMsg = "Background ingestion failed in Data Hub worker"
			}
			if logIDStr := c.Query("log_id"); logIDStr != "" {
				if logUUID, parseErr := uuid.Parse(logIDStr); parseErr == nil {
					var logRow models.DataSourceActivityLog
					if db.First(&logRow, "id = ?", logUUID).Error == nil {
						db.Model(&models.DataSourceActivityLog{}).Where("id = ?", logUUID).Updates(map[string]interface{}{
							"status": "FAILED",
							"event":  "Manual Triggered Sync (Failed)",
						})
						if logRow.ConnectionID != uuid.Nil {
							db.Model(&models.DataSourceConnection{}).Where("id = ?", logRow.ConnectionID).Updates(map[string]interface{}{
								"status":     "Error",
								"last_error": errMsg,
							})
						}
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": status})
	}
}

func dataSourceUpdateMappings(db *gorm.DB, dh *services.DataHubClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
			return
		}
		var conn models.DataSourceConnection
		if err := db.First(&conn, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Data source connection not found"})
			return
		}

		type MappingPayload struct {
			DataMappings []models.TableDataMapping `json:"dataMappings"`
		}
		var req MappingPayload
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}

		conn.DataMappings = models.ApplyAutoMappingDefaults(req.DataMappings, conn.SyncSchedule)
		if err := db.Save(&conn).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update mappings: " + err.Error()})
			return
		}

		hubErr := syncSourceToHubSSLStrict(db, dh, &conn)

		logEntry := models.DataSourceActivityLog{
			ID:             uuid.New(),
			ConnectionID:   conn.ID,
			ConnectionName: conn.Name,
			Type:           conn.Type,
			Event:          "Schema Mappings Updated",
			Status:         "SUCCESS",
			Records:        len(req.DataMappings),
			Duration:       "0.2s",
			Timestamp:      time.Now().Format("15:04:05"),
		}
		db.Create(&logEntry)

		c.JSON(http.StatusOK, gin.H{
			"success":           true,
			"message":           fmt.Sprintf("Successfully saved %d table data mappings for %s.", len(req.DataMappings), conn.Name),
			"data":              conn,
			"dataHubRegistered": hubErr == nil,
			"dataHubError":      errString(hubErr),
		})
	}
}

func dataSourceReconcile(db *gorm.DB, dh *services.DataHubClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
			return
		}
		var conn models.DataSourceConnection
		if err := db.First(&conn, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Data source connection not found"})
			return
		}

		res, err := dh.ReconcileSource(id.String())
		if err != nil || res == nil {
			errMsg := "Data Hub returned empty response"
			if err != nil {
				errMsg = err.Error()
			}
			failLog := models.DataSourceActivityLog{
				ID:             uuid.New(),
				ConnectionID:   conn.ID,
				ConnectionName: conn.Name,
				Type:           conn.Type,
				Event:          "Deletion Reconcile Failed (Data Hub Unreachable)",
				Status:         "FAILED",
				Records:        0,
				Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
				Timestamp:      time.Now().Format("15:04:05"),
			}
			db.Create(&failLog)
			c.JSON(http.StatusBadGateway, gin.H{
				"success":      false,
				"message":      "Deletion reconciliation failed: Data Hub unreachable (" + errMsg + ")",
				"dataHubError": errMsg,
			})
			return
		}

		totalDeletes := 0
		if td, ok := res["total_deletes_detected"].(float64); ok {
			totalDeletes = int(td)
		}
		logEntry := models.DataSourceActivityLog{
			ID:             uuid.New(),
			ConnectionID:   conn.ID,
			ConnectionName: conn.Name,
			Type:           conn.Type,
			Event:          "Deletion Reconcile Completed",
			Status:         "SUCCESS",
			Records:        totalDeletes,
			Duration:       fmt.Sprintf("%.1fs", time.Since(startedAt).Seconds()),
			Timestamp:      time.Now().Format("15:04:05"),
		}
		db.Create(&logEntry)

		c.JSON(http.StatusOK, res)
	}
}

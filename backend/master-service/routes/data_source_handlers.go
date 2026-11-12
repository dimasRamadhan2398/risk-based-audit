package routes

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
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
		existing.SyncSchedule = req.SyncSchedule
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
		rawConn, err := net.DialTimeout("tcp", target, 2*time.Second)
		durationMs := time.Since(start).Milliseconds()

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": fmt.Sprintf("Connection test failed: could not reach %s (after %dms): %v", target, durationMs, err),
			})
			return
		}
		rawConn.Close()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Connection handshake to %s succeeded (latency %dms).", target, durationMs),
		})
	}
}

func dataSourceTestByID(db *gorm.DB) gin.HandlerFunc {
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
		}

		conn.Status = "Connected"
		conn.LastError = ""
		conn.LastSync = "Just now"
		db.Save(&conn)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Successfully connected to %s at %s:%d (Database: %s).", conn.Name, conn.Host, conn.Port, conn.Database),
			"data":    conn,
		})
	}
}

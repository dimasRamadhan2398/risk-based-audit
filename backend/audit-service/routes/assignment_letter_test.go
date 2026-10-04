package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"audit-service/models"
	"audit-service/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestAssignmentLetterCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Auto migrate the AssignmentLetter model
	if err := db.AutoMigrate(&models.AssignmentLetter{}); err != nil {
		t.Fatalf("failed to migrate AssignmentLetter: %v", err)
	}

	// Setup Gin engine
	engine := gin.New()

	// Create mock authentication middleware that allows all requests
	auth := &middleware.AuthMiddleware{}
	mockAuth := func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("roles", []string{"AUDITOR"})
		c.Next()
	}
	engine.Use(mockAuth)

	// Create route handler and register routes
	routeHandler := NewRouteHandler(engine, auth, db, nil)
	routeHandler.RegisterRoutes()

	// Test case: Create assignment letter with Draft status (should succeed)
	letterData := map[string]interface{}{
		"letterNumber":   "",
		"auditTitle":     "Test Audit",
		"leader":         "Test Leader",
		"category":       "Financial",
		"auditYear":      "2024",
		"auditTeam":      "TEST",
		"startPeriod":    "2024-01-01",
		"finishPeriod":   "2024-12-31",
		"workingUnit":    "Test Unit",
		"executionPeriod": "2024-01",
		"auditPurpose":   "Test Purpose",
		"status":         "Draft",
		"membersList": []map[string]string{
			{"name": "John Doe", "role": "Lead Auditor"},
		},
	}

	jsonData, err := json.Marshal(letterData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	// Make POST request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/assignment-letters", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	t.Logf("Response Status: %d", w.Code)
	t.Logf("Response Body: %s", w.Body.String())

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("response success is not true: %v", response)
	}
}


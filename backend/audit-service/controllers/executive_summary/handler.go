package executivesummary

import (
	"strconv"
	"strings"
	"time"

	"audit-service/pkg/logger"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CompilationStatsHandler handles GET /api/v1/executive-summary/compilation-stats?year=2026.
// A missing or invalid year falls back to the current year, as the performance
// endpoints do.
func CompilationStatsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		year, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("year", strconv.Itoa(now.Year()))))
		if err != nil {
			year = now.Year()
		}

		stats, err := LoadCompilationStats(db.WithContext(c.Request.Context()), year, now)
		if err != nil {
			if logger.Log != nil {
				logger.Error("failed to load executive summary compilation stats", zap.Error(err))
			}
			response.InternalServerError(c, "Failed to fetch compilation statistics")
			return
		}
		response.OK(c, "Compilation statistics retrieved successfully", stats)
	}
}

package executivesummary

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"audit-service/pkg/logger"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CompilationStatsHandler handles GET /api/v1/executive-summaries/compilation-stats?year=2026&quarter=1
// (also registered at the legacy /api/v1/executive-summary/compilation-stats, which Kong does not route).
// A missing or invalid year falls back to the current year, as the performance
// endpoints do. quarter is optional (1-4); without it the whole year is counted.
// A quarter outside 1-4 is a 400.
func CompilationStatsHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now()
		year, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("year", strconv.Itoa(now.Year()))))
		if err != nil {
			year = now.Year()
		}

		quarter := 0
		if raw := strings.TrimSpace(c.Query("quarter")); raw != "" {
			quarter, err = strconv.Atoi(raw)
			if err != nil || quarter < 1 || quarter > 4 {
				response.BadRequest(c, ErrInvalidQuarter.Error())
				return
			}
		}

		stats, err := LoadCompilationStats(db.WithContext(c.Request.Context()), year, quarter, now)
		if errors.Is(err, ErrInvalidQuarter) {
			response.BadRequest(c, ErrInvalidQuarter.Error())
			return
		}
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

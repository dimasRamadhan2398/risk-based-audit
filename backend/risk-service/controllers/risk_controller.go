package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"risk-service/pkg/masterclient"
	"risk-service/repositories"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RiskController struct {
	service services.IRiskService
}

func NewRiskController(service services.IRiskService) *RiskController {
	return &RiskController{service: service}
}

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

func badQuery(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INVALID_QUERY",
			"message": msg,
		},
	})
}

// parseRiskFilter reads the list query string. Pagination is on when page or
// page_size is present; otherwise the list is unpaged.
func parseRiskFilter(c *gin.Context) (repositories.RiskFilter, string) {
	var f repositories.RiskFilter

	pageStr, hasPage := c.GetQuery("page")
	sizeStr, hasSize := c.GetQuery("page_size")
	if hasPage || hasSize {
		f.Paged, f.Page, f.PageSize = true, 1, defaultPageSize
		if hasPage && pageStr != "" {
			n, err := strconv.Atoi(pageStr)
			if err != nil || n < 1 {
				return f, "page must be a positive integer"
			}
			f.Page = n
		}
		if hasSize && sizeStr != "" {
			n, err := strconv.Atoi(sizeStr)
			if err != nil {
				return f, "page_size must be an integer"
			}
			f.PageSize = min(max(n, 1), maxPageSize)
		}
	}

	f.Category = strings.TrimSpace(c.Query("category"))
	f.Search = strings.TrimSpace(c.Query("search"))

	if v := strings.TrimSpace(c.Query("fiscal_year")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1900 || n > 3000 {
			return f, "fiscal_year must be a year"
		}
		f.FiscalYear = n
	}
	if v := strings.TrimSpace(c.Query("branch_id")); v != "" {
		if v == UnassignedBranch {
			f.Unassigned = true
		} else {
			id, err := uuid.Parse(v)
			if err != nil {
				return f, "branch_id must be a UUID or " + UnassignedBranch
			}
			f.LocationID = id.String()
		}
	}
	if v := strings.ToLower(strings.TrimSpace(c.Query("period"))); v != "" {
		if v != "q1" && v != "q2" && v != "q3" && v != "q4" {
			return f, "period must be one of q1, q2, q3, q4"
		}
		f.Period = v
	}
	if v := strings.ToLower(strings.TrimSpace(c.Query("sort_by"))); v != "" {
		switch v {
		case repositories.SortSeverity, repositories.SortScore, repositories.SortImpact,
			repositories.SortLikelihood, repositories.SortName, repositories.SortCategory:
			f.SortBy = v
		default:
			return f, "sort_by must be one of severity, score, impact, likelihood, name, category"
		}
	}
	if v := strings.ToLower(strings.TrimSpace(c.Query("sort_dir"))); v != "" {
		desc := v == "desc"
		if v != "asc" && !desc {
			return f, "sort_dir must be asc or desc"
		}
		f.SortDesc = &desc
	}
	return f, ""
}

// UnassignedBranch is the branch_id value that selects risks with no location.
const UnassignedBranch = "unassigned"

// hasListQuery reports whether any filter/sort/page parameter is present.
func hasListQuery(c *gin.Context) bool {
	for _, k := range []string{"page", "page_size", "category", "search", "fiscal_year", "branch_id", "period", "sort_by", "sort_dir"} {
		if _, ok := c.GetQuery(k); ok {
			return true
		}
	}
	return false
}

// ListRisks returns every risk when called without query parameters (the
// original contract). With page/page_size it returns one page plus a
// "pagination" object; filters apply either way.
func (ctrl *RiskController) ListRisks(c *gin.Context) {
	if !hasListQuery(c) {
		data, err := ctrl.service.GetAll(requestContext(c))
		if err != nil {
			listFailed(c)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Risks fetched successfully",
			"data":    data,
		})
		return
	}

	f, msg := parseRiskFilter(c)
	if msg != "" {
		badQuery(c, msg)
		return
	}
	page, err := ctrl.service.List(requestContext(c), f)
	if err != nil {
		listFailed(c)
		return
	}
	body := gin.H{
		"success": true,
		"message": "Risks fetched successfully",
		"data":    page.Data,
	}
	if f.Paged {
		body["pagination"] = gin.H{
			"page":        page.Page,
			"page_size":   page.PageSize,
			"total":       page.Total,
			"total_pages": page.TotalPages,
		}
	}
	c.JSON(http.StatusOK, body)
}

// ListCategories returns the distinct risk categories, optionally as read in
// one fiscal_year.
func (ctrl *RiskController) ListCategories(c *gin.Context) {
	year := 0
	if v := strings.TrimSpace(c.Query("fiscal_year")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1900 || n > 3000 {
			badQuery(c, "fiscal_year must be a year")
			return
		}
		year = n
	}
	cats, err := ctrl.service.Categories(requestContext(c), year)
	if err != nil {
		listFailed(c)
		return
	}
	if cats == nil {
		cats = []string{}
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Risk categories fetched successfully",
		"data":    cats,
	})
}

func listFailed(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to query risks",
		},
	})
}

func (ctrl *RiskController) CreateRisk(c *gin.Context) {
	var req services.RiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	res, err := ctrl.service.Create(requestContext(c), &req)
	if err != nil {
		respondWriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Risk created successfully",
		"data":    res,
	})
}

func (ctrl *RiskController) UpdateRisk(c *gin.Context) {
	idStr := c.Param("id")
	regID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid risk ID format",
			},
		})
		return
	}

	var req services.RiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	res, err := ctrl.service.Update(requestContext(c), regID, &req)
	if err != nil {
		respondWriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Risk updated successfully",
		"data":    res,
	})
}

func (ctrl *RiskController) DeleteRisk(c *gin.Context) {
	idStr := c.Param("id")
	regID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid risk ID format",
			},
		})
		return
	}

	err = ctrl.service.Delete(regID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Risk deleted successfully",
	})
}

// requestContext forwards the caller's Authorization header to master-service
// lookups made while serving this request.
func requestContext(c *gin.Context) context.Context {
	return masterclient.WithAuthorization(c.Request.Context(), c.GetHeader("Authorization"))
}

// respondWriteError maps create/update failures: an unregistered location is
// the client's fault (400), an unreachable Location master is 503, anything
// else stays a DB error.
func respondWriteError(c *gin.Context, err error) {
	var locErr *services.LocationError
	switch {
	case errors.As(err, &locErr):
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    locErr.Code,
				"message": locErr.Message,
			},
		})
	case errors.Is(err, services.ErrLocationsUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "LOCATION_MASTER_UNAVAILABLE",
				"message": "Location master data is unavailable; try again shortly",
			},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": err.Error(),
			},
		})
	}
}

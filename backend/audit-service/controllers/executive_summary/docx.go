package executivesummary

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"audit-service/models"
	"audit-service/pkg/docxbuilder"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const docxMIME = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

// DownloadDocxHandler handles GET /api/v1/executive-summaries/:id/download-docx.
// A Word (.docx) document can only be generated when the executive summary is
// already Approved; Draft or Rejected summaries return 409 Conflict.
func DownloadDocxHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
		if err != nil {
			response.BadRequest(c, "Invalid ExecutiveSummary ID")
			return
		}

		ctx := c.Request.Context()
		var es models.ExecutiveSummary
		if err := db.WithContext(ctx).First(&es, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.NotFound(c, "ExecutiveSummary not found")
				return
			}
			response.InternalServerError(c, "Failed to fetch ExecutiveSummary")
			return
		}

		if !strings.EqualFold(strings.TrimSpace(es.Status), StatusApproved) {
			response.Error(c, http.StatusConflict, "EXECUTIVE_SUMMARY_NOT_APPROVED",
				"Word (.docx) can only be generated for an Approved executive summary", "")
			return
		}

		isCompilation := IsCompilationSummary(es.NomorDokumen)
		switch strings.ToLower(strings.TrimSpace(c.Query("variant"))) {
		case "compilation":
			isCompilation = true
		case "individual":
			isCompilation = false
		}

		opts := docxbuilder.ExecutiveSummaryDocxOptions{
			IsCompilation: isCompilation,
		}
		if !isCompilation && strings.TrimSpace(es.NomorDokumen) != "" {
			var lha models.AuditResultReport
			if err := db.WithContext(ctx).
				Where("TRIM(report_number) = ?", strings.TrimSpace(es.NomorDokumen)).
				Order("created_at DESC").
				First(&lha).Error; err == nil {
				opts.ReportTitle = lha.ReportTitle
				opts.AuditObject = lha.AuditObject
				opts.Department = lha.Department
				if es.AssignmentLetterID == "" && lha.AssignmentLetterID != "" {
					es.AssignmentLetterID = lha.AssignmentLetterID
				}
			}
		}

		docxBytes, err := docxbuilder.GenerateExecutiveSummaryDocx(&es, opts)
		if err != nil {
			response.InternalServerError(c, "Failed to generate Executive Summary Word document")
			return
		}

		cleanNum := strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(strings.TrimSpace(es.NomorDokumen))
		if cleanNum == "" {
			cleanNum = fmt.Sprintf("Q%d_%d", es.Quarter, es.Tahun)
		}
		prefix := "Executive_Summary_Individual_"
		if isCompilation {
			prefix = "Executive_Summary_Compilation_"
		}
		filename := fmt.Sprintf("%s%s.docx", prefix, cleanNum)
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		c.Data(http.StatusOK, docxMIME, docxBytes)
	}
}

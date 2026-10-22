package executivesummary

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/response"
	atrsvc "audit-service/services/action_taken_report"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MaxSignatureImageBytes is the largest decoded approval signature image (2 MB)
const MaxSignatureImageBytes = 2 << 20

// maxApproveBodyBytes bounds the approve request body: a base64 image of
// MaxSignatureImageBytes plus the other fields
const maxApproveBodyBytes = 4 << 20

const (
	maxSignaturePlaceLen = 100
	maxSignerNameLen     = 200
	maxSignerRoleLen     = 100
	dateLayout           = "2006-01-02"
)

// approverRoles are the canonical roles (services/action_taken_report RBAC) allowed
// to approve: ADMIN, and the executive, which the ATR matrix treats as the CAE
// (EXECUTIVE and CHIEF_AUDIT_EXECUTIVE both canonicalise to RoleCAE).
var approverRoles = map[string]bool{
	atrsvc.RoleAdmin: true,
	atrsvc.RoleCAE:   true,
}

// signatureMIMEPrefixes are the accepted data URL prefixes of the signature image
var signatureMIMEPrefixes = []struct {
	prefix string
	mime   string
}{
	{"data:image/png;base64,", "image/png"},
	{"data:image/jpeg;base64,", "image/jpeg"},
}

// ApproveRequest is the body of POST /executive-summaries/:id/approve
type ApproveRequest struct {
	SignaturePlace string `json:"signaturePlace"`
	SignatureDate  string `json:"signatureDate"`
	SignatureImage string `json:"signatureImage"`
	SignerName     string `json:"signerName"`
	SignerRole     string `json:"signerRole"`
}

// ApproverRole returns the first of the caller's roles that may approve, as
// written in the JWT, or "" when none may
func ApproverRole(roles []string) string {
	for _, r := range roles {
		if approverRoles[atrsvc.CanonicalRole(r)] {
			return strings.TrimSpace(r)
		}
	}
	return ""
}

// validateSignatureImage checks the data URL: a PNG or JPEG image within
// MaxSignatureImageBytes whose bytes really are that type.
func validateSignatureImage(raw string) (string, *apperrors.AppError) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", apperrors.New("SIGNATURE_REQUIRED", "signatureImage is required to approve an executive summary", http.StatusBadRequest)
	}
	lower := strings.ToLower(s)
	for _, p := range signatureMIMEPrefixes {
		if !strings.HasPrefix(lower, p.prefix) {
			continue
		}
		payload := s[len(p.prefix):]
		if len(payload) > base64.StdEncoding.EncodedLen(MaxSignatureImageBytes) {
			return "", apperrors.New("SIGNATURE_TOO_LARGE", "signature image must be at most 2 MB", http.StatusRequestEntityTooLarge)
		}
		decoded, err := base64.StdEncoding.DecodeString(payload)
		if err != nil || len(decoded) == 0 {
			return "", apperrors.New("INVALID_SIGNATURE", "signatureImage is not valid base64 image data", http.StatusBadRequest)
		}
		if len(decoded) > MaxSignatureImageBytes {
			return "", apperrors.New("SIGNATURE_TOO_LARGE", "signature image must be at most 2 MB", http.StatusRequestEntityTooLarge)
		}
		if http.DetectContentType(decoded) != p.mime {
			return "", apperrors.New("INVALID_SIGNATURE", "signatureImage must be a PNG or JPEG image", http.StatusBadRequest)
		}
		return s, nil
	}
	return "", apperrors.New("INVALID_SIGNATURE", "signatureImage must be a PNG or JPEG data URL", http.StatusBadRequest)
}

// parseSignatureDate returns the signature date as YYYY-MM-DD (the format the
// executive summary form stores). Empty means today (UTC).
func parseSignatureDate(raw string, now time.Time) (string, *apperrors.AppError) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return now.Format(dateLayout), nil
	}
	for _, layout := range []string{dateLayout, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Format(dateLayout), nil
		}
	}
	return "", apperrors.New("INVALID_SIGNATURE_DATE", "signatureDate must be a date (YYYY-MM-DD)", http.StatusBadRequest)
}

// ApproveHandler handles POST /executive-summaries/:id/approve. It moves a
// Draft summary to Approved and stores the signature and approval record.
//
//   - 403 unless the caller holds an approver role (ADMIN or EXECUTIVE)
//   - 404 when the summary does not exist
//   - 409 unless the summary is Draft
//   - 400 when signatureImage, signerName or another field is invalid
//
// The status change is a conditional UPDATE (status = Draft), so two
// concurrent approvals cannot both succeed.
func ApproveHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawRoles, _ := c.Get("roles")
		roles, _ := rawRoles.([]string)
		approverRole := ApproverRole(roles)
		if approverRole == "" {
			response.Forbidden(c, "Only ADMIN or EXECUTIVE can approve an executive summary")
			return
		}

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			response.BadRequest(c, "Invalid ExecutiveSummary ID")
			return
		}

		ctx := c.Request.Context()
		var existing models.ExecutiveSummary
		if err := db.WithContext(ctx).First(&existing, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.NotFound(c, "ExecutiveSummary not found")
				return
			}
			response.InternalServerError(c, "Failed to fetch ExecutiveSummary")
			return
		}
		if existing.Status != StatusDraft {
			response.Error(c, http.StatusConflict, "EXECUTIVE_SUMMARY_NOT_DRAFT",
				"Only a Draft executive summary can be approved (current status: "+existing.Status+")", "")
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxApproveBodyBytes)
		var req ApproveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		now := time.Now().UTC()
		image, appErr := validateSignatureImage(req.SignatureImage)
		if appErr != nil {
			writeAppError(c, appErr)
			return
		}
		signerName := strings.TrimSpace(req.SignerName)
		if signerName == "" {
			writeAppError(c, apperrors.New("SIGNER_NAME_REQUIRED", "signerName is required", http.StatusBadRequest))
			return
		}
		if utf8.RuneCountInString(signerName) > maxSignerNameLen {
			writeAppError(c, apperrors.New("INVALID_SIGNER_NAME", "signerName is too long", http.StatusBadRequest))
			return
		}
		signerRole := strings.TrimSpace(req.SignerRole)
		if utf8.RuneCountInString(signerRole) > maxSignerRoleLen {
			writeAppError(c, apperrors.New("INVALID_SIGNER_ROLE", "signerRole is too long", http.StatusBadRequest))
			return
		}
		place := strings.TrimSpace(req.SignaturePlace)
		if utf8.RuneCountInString(place) > maxSignaturePlaceLen {
			writeAppError(c, apperrors.New("INVALID_SIGNATURE_PLACE", "signaturePlace is too long", http.StatusBadRequest))
			return
		}
		date, appErr := parseSignatureDate(req.SignatureDate, now)
		if appErr != nil {
			writeAppError(c, appErr)
			return
		}

		approvedBy := callerName(c)
		res := db.WithContext(ctx).Model(&models.ExecutiveSummary{}).
			Where("id = ? AND status = ?", id, StatusDraft).
			Updates(map[string]interface{}{
				"status":            StatusApproved,
				"signature_tempat":  place,
				"signature_tanggal": date,
				"signature_image":   image,
				"signature_name":    signerName,
				"signature_role":    signerRole,
				"approved_by":       approvedBy,
				"approved_by_role":  approverRole,
				"approved_at":       now,
				"updated_at":        now,
			})
		if res.Error != nil {
			response.InternalServerError(c, "Failed to approve ExecutiveSummary")
			return
		}
		if res.RowsAffected == 0 {
			// Changed by another request between the status check and the update
			response.Error(c, http.StatusConflict, "EXECUTIVE_SUMMARY_NOT_DRAFT",
				"The executive summary is no longer a Draft; reload and try again", "")
			return
		}

		var approved models.ExecutiveSummary
		if err := db.WithContext(ctx).First(&approved, "id = ?", id).Error; err != nil {
			response.InternalServerError(c, "Failed to fetch ExecutiveSummary")
			return
		}
		response.OK(c, "Executive summary approved successfully", approved)
	}
}

// callerName is the authenticated user's username, falling back to the user id
func callerName(c *gin.Context) string {
	username, _ := c.Get("username")
	if s, ok := username.(string); ok && strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s)
	}
	userID, _ := c.Get("user_id")
	s, _ := userID.(string)
	return s
}

func writeAppError(c *gin.Context, e *apperrors.AppError) {
	response.Error(c, e.StatusCode, e.Code, e.Message, e.Details)
}

// ListDecorator is the crud.ItemsDecorator of GET /executive-summaries. The
// approval signature image can be large, so list rows leave it out; the detail
// response (GET /executive-summaries/:id) keeps it.
func ListDecorator(_ *gin.Context, items interface{}) (interface{}, error) {
	if list, ok := items.(*[]models.ExecutiveSummary); ok {
		for i := range *list {
			(*list)[i].SignatureImage = ""
		}
	}
	return items, nil
}

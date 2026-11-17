package ic_assessment

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apperrors "audit-service/pkg/errors"
	atrsvc "audit-service/services/action_taken_report"
)

// Roles (plan 7.2 with decision Q1), compared through atrsvc.CanonicalRole
// (any casing/spacing; the seeded EXECUTIVE is the CAE)
var (
	// AssessorRoles may create, propose, update, submit and delete own drafts
	AssessorRoles = []string{atrsvc.RoleAuditor, atrsvc.RoleAuditManager, atrsvc.RoleCAE, atrsvc.RoleAdmin}
	// ApproverRoles may return, approve and reopen, and lock/unlock periods
	ApproverRoles = []string{atrsvc.RoleCAE, atrsvc.RoleAdmin}
	// scopedRoles only see APPROVED assessments of their own department
	scopedRoles = []string{atrsvc.RoleDeptHead, atrsvc.RoleAuditee}
)

// Actor is the authenticated caller
type Actor struct {
	UserID       string
	Name         string
	Roles        []string
	EmployeeCode string
	IP           string
	UserAgent    string
}

// IsAssessor reports an assessor role
func (a Actor) IsAssessor() bool { return atrsvc.HasCanonicalRole(a.Roles, AssessorRoles...) }

// IsApprover reports CAE or ADMIN
func (a Actor) IsApprover() bool { return atrsvc.HasCanonicalRole(a.Roles, ApproverRoles...) }

// IsAdmin reports ADMIN
func (a Actor) IsAdmin() bool { return atrsvc.IsAdmin(a.Roles) }

// IsScopedReader: every role of the caller is DEPARTMENT_HEAD or AUDITEE, so
// they read only APPROVED assessments of their own department
func (a Actor) IsScopedReader() bool {
	if len(a.Roles) == 0 {
		return false
	}
	for _, r := range a.Roles {
		c := atrsvc.CanonicalRole(r)
		scoped := false
		for _, s := range scopedRoles {
			if c == s {
				scoped = true
			}
		}
		if !scoped {
			return false
		}
	}
	return true
}

// CanRead: any role at all (scoped readers are narrowed elsewhere)
func (a Actor) CanRead() bool { return len(a.Roles) > 0 }

// ApproverRoleName is the first approver role as written in the JWT
func (a Actor) ApproverRoleName() string {
	for _, r := range a.Roles {
		if atrsvc.HasCanonicalRole([]string{r}, ApproverRoles...) {
			return strings.TrimSpace(r)
		}
	}
	return ""
}

// RolesString is the roles as stored on events
func (a Actor) RolesString() string { return strings.Join(a.Roles, ",") }

// SameUser compares user ids case-insensitively; empty ids never match
func SameUser(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	return a != "" && b != "" && strings.EqualFold(a, b)
}

// Errors returned to the HTTP layer

func invalidf(format string, args ...interface{}) *apperrors.AppError {
	return apperrors.New("VALIDATION_ERROR", fmt.Sprintf(format, args...), http.StatusUnprocessableEntity)
}

func invalidList(message string, issues []string) *apperrors.AppError {
	e := apperrors.New("VALIDATION_ERROR", message, http.StatusUnprocessableEntity)
	if b, err := json.Marshal(issues); err == nil {
		e.Details = string(b)
	}
	return e
}

func conflictf(code, format string, args ...interface{}) *apperrors.AppError {
	return apperrors.New(code, fmt.Sprintf(format, args...), http.StatusConflict)
}

func forbiddenf(format string, args ...interface{}) *apperrors.AppError {
	return apperrors.New("FORBIDDEN", fmt.Sprintf(format, args...), http.StatusForbidden)
}

func notFound() *apperrors.AppError {
	return apperrors.New("NOT_FOUND", "Internal control assessment not found", http.StatusNotFound)
}

func unavailablef(format string, args ...interface{}) *apperrors.AppError {
	return apperrors.New("SOURCE_UNAVAILABLE", fmt.Sprintf(format, args...), http.StatusBadGateway)
}

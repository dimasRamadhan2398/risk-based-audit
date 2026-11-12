package employee

import (
	"net/http"
	"strings"

	"master-service/models"
	"master-service/pkg/authclient"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/logger"

	"go.uber.org/zap"
)

// The login account follows the employee: deactivating, reactivating or
// deleting an employee does the same to its linked auth-service account.
// Auth-service may refuse (409 SEAT_LIMIT_REACHED, 409 LAST_ADMIN); the
// employee is then left as it was and the 409 is passed through unchanged.

// linkedUser returns the auth-service account linked to emp, or nil when there
// is none (or account sync is off, NewEmployeeService). The link is
// users.employee_id = employee_code; an unlinked account with the employee's
// email counts too (the same fallback as Reset Password). An account linked to
// another employee is never touched.
func (s *EmployeeService) linkedUser(ctx *base.BaseService, emp *models.Employee) (*authclient.User, *apperrors.AppError) {
	if s.skipAccounts {
		return nil, nil
	}
	if s.accounts == nil {
		return nil, apperrors.New("LOGIN_ACCOUNT_NOT_CONFIGURED",
			"The employee was not changed: the login account service is not configured on this server (AUTH_SERVICE_URL is empty).",
			http.StatusServiceUnavailable)
	}
	goCtx := ctx.GetContext()
	user, err := s.accounts.FindUser(goCtx, emp.EmployeeCode, emp.Email, authclient.AuthorizationFrom(goCtx))
	if err != nil {
		if appErr := accountSyncError(err); appErr != nil {
			return nil, appErr
		}
		return nil, nil
	}
	if user == nil {
		return nil, nil
	}
	switch {
	case user.EmployeeID == emp.EmployeeCode:
		return user, nil
	case user.EmployeeID == "" && strings.EqualFold(user.Email, emp.Email):
		return user, nil
	default:
		logger.Warn("Login account with this email is linked to another employee; left unchanged",
			zap.String("employee_code", emp.EmployeeCode), zap.String("username", user.Username))
		return nil, nil
	}
}

// setUserActive calls the auth-service status endpoint. A 404 (the account
// disappeared since the lookup) counts as done.
func (s *EmployeeService) setUserActive(ctx *base.BaseService, user *authclient.User, active bool) *apperrors.AppError {
	goCtx := ctx.GetContext()
	if err := s.accounts.SetUserStatus(goCtx, user.ID, active, authclient.AuthorizationFrom(goCtx)); err != nil {
		return accountSyncError(err)
	}
	return nil
}

// deleteUser hard-deletes the account. A 404 counts as done.
func (s *EmployeeService) deleteUser(ctx *base.BaseService, user *authclient.User) *apperrors.AppError {
	goCtx := ctx.GetContext()
	if err := s.accounts.DeleteUser(goCtx, user.ID, true, authclient.AuthorizationFrom(goCtx)); err != nil {
		return accountSyncError(err)
	}
	return nil
}

// accountSyncError maps an auth-service failure on a status change or delete
// to what the admin sees, or nil for a 404 (nothing left to change).
// SEAT_LIMIT_REACHED and LAST_ADMIN keep their code, message and data.
func accountSyncError(err error) *apperrors.AppError {
	apiErr, ok := err.(*authclient.APIError)
	if !ok {
		return apperrors.New("LOGIN_ACCOUNT_UNAVAILABLE",
			"The employee was not changed: the login account service could not be reached. Try again.",
			http.StatusBadGateway)
	}

	switch {
	case apiErr.StatusCode == http.StatusNotFound:
		return nil
	case apiErr.StatusCode == http.StatusConflict &&
		(apiErr.Code == apperrors.CodeSeatLimitReached || apiErr.Code == apperrors.CodeLastAdmin):
		message := apiErr.Message
		if message == "" {
			message = defaultSyncConflictMessage(apiErr.Code)
		}
		out := apperrors.New(apiErr.Code, message, http.StatusConflict)
		if len(apiErr.Data) > 0 {
			out = out.WithData(apiErr.Data)
		}
		return out
	case apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden:
		return apperrors.New("LOGIN_ACCOUNT_FORBIDDEN",
			"The employee was not changed: your account is not allowed to manage login accounts. An ADMIN must do this.",
			http.StatusForbidden)
	case apiErr.StatusCode == http.StatusConflict:
		return apperrors.New("LOGIN_ACCOUNT_CONFLICT",
			"The employee was not changed: "+apiErr.Message, http.StatusConflict)
	case apiErr.StatusCode >= 400 && apiErr.StatusCode < 500:
		return apperrors.New("LOGIN_ACCOUNT_INVALID",
			"The employee was not changed: "+apiErr.Message, http.StatusBadRequest)
	default:
		return apperrors.New("LOGIN_ACCOUNT_UNAVAILABLE",
			"The employee was not changed: the login account service failed. Try again.",
			http.StatusBadGateway)
	}
}

func defaultSyncConflictMessage(code string) string {
	if code == apperrors.CodeLastAdmin {
		return "This employee's login is the last active admin account; keep at least one active admin."
	}
	return "All full user seats are in use."
}

// employeeInUse is the 409 for a hard delete of a referenced employee.
func employeeInUse(refs []models.EmployeeReference) *apperrors.AppError {
	labels := make([]string, 0, len(refs))
	for _, r := range refs {
		labels = append(labels, r.Label)
	}
	return apperrors.New(apperrors.CodeEmployeeInUse,
		"The employee cannot be deleted permanently because it is still used in: "+strings.Join(labels, ", ")+". Reassign those records first, or delete without hard=true.",
		http.StatusConflict).WithData(map[string]any{"references": refs})
}

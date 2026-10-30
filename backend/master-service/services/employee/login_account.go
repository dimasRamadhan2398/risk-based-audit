package employee

import (
	"context"
	"net/http"
	"strings"

	"master-service/models"
	"master-service/pkg/authclient"
	"master-service/pkg/base"
	"master-service/pkg/logger"
	apperrors "master-service/pkg/errors"

	"go.uber.org/zap"
)

// loginEmailSuffix marks the employees who get a login account on creation.
// Matched case-insensitively against the trimmed email.
const loginEmailSuffix = "@aifl.co.id"

// loginRole is the role given to employees' new login accounts.
const loginRole = "AUDITEE"

// maxLoginUsernameLength is auth-service's username limit. The username is the
// full email, so longer emails cannot have an account.
const maxLoginUsernameLength = 50

// AccountClient is the part of auth-service the employee service needs.
// *authclient.Client satisfies it.
type AccountClient interface {
	FindUserByEmail(ctx context.Context, email, authHeader string) (*authclient.User, error)
	CreateUser(ctx context.Context, in authclient.CreateUserInput, authHeader string) (*authclient.CreatedUser, error)
	LinkEmployee(ctx context.Context, userID, employeeCode, authHeader string) error
}

// requiresLoginAccount reports whether an employee with this email gets a login
// account when created.
func requiresLoginAccount(email string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(email)), loginEmailSuffix)
}

// checkLoginAccountPreflight runs before anything is written, so a create that
// can never get its account leaves no employee row behind.
func (s *EmployeeService) checkLoginAccountPreflight(email string) error {
	if s.accounts == nil {
		return apperrors.New("LOGIN_ACCOUNT_NOT_CONFIGURED",
			"Employees with an @aifl.co.id email cannot be created because login account provisioning is not configured on this server.",
			http.StatusServiceUnavailable)
	}
	if len(email) > maxLoginUsernameLength {
		return apperrors.New("LOGIN_EMAIL_TOO_LONG",
			"The email is too long for a login account (at most 50 characters).",
			http.StatusBadRequest).WithFields(map[string]string{"email": apperrors.FieldInvalid})
	}
	return nil
}

// provisionLoginAccount gives a freshly created employee its login account. An
// existing account with the same email is linked, not replaced. On any failure
// the employee row is removed again, so the create either fully succeeds or
// leaves nothing behind.
func (s *EmployeeService) provisionLoginAccount(ctx *base.BaseService, emp *models.Employee) (*models.EmployeeAccount, error) {
	goCtx := ctx.GetContext()
	authHeader := authclient.AuthorizationFrom(goCtx)

	existing, err := s.accounts.FindUserByEmail(goCtx, emp.Email, authHeader)
	if err != nil {
		return nil, s.rollbackEmployee(emp, accountCallError(err))
	}

	if existing != nil {
		if existing.EmployeeID != emp.EmployeeCode {
			if existing.EmployeeID != "" {
				return nil, s.rollbackEmployee(emp, apperrors.New("LOGIN_ACCOUNT_LINKED_ELSEWHERE",
					"Employee was not created: a login account with this email is already linked to another employee.",
					http.StatusConflict))
			}
			if err := s.accounts.LinkEmployee(goCtx, existing.ID, emp.EmployeeCode, authHeader); err != nil {
				return nil, s.rollbackEmployee(emp, accountCallError(err))
			}
		}
		logger.Info("Login account linked to new employee",
			zap.String("employee_code", emp.EmployeeCode), zap.String("username", existing.Username))
		return &models.EmployeeAccount{Status: "linked", Username: existing.Username}, nil
	}

	created, err := s.accounts.CreateUser(goCtx, authclient.CreateUserInput{
		Username:         emp.Email,
		Email:            emp.Email,
		FullName:         emp.FullName,
		Roles:            []string{loginRole},
		EmployeeID:       emp.EmployeeCode,
		GeneratePassword: true,
	}, authHeader)
	if err != nil {
		return nil, s.rollbackEmployee(emp, accountCallError(err))
	}

	// The temporary password goes back to the caller once. It is never logged.
	logger.Info("Login account created for new employee",
		zap.String("employee_code", emp.EmployeeCode), zap.String("username", created.Username))
	return &models.EmployeeAccount{
		Status:            "created",
		Username:          created.Username,
		TemporaryPassword: created.TemporaryPassword,
	}, nil
}

// rollbackEmployee removes the employee row and returns cause. If the removal
// itself fails, the error says the employee has to be removed by hand, because
// the caller would otherwise be told the create failed while the row exists.
func (s *EmployeeService) rollbackEmployee(emp *models.Employee, cause *apperrors.AppError) error {
	if err := s.employeeRepo.HardDelete(emp.ID); err != nil {
		logger.Error("Failed to roll back employee after login account failure",
			zap.String("employee_code", emp.EmployeeCode), zap.Error(err))
		return apperrors.New("LOGIN_ACCOUNT_ROLLBACK_FAILED",
			"The employee was saved but its login account could not be created, and the employee could not be removed. Delete employee "+emp.EmployeeCode+" manually before retrying.",
			http.StatusInternalServerError)
	}
	logger.Warn("Employee rolled back after login account failure",
		zap.String("employee_code", emp.EmployeeCode), zap.String("reason", cause.Code))
	return cause
}

// accountCallError maps an auth-service failure to the error the admin sees.
// Auth-service's own messages are user-facing, so they are passed through.
// Nothing from the request or response body is logged here.
func accountCallError(err error) *apperrors.AppError {
	apiErr, ok := err.(*authclient.APIError)
	if !ok {
		return apperrors.New("LOGIN_ACCOUNT_UNAVAILABLE",
			"Employee was not created: the login account service could not be reached. Try again.",
			http.StatusBadGateway)
	}

	switch {
	case apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden:
		return apperrors.New("LOGIN_ACCOUNT_FORBIDDEN",
			"Employee was not created: your account is not allowed to create login accounts. An ADMIN must create this employee.",
			http.StatusForbidden)
	case apiErr.StatusCode == http.StatusConflict:
		return apperrors.New("LOGIN_ACCOUNT_CONFLICT",
			"Employee was not created: "+apiErr.Message, http.StatusConflict)
	case apiErr.StatusCode >= 400 && apiErr.StatusCode < 500:
		return apperrors.New("LOGIN_ACCOUNT_INVALID",
			"Employee was not created: "+apiErr.Message, http.StatusBadRequest)
	default:
		return apperrors.New("LOGIN_ACCOUNT_UNAVAILABLE",
			"Employee was not created: the login account service failed. Try again.",
			http.StatusBadGateway)
	}
}

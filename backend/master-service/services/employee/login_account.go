package employee

import (
	"context"
	"net/http"
	"strings"

	"master-service/models"
	"master-service/pkg/authclient"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// JobRoleLookup reads a job role by id. The job role repository satisfies it.
type JobRoleLookup interface {
	FindByID(id uuid.UUID) (*models.JobRole, error)
}

// resolveLoginRole picks the role for a new account: the explicit override on
// the request, else the job role's default_access_role, else AUDITEE. A lookup
// failure falls back to AUDITEE (least privilege) and is logged.
func (s *EmployeeService) resolveLoginRole(emp *models.Employee) string {
	if override, _ := models.NormalizeAccessRole(emp.AccessRole); override != "" {
		return override
	}
	if s.jobRoles == nil {
		return models.DefaultLoginAccessRole
	}
	jr, err := s.jobRoles.FindByID(emp.JobRoleID)
	if err != nil {
		logger.Warn("Job role lookup failed; using default login role",
			zap.String("employee_code", emp.EmployeeCode), zap.Error(err))
		return models.DefaultLoginAccessRole
	}
	if role, ok := models.NormalizeAccessRole(jr.DefaultAccessRole); ok && role != "" {
		return role
	}
	return models.DefaultLoginAccessRole
}

// maxLoginUsernameLength is auth-service's username limit. The username is the
// full email, so longer emails cannot have an account.
const maxLoginUsernameLength = 50

// AccountClient is the part of auth-service the employee service needs.
// *authclient.Client satisfies it.
type AccountClient interface {
	FindUser(ctx context.Context, employeeCode, email, authHeader string) (*authclient.User, error)
	CreateUser(ctx context.Context, in authclient.CreateUserInput, authHeader string) (*authclient.CreatedUser, error)
	LinkEmployee(ctx context.Context, userID, employeeCode, authHeader string) error
}

// checkLoginAccountPreflight runs before anything is written, so a create that
// can never get its account leaves no employee row behind.
func (s *EmployeeService) checkLoginAccountPreflight(email, accessRole string) error {
	if _, ok := models.NormalizeAccessRole(accessRole); !ok {
		return apperrors.ValidationFailed("access_role must be one of: "+strings.Join(models.AllowedAccessRoles(), ", ")+".",
			map[string]string{"access_role": apperrors.FieldInvalid})
	}
	if s.accounts == nil {
		return apperrors.New("LOGIN_ACCOUNT_NOT_CONFIGURED",
			"Login account provisioning is not configured on this server (AUTH_SERVICE_URL is empty).",
			http.StatusServiceUnavailable)
	}
	if len(email) > maxLoginUsernameLength {
		return apperrors.New("LOGIN_EMAIL_TOO_LONG",
			"The email is too long for a login account (at most 50 characters).",
			http.StatusBadRequest).WithFields(map[string]string{"email": apperrors.FieldInvalid})
	}
	return nil
}

// ensureLoginAccount makes sure the employee has a login account and returns
// its status: "created" (new account, temporary password set once), "linked"
// (an account with the same email existed and is now linked) or "exists" (an
// account was already linked to this employee; nothing changed, no password).
// The lookup is by employee code then email, the same as Reset Password.
func (s *EmployeeService) ensureLoginAccount(ctx *base.BaseService, emp *models.Employee) (*models.EmployeeAccount, *apperrors.AppError) {
	goCtx := ctx.GetContext()
	authHeader := authclient.AuthorizationFrom(goCtx)

	existing, err := s.accounts.FindUser(goCtx, emp.EmployeeCode, emp.Email, authHeader)
	if err != nil {
		return nil, accountCallError(err)
	}

	if existing != nil {
		if existing.EmployeeID == emp.EmployeeCode {
			return &models.EmployeeAccount{Status: "exists", Username: existing.Username}, nil
		}
		if existing.EmployeeID != "" {
			return nil, apperrors.New("LOGIN_ACCOUNT_LINKED_ELSEWHERE",
				"A login account with this email is already linked to another employee.",
				http.StatusConflict)
		}
		if err := s.accounts.LinkEmployee(goCtx, existing.ID, emp.EmployeeCode, authHeader); err != nil {
			return nil, accountCallError(err)
		}
		logger.Info("Login account linked to employee",
			zap.String("employee_code", emp.EmployeeCode), zap.String("username", existing.Username))
		return &models.EmployeeAccount{Status: "linked", Username: existing.Username}, nil
	}

	created, err := s.accounts.CreateUser(goCtx, authclient.CreateUserInput{
		Username:         emp.Email,
		Email:            emp.Email,
		FullName:         emp.FullName,
		Roles:            []string{s.resolveLoginRole(emp)},
		EmployeeID:       emp.EmployeeCode,
		GeneratePassword: true,
	}, authHeader)
	if err != nil {
		return nil, accountCallError(err)
	}

	// The temporary password goes back to the caller once. It is never logged.
	logger.Info("Login account created for employee",
		zap.String("employee_code", emp.EmployeeCode), zap.String("username", created.Username))
	return &models.EmployeeAccount{
		Status:            "created",
		Username:          created.Username,
		TemporaryPassword: created.TemporaryPassword,
	}, nil
}

// provisionLoginAccount gives a freshly created employee its login account. On
// any failure the employee row is removed again, so the create either fully
// succeeds or leaves nothing behind.
func (s *EmployeeService) provisionLoginAccount(ctx *base.BaseService, emp *models.Employee) (*models.EmployeeAccount, error) {
	account, appErr := s.ensureLoginAccount(ctx, emp)
	if appErr != nil {
		return nil, s.rollbackEmployee(emp, appErr)
	}
	return account, nil
}

// ProvisionLoginAccount creates (or links) the login account of an existing
// employee on demand. Idempotent: an already linked account is returned with
// status "exists" and no password. accessRole optionally overrides the role,
// else the job role's default_access_role, else AUDITEE. Nothing is rolled back
// on failure; the employee row is untouched.
func (s *EmployeeService) ProvisionLoginAccount(ctx *base.BaseService, id, accessRole string) (*models.EmployeeAccount, error) {
	employeeID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.InvalidID(err)
	}
	emp, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, apperrors.ErrEmployeeNotFound
		}
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load the employee.")
	}
	if err := s.checkLoginAccountPreflight(emp.Email, accessRole); err != nil {
		return nil, err
	}
	// Work on a copy so the request-only override never touches the stored row.
	target := *emp
	target.AccessRole = accessRole
	account, appErr := s.ensureLoginAccount(ctx, &target)
	if appErr != nil {
		return nil, appErr
	}
	return account, nil
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

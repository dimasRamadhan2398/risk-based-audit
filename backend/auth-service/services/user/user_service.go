package services

import (
	"context"
	"encoding/base64"
	"net/http"
	"regexp"
	"strings"

	"auth-service/models"
	baseService "auth-service/pkg/base"
	"auth-service/pkg/errors"
	"auth-service/pkg/kafka"
	"auth-service/pkg/utils"
	"auth-service/repositories"

	"github.com/google/uuid"
)

// UserServiceInterface defines the user service interface
type UserServiceInterface interface {
	CreateUser(ctx context.Context, req *models.CreateUserRequest) error
	UpdateUser(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUser(ctx context.Context, id uuid.UUID) (*models.UserResponse, error)
	ListUsers(ctx context.Context, req *models.ListUsersRequest) ([]*models.UserResponse, *utils.PaginationResponse, error)
	FindUserByEmployee(ctx context.Context, employeeCode, email string) (*models.UserResponse, error)
	AdminResetPassword(ctx context.Context, actorID, targetID uuid.UUID, ipAddress string) (*models.AdminResetPasswordResponse, error)
}

// UserService handles user business logic
type UserService struct {
	*baseService.BaseService
	userRepo     repositories.UserRepositoryInterface
	kafkaProducer *kafka.Producer
}

// NewUserService creates a new user service
func NewUserService(userRepo repositories.UserRepositoryInterface, kafkaProducer *kafka.Producer) UserServiceInterface {
	return &UserService{
		BaseService:   baseService.NewBaseService(),
		userRepo:      userRepo,
		kafkaProducer: kafkaProducer,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest) error {
	// Check if username exists
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return errors.Wrap(errors.ErrDuplicateEntry.Code, "Username already exists", errors.ErrDuplicateEntry.StatusCode, nil)
	}

	// Check if email exists
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return errors.Wrap(errors.ErrDuplicateEntry.Code, "Email already exists", errors.ErrDuplicateEntry.StatusCode, nil)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.LogError("Failed to hash password", utils.LogField("error", err))
		return errors.ErrInternalServer
	}

	// Create user
	user := &models.User{
		Username:   req.Username,
		Email:      req.Email,
		PasswordHash:   hashedPassword,
		FullName:   req.FullName,
		Phone:      req.Phone,
		Department: req.Department,
		IsActive:   true,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.LogError("Failed to create user", utils.LogField("error", err))
		return errors.ErrInternalServer
	}

	s.LogInfo("User created successfully", utils.LogField("user_id", user.ID))
	return nil
}

// MaxAvatarDataURLLength caps the avatar data URL string (about 220 KB of image bytes).
const MaxAvatarDataURLLength = 300000

var avatarDataURLPattern = regexp.MustCompile(`^data:image/(png|jpeg|webp);base64,([A-Za-z0-9+/]+={0,2})$`)

// invalidAvatar builds the 400 error returned when avatar_url is not acceptable.
func invalidAvatar(msg string) *errors.AppError {
	return errors.New("INVALID_AVATAR", msg, http.StatusBadRequest)
}

// ValidateAvatarDataURL validates a non-empty avatar data URL.
func ValidateAvatarDataURL(v string) error {
	if len(v) > MaxAvatarDataURLLength {
		return invalidAvatar("avatar_url is too large (max 300000 characters)")
	}
	m := avatarDataURLPattern.FindStringSubmatch(v)
	if m == nil {
		return invalidAvatar("avatar_url must be a data URL like data:image/(png|jpeg|webp);base64,<data>")
	}
	if dec, err := base64.StdEncoding.DecodeString(m[2]); err != nil || len(dec) == 0 {
		return invalidAvatar("avatar_url contains invalid base64 data")
	}
	return nil
}

func avatarOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// UpdateUser updates a user. Only fields present in the request are changed.
func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.UserResponse, error) {
	if req.AvatarURL != nil && *req.AvatarURL != "" {
		if err := ValidateAvatarDataURL(*req.AvatarURL); err != nil {
			return nil, err
		}
	}

	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Department != nil {
		user.Department = *req.Department
	}
	if req.Position != nil {
		user.Position = *req.Position
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.AvatarURL != nil {
		if *req.AvatarURL == "" {
			user.AvatarURL = nil
		} else {
			v := *req.AvatarURL
			user.AvatarURL = &v
		}
	}

	if err := s.userRepo.Update(user); err != nil {
		s.LogError("Failed to update user", utils.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("User updated successfully", utils.LogField("user_id", user.ID))
	return s.userToResponse(user), nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepo.Delete(id); err != nil {
		s.LogError("Failed to delete user", utils.LogField("error", err))
		return errors.ErrInternalServer
	}

	s.LogInfo("User deleted successfully", utils.LogField("user_id", id))
	return nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.userToResponse(user), nil
}

// ListUsers retrieves a list of users
func (s *UserService) ListUsers(ctx context.Context, req *models.ListUsersRequest) ([]*models.UserResponse, *utils.PaginationResponse, error) {
	pagination := utils.NewPagination(req.Page, req.PageSize, "created_at", "desc")

	users, err := s.userRepo.FindMany(
		pagination.GetOffset(),
		pagination.GetLimit(),
		req.Search,
		req.Department,
		req.IsActive,
	)
	if err != nil {
		s.LogError("Failed to list users", utils.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	totalCount, err := s.userRepo.Count(req.Search, req.Department, req.IsActive)
	if err != nil {
		s.LogError("Failed to count users", utils.LogField("error", err))
		return nil, nil, errors.ErrInternalServer
	}

	responses := make([]*models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.userToResponse(user)
	}

	paginationResp := utils.BuildPaginationResponse(pagination.Page, pagination.PageSize, totalCount)

	return responses, paginationResp, nil
}

// FindUserByEmployee finds the login account linked to a master-data employee.
// Accounts are linked through users.employee_id = employees.employee_code; the
// email is used as a fallback for accounts created before the codes lined up.
func (s *UserService) FindUserByEmployee(ctx context.Context, employeeCode, email string) (*models.UserResponse, error) {
	if employeeCode != "" {
		user, err := s.userRepo.FindByEmployeeID(employeeCode)
		if err == nil {
			return s.userToResponse(user), nil
		}
		if !errors.Is(err, errors.ErrNotFound) {
			s.LogError("Failed to find user by employee code", utils.LogField("error", err))
			return nil, errors.ErrInternalServer
		}
	}

	if email != "" {
		user, err := s.userRepo.FindByEmail(email)
		if err == nil {
			return s.userToResponse(user), nil
		}
		if !errors.Is(err, errors.ErrNotFound) {
			s.LogError("Failed to find user by email", utils.LogField("error", err))
			return nil, errors.ErrInternalServer
		}
	}

	return nil, errors.Wrap("USER_ACCOUNT_NOT_FOUND", "This employee has no login account", 404, nil)
}

// AdminResetPassword replaces a user's password with a random temporary one and
// forces them to change it at next login. Admins cannot reset their own
// password here (they use change-password) or another admin's password (that
// goes through the platform super-admin), so one compromised admin account
// cannot take over the others.
func (s *UserService) AdminResetPassword(ctx context.Context, actorID, targetID uuid.UUID, ipAddress string) (*models.AdminResetPasswordResponse, error) {
	if actorID == targetID {
		return nil, errors.Wrap("CANNOT_RESET_OWN_PASSWORD", "Use change password to update your own password", 400, nil)
	}

	user, err := s.userRepo.FindByID(targetID)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, errors.Wrap(errors.ErrNotFound.Code, "User not found", errors.ErrNotFound.StatusCode, nil)
		}
		s.LogError("Failed to find user for password reset", utils.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	for _, role := range user.Roles {
		if strings.EqualFold(role.Name, "ADMIN") {
			return nil, errors.Wrap("CANNOT_RESET_ADMIN_PASSWORD", "Admin passwords can only be reset by the platform administrator", 403, nil)
		}
	}

	tempPassword, err := utils.GenerateTemporaryPassword()
	if err != nil {
		s.LogError("Failed to generate temporary password", utils.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	hashedPassword, err := utils.HashPassword(tempPassword)
	if err != nil {
		s.LogError("Failed to hash password", utils.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	user.PasswordHash = hashedPassword
	user.MustChangePassword = true
	user.LockedUntil = nil
	if err := s.userRepo.Update(user); err != nil {
		s.LogError("Failed to reset password", utils.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	s.LogInfo("Password reset by admin",
		utils.LogField("user_id", user.ID),
		utils.LogField("reset_by", actorID),
	)

	// Send to Kafka for centralized audit logging
	if s.kafkaProducer != nil && s.kafkaProducer.IsEnabled() {
		s.kafkaProducer.Info(ctx, "Password reset by admin", map[string]interface{}{
			"user_id":    user.ID.String(),
			"username":   user.Username,
			"reset_by":   actorID.String(),
			"ip_address": ipAddress,
		})
	}

	return &models.AdminResetPasswordResponse{
		UserID:            user.ID.String(),
		Username:          user.Username,
		TemporaryPassword: tempPassword,
	}, nil
}

// userToResponse converts a user model to a response DTO
func (s *UserService) userToResponse(user *models.User) *models.UserResponse {
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	return &models.UserResponse{
		ID:         user.ID.String(),
		EmployeeID: user.EmployeeID,
		Username:   user.Username,
		Email:      user.Email,
		FullName:   user.FullName,
		Phone:      user.Phone,
		Department: user.Department,
		Position:   user.Position,
		IsActive:   user.IsActive,
		AvatarURL:  avatarOrEmpty(user.AvatarURL),
		Roles:      roles,
		MustChangePassword: user.MustChangePassword,
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

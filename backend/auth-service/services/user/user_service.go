package services

import (
	"context"

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
	UpdateUser(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	GetUser(ctx context.Context, id uuid.UUID) (*models.UserResponse, error)
	ListUsers(ctx context.Context, req *models.ListUsersRequest) ([]*models.UserResponse, *utils.PaginationResponse, error)
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

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Update fields
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

	if err := s.userRepo.Update(user); err != nil {
		s.LogError("Failed to update user", utils.LogField("error", err))
		return errors.ErrInternalServer
	}

	s.LogInfo("User updated successfully", utils.LogField("user_id", user.ID))
	return nil
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

// userToResponse converts a user model to a response DTO
func (s *UserService) userToResponse(user *models.User) *models.UserResponse {
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	return &models.UserResponse{
		ID:         user.ID.String(),
		Username:   user.Username,
		Email:      user.Email,
		FullName:   user.FullName,
		Phone:      user.Phone,
		Department: user.Department,
		Position:   user.Position,
		IsActive:   user.IsActive,
		Roles:      roles,
		CreatedAt:  user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

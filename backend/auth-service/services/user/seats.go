package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"auth-service/models"
	"auth-service/pkg/activitylog"
	"auth-service/pkg/errors"
	"auth-service/pkg/utils"
	"auth-service/repositories"

	"github.com/google/uuid"
)

// MaxFullSeats is the fixed number of full seats of every tenant. A seat is an
// active, non-deleted user holding at least one role other than VIEWER (each
// user counts once, ADMIN and custom roles included). It is not configurable.
const MaxFullSeats = 50

// Role names the seat rules look at (compared case-insensitively).
const (
	ViewerRole = "VIEWER"
	AdminRole  = "ADMIN"
)

// Error codes of the seat and last-admin rules.
const (
	CodeSeatLimitReached = "SEAT_LIMIT_REACHED"
	CodeLastAdmin        = "LAST_ADMIN"
)

// IsSeatRole reports whether holding the role takes a full seat: any role
// other than VIEWER does (deny by default).
func IsSeatRole(name string) bool {
	n := strings.TrimSpace(name)
	return n != "" && !strings.EqualFold(n, ViewerRole)
}

func holdsSeatRole(roles []models.Role) bool {
	for _, r := range roles {
		if IsSeatRole(r.Name) {
			return true
		}
	}
	return false
}

func holdsAdmin(roles []models.Role) bool {
	for _, r := range roles {
		if strings.EqualFold(strings.TrimSpace(r.Name), AdminRole) {
			return true
		}
	}
	return false
}

func roleNames(roles []models.Role) []string {
	out := make([]string, len(roles))
	for i, r := range roles {
		out[i] = r.Name
	}
	return out
}

// cleanRoleNames trims the names and drops blanks.
func cleanRoleNames(in []string) []string {
	out := make([]string, 0, len(in))
	for _, r := range in {
		if name := strings.TrimSpace(r); name != "" {
			out = append(out, name)
		}
	}
	return out
}

// NewSeatUsage builds the seats report for a used count.
func NewSeatUsage(used int64) models.SeatUsage {
	remaining := int64(MaxFullSeats) - used
	if remaining < 0 {
		remaining = 0
	}
	return models.SeatUsage{
		Limit:     MaxFullSeats,
		Used:      used,
		Remaining: remaining,
		OverLimit: used > MaxFullSeats,
	}
}

// seatLimitError is the 409 returned when an action needs a free full seat.
func seatLimitError(used int64) *errors.AppError {
	u := NewSeatUsage(used)
	return errors.New(CodeSeatLimitReached,
		fmt.Sprintf("All %d full user seats are in use", MaxFullSeats),
		http.StatusConflict,
	).WithData(map[string]interface{}{
		"limit":     u.Limit,
		"used":      u.Used,
		"remaining": u.Remaining,
	})
}

func lastAdminError() *errors.AppError {
	return errors.New(CodeLastAdmin, "This is the last active admin account; keep at least one active admin", http.StatusConflict)
}

// internal logs err and returns the generic 500, keeping AppErrors as they are.
func (s *UserService) internal(msg string, err error) error {
	if appErr, ok := err.(*errors.AppError); ok && appErr.Code != errors.ErrDatabase.Code {
		return appErr
	}
	s.LogError(msg, utils.LogField("error", err))
	return errors.ErrInternalServer
}

func userNotFound() *errors.AppError {
	return errors.Wrap(errors.ErrNotFound.Code, "User not found", errors.ErrNotFound.StatusCode, nil)
}

// findUser loads a user in repo, mapping a missing row to 404.
func (s *UserService) findUser(repo repositories.UserRepositoryInterface, id uuid.UUID) (*models.User, error) {
	user, err := repo.FindByID(id)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, userNotFound()
		}
		return nil, s.internal("Failed to find user", err)
	}
	return user, nil
}

// guardLastAdmin refuses a change that leaves no active admin. user is the
// state before the change; it only matters when it is an active admin.
func (s *UserService) guardLastAdmin(repo repositories.UserRepositoryInterface, user *models.User) error {
	if !user.IsActive || !holdsAdmin(user.Roles) {
		return nil
	}
	admins, err := repo.CountActiveAdmins()
	if err != nil {
		return s.internal("Failed to count active admins", err)
	}
	if admins <= 1 {
		return lastAdminError()
	}
	return nil
}

func (s *UserService) record(ctx context.Context, action string, user *models.User, summary string, details map[string]interface{}) {
	if s.activity == nil {
		return
	}
	s.activity.Record(ctx, activitylog.Entry{
		Action:     action,
		EntityType: "user",
		EntityID:   user.ID.String(),
		Result:     activitylog.ResultSuccess,
		Summary:    summary,
		Details:    details,
	})
}

// GetSeatUsage reports the full seats in use against MaxFullSeats.
func (s *UserService) GetSeatUsage(ctx context.Context) (*models.SeatUsage, error) {
	used, err := s.userRepo.CountSeats()
	if err != nil {
		return nil, s.internal("Failed to count seats", err)
	}
	u := NewSeatUsage(used)
	return &u, nil
}

// SetUserStatus activates or deactivates a user. Reactivating a user holding
// a seat role needs a free seat (409 SEAT_LIMIT_REACHED otherwise; roles are
// never changed here). Deactivating the last active admin is refused (409
// LAST_ADMIN). Setting the current value again is a no-op.
func (s *UserService) SetUserStatus(ctx context.Context, id uuid.UUID, active bool) (*models.UserResponse, error) {
	var out *models.User
	err := s.userRepo.WithSeatLock(func(tx repositories.UserRepositoryInterface) error {
		user, err := s.findUser(tx, id)
		if err != nil {
			return err
		}
		if err := s.checkStatusChange(tx, user, active); err != nil {
			return err
		}
		if user.IsActive != active {
			if err := tx.SetActive(id, active); err != nil {
				return s.internal("Failed to update user status", err)
			}
			user.IsActive = active
		}
		out = user
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.LogInfo("User status changed", utils.LogField("user_id", id), utils.LogField("is_active", active))
	s.record(ctx, activitylog.ActionUserStatusChanged, out,
		fmt.Sprintf("%s set %s", out.Username, map[bool]string{true: "active", false: "inactive"}[active]),
		map[string]interface{}{"target_username": out.Username, "is_active": active})
	return s.userToResponse(out), nil
}

// checkStatusChange applies the seat and last-admin rules to setting
// user.IsActive to active. user is the state before the change.
func (s *UserService) checkStatusChange(tx repositories.UserRepositoryInterface, user *models.User, active bool) error {
	if user.IsActive == active {
		return nil
	}
	if !active {
		return s.guardLastAdmin(tx, user)
	}
	if !holdsSeatRole(user.Roles) {
		return nil
	}
	used, err := tx.CountSeats()
	if err != nil {
		return s.internal("Failed to count seats", err)
	}
	if used >= MaxFullSeats {
		return seatLimitError(used)
	}
	return nil
}

// ReplaceUserRoles replaces all roles of a user. Taking an active user from no
// seat role to a seat role needs a free seat (409 SEAT_LIMIT_REACHED).
// Moving between seat roles or down to VIEWER is always allowed, except that
// removing ADMIN from the last active admin is refused (409 LAST_ADMIN).
// Unknown role names are ignored; if none of the names exist the request is
// refused (400 ROLE_NOT_FOUND) and nothing changes.
func (s *UserService) ReplaceUserRoles(ctx context.Context, id uuid.UUID, requested []string) (*models.UserResponse, error) {
	roles := cleanRoleNames(requested)
	if len(roles) == 0 {
		return nil, errors.New("ROLE_REQUIRED", "At least one role is required", http.StatusBadRequest)
	}

	var before []string
	var out *models.User
	err := s.userRepo.WithSeatLock(func(tx repositories.UserRepositoryInterface) error {
		user, err := s.findUser(tx, id)
		if err != nil {
			return err
		}
		before = roleNames(user.Roles)

		// Seats in use before the change; the user is in it iff it already holds a seat role.
		var used int64
		if user.IsActive && !holdsSeatRole(user.Roles) {
			if used, err = tx.CountSeats(); err != nil {
				return s.internal("Failed to count seats", err)
			}
		}

		if err := tx.AssignRoles(id, roles); err != nil {
			return s.internal("Failed to assign roles", err)
		}
		updated, err := s.findUser(tx, id)
		if err != nil {
			return err
		}
		if len(updated.Roles) == 0 {
			return errors.New("ROLE_NOT_FOUND", "None of the requested roles exist", http.StatusBadRequest)
		}

		if user.IsActive && !holdsSeatRole(user.Roles) && holdsSeatRole(updated.Roles) && used >= MaxFullSeats {
			return seatLimitError(used)
		}
		// Removing ADMIN must leave at least one active admin (counted after the change).
		if user.IsActive && holdsAdmin(user.Roles) && !holdsAdmin(updated.Roles) {
			admins, err := tx.CountActiveAdmins()
			if err != nil {
				return s.internal("Failed to count active admins", err)
			}
			if admins == 0 {
				return lastAdminError()
			}
		}
		out = updated
		return nil
	})
	if err != nil {
		return nil, err
	}

	after := roleNames(out.Roles)
	s.LogInfo("User roles replaced", utils.LogField("user_id", id))
	s.record(ctx, activitylog.ActionUserRolesChanged, out,
		fmt.Sprintf("Roles of %s set to %s", out.Username, strings.Join(after, ", ")),
		map[string]interface{}{"target_username": out.Username, "roles_before": before, "roles_after": after})
	return s.userToResponse(out), nil
}

// DeleteUser deletes a user: a soft delete (deleted_at) by default, or a
// permanent delete of the user, its role links, MFA setups, trusted devices,
// refresh tokens and confidentiality agreements when hard is true. Deleting
// the last active admin is refused (409 LAST_ADMIN) either way.
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID, hard bool) error {
	var target *models.User
	err := s.userRepo.WithSeatLock(func(tx repositories.UserRepositoryInterface) error {
		user, err := tx.FindByID(id)
		if err != nil {
			if !errors.Is(err, errors.ErrNotFound) {
				return s.internal("Failed to find user", err)
			}
			// A soft-deleted user can still be removed for good.
			if !hard {
				return userNotFound()
			}
			user = &models.User{ID: id}
		}
		if err := s.guardLastAdmin(tx, user); err != nil {
			return err
		}
		target = user
		if hard {
			if err := tx.HardDelete(id); err != nil {
				if errors.Is(err, errors.ErrNotFound) {
					return userNotFound()
				}
				return s.internal("Failed to hard delete user", err)
			}
			return nil
		}
		if err := tx.Delete(id); err != nil {
			return s.internal("Failed to delete user", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.LogInfo("User deleted successfully", utils.LogField("user_id", id), utils.LogField("hard", hard))
	s.record(ctx, activitylog.ActionUserDeleted, target,
		fmt.Sprintf("User %s deleted", target.Username),
		map[string]interface{}{"target_username": target.Username, "hard": hard})
	return nil
}

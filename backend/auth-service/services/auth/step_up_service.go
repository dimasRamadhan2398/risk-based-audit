package services

import (
	"context"
	"net/http"
	"time"
	"unicode/utf8"

	"auth-service/models"
	baseService "auth-service/pkg/base"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/kafka"
	"auth-service/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Password step-up (docs/org-structure-model-plan.md 9.3).
const (
	// StepUpMaxFailures wrong passwords per user within StepUpFailureWindow
	// lock POST /auth/verify-password for that user until the window ends.
	StepUpMaxFailures   = 5
	StepUpFailureWindow = 15 * time.Minute
	// stepUpAttemptIdentifier prefixes the user id passed to the login-attempt
	// helpers, so the Redis key is auth:loginattempt:stepup:<user_id> and never
	// shares a counter with a login identifier (username).
	stepUpAttemptIdentifier = "stepup:"
)

// stepUpPurposes is the purpose whitelist.
var stepUpPurposes = map[string]bool{
	utils.StepUpPurposeOrgLevels: true,
}

// Outcomes, used as the log/event reason.
const (
	StepUpOutcomeSuccess        = "success"
	StepUpOutcomeWrongPassword  = "wrong_password"
	StepUpOutcomeLockedOut      = "locked_out"
	StepUpOutcomeInvalidPurpose = "invalid_purpose"
	StepUpOutcomeUnavailable    = "unavailable"
	StepUpOutcomeUserInvalid    = "user_invalid"
)

// Errors. A wrong password is 403, not 401: the frontend treats a 401 from
// auth-service (outside login/change-password) as an ended session.
var (
	ErrStepUpPurposeInvalid = apperrors.New("STEP_UP_PURPOSE_INVALID",
		"This action cannot be confirmed with a password.", http.StatusBadRequest)
	ErrStepUpPasswordInvalid = apperrors.New("STEP_UP_PASSWORD_INVALID",
		"The password is incorrect.", http.StatusForbidden)
	ErrStepUpLocked = apperrors.New("STEP_UP_LOCKED",
		"Too many incorrect passwords. Please try again in 15 minutes.", http.StatusTooManyRequests)
	ErrStepUpUnavailable = apperrors.New("STEP_UP_UNAVAILABLE",
		"The password cannot be checked right now. Please try again.", http.StatusServiceUnavailable)
)

// StepUpUserFinder is the part of the user repository the step-up needs.
type StepUpUserFinder interface {
	FindByID(id uuid.UUID) (*models.User, error)
}

// StepUpAttemptStore is the part of repositories.ICacheRepository the
// step-up needs (the login-attempt helpers).
type StepUpAttemptStore interface {
	TrackLoginAttempt(ctx context.Context, identifier string, expiration time.Duration) (int64, error)
	GetLoginAttempts(ctx context.Context, identifier string) (int64, error)
	ClearLoginAttempts(ctx context.Context, identifier string) error
}

// StepUpEventPublisher is the part of kafka.IEventPublisher the step-up needs.
type StepUpEventPublisher interface {
	Publish(ctx context.Context, event *kafka.Event) error
}

// VerifyPasswordInput: UserID and Username come from the access token.
type VerifyPasswordInput struct {
	UserID    string
	Username  string
	Password  string
	Purpose   string
	IP        string
	UserAgent string
}

// StepUpServiceInterface issues step-up tokens.
type StepUpServiceInterface interface {
	VerifyPassword(ctx context.Context, in VerifyPasswordInput) (*models.VerifyPasswordResponse, error)
}

// StepUpService checks the caller's password and issues a step-up token.
type StepUpService struct {
	*baseService.BaseService
	users     StepUpUserFinder
	attempts  StepUpAttemptStore
	publisher StepUpEventPublisher
	jwtSecret string
	now       func() time.Time
}

// NewStepUpService: attempts or publisher may be nil (no Redis: every call
// answers 503; no Kafka: no event, logs only).
func NewStepUpService(users StepUpUserFinder, attempts StepUpAttemptStore, publisher StepUpEventPublisher, jwtSecret string) *StepUpService {
	return &StepUpService{
		BaseService: baseService.NewBaseService(),
		users:       users,
		attempts:    attempts,
		publisher:   publisher,
		jwtSecret:   jwtSecret,
		now:         time.Now,
	}
}

// VerifyPassword: purpose whitelist, lockout check, password check, token.
func (s *StepUpService) VerifyPassword(ctx context.Context, in VerifyPasswordInput) (*models.VerifyPasswordResponse, error) {
	if !stepUpPurposes[in.Purpose] {
		// The purpose is caller input: log only a bounded prefix of it.
		s.record(ctx, in, StepUpOutcomeInvalidPurpose, 0, "", zap.String("requested_purpose", boundedString(in.Purpose, 64)))
		return nil, ErrStepUpPurposeInvalid
	}

	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		s.record(ctx, in, StepUpOutcomeUserInvalid, 0, "")
		return nil, apperrors.ErrUnauthorized
	}
	if s.attempts == nil {
		s.record(ctx, in, StepUpOutcomeUnavailable, 0, "")
		return nil, ErrStepUpUnavailable
	}
	counterID := stepUpAttemptIdentifier + userID.String()

	// Fail closed: without the counter the lockout cannot be enforced.
	failed, err := s.attempts.GetLoginAttempts(ctx, counterID)
	if err != nil {
		s.record(ctx, in, StepUpOutcomeUnavailable, 0, "", zap.Error(err))
		return nil, ErrStepUpUnavailable
	}
	if failed >= StepUpMaxFailures {
		s.record(ctx, in, StepUpOutcomeLockedOut, failed, "")
		return nil, ErrStepUpLocked
	}

	user, err := s.users.FindByID(userID)
	if err != nil || user == nil || !user.IsActive {
		if err != nil && !apperrors.Is(err, apperrors.ErrNotFound) {
			s.record(ctx, in, StepUpOutcomeUnavailable, failed, "", zap.Error(err))
			return nil, ErrStepUpUnavailable
		}
		s.record(ctx, in, StepUpOutcomeUserInvalid, failed, "")
		return nil, apperrors.ErrUnauthorized
	}

	if !utils.CheckPassword(in.Password, user.PasswordHash) {
		count, trackErr := s.attempts.TrackLoginAttempt(ctx, counterID, StepUpFailureWindow)
		if trackErr != nil {
			s.LogError("Step-up failure counter not updated", zap.String("user_id", in.UserID), zap.Error(trackErr))
			count = failed + 1
		}
		s.record(ctx, in, StepUpOutcomeWrongPassword, count, "")
		if count >= StepUpMaxFailures {
			s.record(ctx, in, StepUpOutcomeLockedOut, count, "")
			return nil, ErrStepUpLocked
		}
		return nil, ErrStepUpPasswordInvalid
	}

	if failed > 0 {
		if err := s.attempts.ClearLoginAttempts(ctx, counterID); err != nil {
			s.LogWarning("Step-up failure counter not cleared", zap.String("user_id", in.UserID), zap.Error(err))
		}
	}

	// sub is the access token's user_id string exactly as received, so it
	// compares equal in master-service RequireStepUp.
	token, jti, err := utils.GenerateStepUpToken(in.UserID, in.Purpose, s.jwtSecret, s.now())
	if err != nil {
		s.LogError("Failed to sign step-up token", zap.String("user_id", in.UserID), zap.Error(err))
		return nil, apperrors.ErrInternalServer
	}
	s.record(ctx, in, StepUpOutcomeSuccess, 0, jti)
	return &models.VerifyPasswordResponse{StepUpToken: token, ExpiresIn: int(utils.StepUpTTL / time.Second)}, nil
}

// record writes the structured log and emits the security.step_up event.
// It never receives the password or the token; jti identifies an issued
// token without being usable as one.
func (s *StepUpService) record(ctx context.Context, in VerifyPasswordInput, outcome string, failedAttempts int64, jti string, extra ...zap.Field) {
	success := outcome == StepUpOutcomeSuccess
	purpose := ""
	if stepUpPurposes[in.Purpose] {
		purpose = in.Purpose
	}
	fields := []zap.Field{
		zap.String("event", string(kafka.EventStepUp)),
		zap.String("outcome", outcome),
		zap.String("user_id", in.UserID),
		zap.String("username", in.Username),
		zap.String("purpose", purpose),
		zap.String("ip", in.IP),
		zap.Int64("failed_attempts", failedAttempts),
	}
	if jti != "" {
		fields = append(fields, zap.String("jti", jti))
	}
	fields = append(fields, extra...)
	switch outcome {
	case StepUpOutcomeSuccess:
		s.LogInfo("Step-up password verified", fields...)
	case StepUpOutcomeUnavailable:
		s.LogError("Step-up password check unavailable", fields...)
	default:
		s.LogWarning("Step-up password verification rejected", fields...)
	}

	if s.publisher == nil {
		return
	}
	data := map[string]interface{}{
		"outcome":         outcome,
		"purpose":         purpose,
		"failed_attempts": failedAttempts,
	}
	if jti != "" {
		data["jti"] = jti
	}
	event := &kafka.Event{
		Type:      kafka.EventStepUp,
		Timestamp: s.now().UTC(),
		UserID:    in.UserID,
		Username:  in.Username,
		IPAddress: in.IP,
		UserAgent: in.UserAgent,
		Success:   success,
		Data:      data,
	}
	if !success {
		event.ErrorReason = outcome
	}
	if err := s.publisher.Publish(ctx, event); err != nil {
		s.LogWarning("Failed to publish step-up event", zap.String("user_id", in.UserID), zap.Error(err))
	}
}

// boundedString cuts s to at most n bytes on a rune boundary.
func boundedString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

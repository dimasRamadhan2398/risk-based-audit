package mfa

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	pkgKafka "auth-service/pkg/kafka"
	"auth-service/pkg/redis"
	"auth-service/pkg/utils"
	"auth-service/repositories"
	"auth-service/services/email"

	"github.com/google/uuid"
)

// MfaServiceInterface defines the MFA service interface
type MfaServiceInterface interface {
	SetupMFA(ctx context.Context, userID uuid.UUID, mfaSetupReq models.MFAType, ipAddress string) (*models.MfaSetupResponse, error)
	VerifyMFA(ctx context.Context, userID uuid.UUID, code string, ipAddress string) (bool, error)
	EnableMFA(ctx context.Context, userID uuid.UUID, code string, ipAddress string) error
	DisableMFA(ctx context.Context, userID uuid.UUID, password string, ipAddress string) error
	GetMFAStatus(ctx context.Context, userID uuid.UUID) (*models.MfaSetupResponse, error)
	GenerateEmailCode(ctx context.Context, userID uuid.UUID, email string, ipAddress string) error
	VerifyEmailCode(ctx context.Context, userID uuid.UUID, code string) (bool, error)
}

// MfaService handles MFA operations
type MfaService struct {
	mfaRepo      repositories.MFASetupRepositoryInterface
	userRepo     repositories.UserRepositoryInterface
	publisher    pkgKafka.IEventPublisher
	redis        *redis.Client
	emailService email.EmailServiceInterface
}

// NewMfaService creates a new MFA service
func NewMfaService(
	mfaRepo repositories.MFASetupRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	publisher pkgKafka.IEventPublisher,
	rdb *redis.Client,
	emailService email.EmailServiceInterface,
) MfaServiceInterface {
	return &MfaService{
		mfaRepo:      mfaRepo,
		userRepo:     userRepo,
		publisher:    publisher,
		redis:        rdb,
		emailService: emailService,
	}
}

// SetupMFA sets up MFA for a user
func (s *MfaService) SetupMFA(ctx context.Context, userID uuid.UUID, mfaType models.MFAType, ipAddress string) (*models.MfaSetupResponse, error) {
	// Check if MFA already exists
	existingMFA, _ := s.mfaRepo.FindByUserID(userID)
	if existingMFA != nil && existingMFA.IsEnabled {
		return nil, apperrors.Wrap("MFA_ALREADY_ENABLED", "MFA is already enabled", 400, nil)
	}

	// Get user for email
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, apperrors.Wrap("USER_NOT_FOUND", "User not found", 404, nil)
	}

	// Create or update MFA setup
	mfa := &models.MFASetup{
		ID:        uuid.New(),
		UserID:    userID,
		MFAType:   mfaType,
		IsEnabled: false,
	}

	response := &models.MfaSetupResponse{
		MFAType:   mfaType,
		IsEnabled: false,
	}

	switch mfaType {
	case models.MFATypeTOTP:
		secret, err := generateSecret()
		if err != nil {
			return nil, apperrors.Wrap("GENERATE_SECRET_FAILED", "Failed to generate secret", 500, nil)
		}
		response.Secret = secret
		response.QRCodeURL = generateTOTPURL(user.Email, secret, "auth-service")
	case models.MFATypeEmail:
		mfa.SecretKey = ""

	default:
        return nil, apperrors.Wrap("INVALID_MFA_TYPE", "Unsupported MFA type", 400, nil)
    }


	if existingMFA != nil {
		mfa.ID = existingMFA.ID
		if err := s.mfaRepo.Update(mfa); err != nil {
			return nil, apperrors.Wrap("UPDATE_MFA_FAILED", "Failed to update MFA", 500, nil)
		}
	} else {
		if err := s.mfaRepo.Create(mfa); err != nil {
			return nil, apperrors.Wrap("CREATE_MFA_FAILED", "Failed to create MFA", 500, nil)
		}
	}

	// Publish MFA setup event
	if s.publisher != nil {
		s.publisher.PublishMFAEvent(ctx, userID.String(), user.Username, pkgKafka.EventMFAEnabled, true, ipAddress)
	}

	return response, nil
}

// VerifyMFA verifies an MFA code
func (s *MfaService) VerifyMFA(ctx context.Context, userID uuid.UUID, code string, ipAddress string) (bool, error) {
	mfa, err := s.mfaRepo.FindByUserID(userID)
	if err != nil {
		return false, apperrors.Wrap("MFA_NOT_SETUP", "MFA is not set up", 404, nil)
	}

	if !mfa.IsEnabled {
		return false, apperrors.Wrap("MFA_NOT_ENABLED", "MFA is not enabled", 400, nil)
	}

	// Get user for logging
	user, _ := s.userRepo.FindByID(userID)

	// Verify based on type
	var valid bool
	switch mfa.MFAType {
	case models.MFATypeTOTP:
		valid, _ = s.verifyTOTP(mfa.SecretKey, code)
	case models.MFATypeEmail:
		valid, _ = s.verifyEmailCode(ctx, userID, code)
	default:
		return false, apperrors.Wrap("INVALID_MFA_TYPE", "Invalid MFA type", 400, nil)
	}

	// Publish MFA verification event
	if s.publisher != nil {
		username := ""
		if user != nil {
			username = user.Username
		}
		if valid {
			s.publisher.PublishMFAEvent(ctx, userID.String(), username, pkgKafka.EventMFAVerified, true, ipAddress)
		} else {
			s.publisher.PublishMFAEvent(ctx, userID.String(), username, pkgKafka.EventMFAFailed, false, ipAddress)
		}
	}

	return valid, nil
}

// EnableMFA enables MFA for a user after verifying the code
func (s *MfaService) EnableMFA(ctx context.Context, userID uuid.UUID, code string, ipAddress string) error {
	mfa, err := s.mfaRepo.FindByUserID(userID)
	if err != nil {
		return apperrors.Wrap("MFA_NOT_SETUP", "MFA is not set up", 404, nil)
	}

	if mfa.IsEnabled {
		return apperrors.Wrap("MFA_ALREADY_ENABLED", "MFA is already enabled", 400, nil)
	}

	// Verify the code first
	var valid bool
	switch mfa.MFAType {
	case models.MFATypeTOTP:
		valid, _ = s.verifyTOTP(mfa.SecretKey, code)
	case models.MFATypeEmail:
		valid, _ = s.verifyEmailCode(ctx, userID, code)
	default:
		return apperrors.Wrap("INVALID_MFA_TYPE", "Invalid MFA type", 400, nil)
	}

	if !valid {
		return apperrors.Wrap("INVALID_CODE", "Invalid verification code", 400, nil)
	}

	// Enable MFA
	if err := s.mfaRepo.Enable(userID); err != nil {
		return apperrors.Wrap("ENABLE_MFA_FAILED", "Failed to enable MFA", 500, nil)
	}

	s.publisher.PublishMFAEvent(ctx, userID.String(), mfa.User.Username, pkgKafka.EventMFAEnabled, true, ipAddress)
	return nil
}

// DisableMFA disables MFA for a user
func (s *MfaService) DisableMFA(ctx context.Context, userID uuid.UUID, password string, ipAddress string) error {
	// Verify password first
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return apperrors.Wrap("USER_NOT_FOUND", "User not found", 404, nil)
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return apperrors.ErrInvalidCredentials
	}

	// Disable MFA
	if err := s.mfaRepo.Disable(userID); err != nil {
		return apperrors.Wrap("DISABLE_MFA_FAILED", "Failed to disable MFA", 500, nil)
	}

	// Publish MFA disabled event
	if s.publisher != nil {
		s.publisher.PublishMFAEvent(ctx, userID.String(), user.Username, pkgKafka.EventMFADisabled, true, ipAddress)
	}

	return nil
}

// GetMFAStatus gets the MFA status for a user
func (s *MfaService) GetMFAStatus(ctx context.Context, userID uuid.UUID) (*models.MfaSetupResponse, error) {
	mfa, err := s.mfaRepo.FindByUserID(userID)
	if err != nil {
		// Return not setup status
		return &models.MfaSetupResponse{
			IsEnabled: false,
		}, nil
	}

	return &models.MfaSetupResponse{
		MFAType:   mfa.MFAType,
		IsEnabled: mfa.IsEnabled,
	}, nil
}

// GenerateEmailCode generates and stores an email verification code
func (s *MfaService) GenerateEmailCode(ctx context.Context, userID uuid.UUID, email string, ipAddress string) error {
	// Generate 6-digit code
	code, err := generateCode(6)
	if err != nil {
		return apperrors.Wrap("GENERATE_CODE_FAILED", "Failed to generate code", 500, nil)
	}

	// Store code in Redis with 5 minute expiry
	key := fmt.Sprintf("mfa:email:%s", userID.String())
	codeData := map[string]string{
		"code":  code,
		"email": email,
	}
	data, _ := json.Marshal(codeData)

	if err := s.redis.Set(ctx, key, string(data), 5*time.Minute).Err(); err != nil {
		return apperrors.Wrap("STORE_CODE_FAILED", "Failed to store code", 500, nil)
	}

	// Send email with the code
	if s.emailService != nil {
		_, err := s.emailService.SendOTPEmail(ctx, email, code, ipAddress, "Unknown Device")
		if err != nil {
			return err
		}
	} else {
		// For now, we'll log it if email service is not available
		fmt.Printf("MFA Email Code for user %s: %s\n", userID.String(), code)
	}

	// Publish MFA code generation event
	if s.publisher != nil {
		s.publisher.PublishMFAEvent(ctx, userID.String(), email, pkgKafka.EventMFAEnabled, true, ipAddress)
	}

	return nil
}

// VerifyEmailCode verifies an email code
func (s *MfaService) verifyEmailCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	key := fmt.Sprintf("mfa:email:%s", userID.String())
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return false, nil
	}

	var codeData map[string]string
	if err := json.Unmarshal([]byte(data), &codeData); err != nil {
		return false, nil
	}

	// Use constant-time comparison to prevent timing attacks
	valid := subtle.ConstantTimeCompare([]byte(codeData["code"]), []byte(code)) == 1

	if valid {
		// Delete the code after successful verification
		s.redis.Del(ctx, key)
	}

	return valid, nil
}

// VerifyEmailCode is the public method to verify email code
func (s *MfaService) VerifyEmailCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	return s.verifyEmailCode(ctx, userID, code)
}

// verifyTOTP verifies a TOTP code
func (s *MfaService) verifyTOTP(secret, code string) (bool, error) {
	// Remove spaces from secret
	secret = strings.ReplaceAll(secret, " ", "")

	// Decode base32 secret
	secretBytes, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil {
		return false, nil
	}

	// Get current time step
	currentStep := time.Now().Unix() / 30

	// Check current step and adjacent steps for tolerance
	for _, step := range []int64{currentStep, currentStep - 1, currentStep + 1} {
		expectedCode := generateHOTP(secretBytes, step)
		if subtle.ConstantTimeCompare([]byte(expectedCode), []byte(code)) == 1 {
			return true, nil
		}
	}

	return false, nil
}

// generateSecret generates a random secret for TOTP
func generateSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(secret), nil
}

// generateCode generates a numeric code of specified length
func generateCode(length int) (string, error) {
	code := make([]byte, length)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = byte(n.Int64()) + '0'
	}
	return string(code), nil
}

// generateTOTPURL generates the TOTP URL for QR code
func generateTOTPURL(email, secret, issuer string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
		issuer, email, secret, issuer)
}

// generateHOTP generates an HOTP code
func generateHOTP(secret []byte, counter int64) string {
	// Convert counter to bytes
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = byte(counter & 0xff)
		counter >>= 8
	}

	// Calculate HMAC-SHA1
	h := hmac.New(sha1.New, secret)
	h.Write(buf)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	truncated := (int(hash[offset])&0x7f)<<24 |
		(int(hash[offset+1])&0xff)<<16 |
		(int(hash[offset+2])&0xff)<<8 |
		(int(hash[offset+3]) & 0xff)

	code := truncated % 1000000
	return fmt.Sprintf("%06d", code)
}

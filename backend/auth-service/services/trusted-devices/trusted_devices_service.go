package trusteddevices

import (
	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	pkgKafka "auth-service/pkg/kafka"
	"auth-service/pkg/redis"
	"auth-service/repositories"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TrustedDevicesControllerInterface defines the trusted devices controller interface
type TrustedDevicesServiceInterface interface {
	GetTrustedDevices(ctx context.Context, userID uuid.UUID) ([]models.TrustedDeviceResponse, error)
	GenerateEnrollmentQR(ctx context.Context, userID uuid.UUID) (*models.QREnrollmentResponse, error)
	VerifyEnrollmentToken(ctx context.Context, userID uuid.UUID, token string, req models.TrustedDeviceRequest, ipAddress, userAgent string) error
	UnenrollTrustedDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID, ipAddress, userAgent string) error
}

type TrustedDevicesService struct {
	devicesRepo repositories.ITrustedDeviceRepository
	userRepo    repositories.IUserRepository
	publisher   pkgKafka.IEventPublisher
	redis       *redis.Client
}

func (s TrustedDevicesService) EnrollTrustedDevice(context context.Context, id uuid.UUID, name string, ipAddress string, userAgent string) any {
	panic("unimplemented")
}

// GenerateEnrollmentQR implements TrustedDevicesServiceInterface.
func (s *TrustedDevicesService) GenerateEnrollmentQR(ctx context.Context, userID uuid.UUID) (*models.QREnrollmentResponse, error) {
	tokenBytes := make([]byte, 32)

	_, err := rand.Read(tokenBytes)
	if err != nil {
		return nil, apperrors.Wrap("FAILED_TO_GENERATE_TOKEN", "Failed to generate token", 500, nil)
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)

	key := fmt.Sprintf("device:enroll:%s:%s", userID.String(), uuid.New().String())

	value := map[string]interface{}{
		"user_id": userID.String(),
		"token":   tokenBytes,
	}

	jsonValue, _ := json.Marshal(value)
	if err := s.redis.Set(ctx, key, jsonValue, 5*time.Minute).Err(); err != nil {
		return nil, apperrors.Wrap("STORE_TOKEN_FAILED", "Failed to store enrollment token", 500, nil)
	}

	qrContent := fmt.Sprintf("myapp://trust-device?token=%s", token)

	return &models.QREnrollmentResponse{
		Token:     token,
		QRCodeURL: qrContent,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}, nil
}

// VerifyEnrollmentToken implements TrustedDevicesServiceInterface.
func (s *TrustedDevicesService) VerifyEnrollmentToken(ctx context.Context, userID uuid.UUID, token string, req models.TrustedDeviceRequest, ipAddress string, userAgent string) error {
	key := fmt.Sprintf("device:enroll:%s:%s", userID.String(), token)
	storedUserID, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return apperrors.Wrap("INVALID_TOKEN", "Invalid or expired enrollment token", 400, nil)
	}

	if storedUserID != userID.String() {
		return apperrors.Wrap("TOKEN_USER_MISMATCH", "Token does not belong to this user", 403, nil)
	}

	existing, _ := s.devicesRepo.FindByFingerprint(userID, req.DeviceFingerprint)

	if existing != nil {
		return apperrors.Wrap("DEVICE_ALREADY_ENROLLED", "Device is already enrolled", 400, nil)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return apperrors.Wrap("USER_NOT_FOUND", "User not found", 400, nil)
	}

	expiresAt := time.Now().Add(90 * 24 * time.Hour)

	device := models.TrustedDevice{
		ID:                uuid.New(),
		UserID:            userID,
		DeviceFingerprint: req.DeviceFingerprint,
		DeviceName:        req.DeviceName,
		DeviceType:        req.DeviceType,
		UserAgent:         userAgent,
		IPAddress:         ipAddress,
		ExpiresAt:         &expiresAt,
		CreatedAt:         time.Now(),
	}

	if err := s.devicesRepo.Create(&device); err != nil {
		return err
	}

	s.redis.Del(ctx, key)

	s.publisher.PublishSecurityEvent(ctx, userID.String(), user.Username, pkgKafka.EventTrustedDeviceAdded, ipAddress, map[string]interface{}{
		"device_id":   device.ID.String(),
		"device_name": device.DeviceName,
		"device_type": device.DeviceType,
	})

	return nil
}

// GetTrustedDevices implements TrustedDevicesServiceInterface.
func (s *TrustedDevicesService) GetTrustedDevices(ctx context.Context, userID uuid.UUID) ([]models.TrustedDeviceResponse, error) {
	devices, err := s.devicesRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	deviceResults := make([]models.TrustedDeviceResponse, len(devices))
	for _, device := range devices {
		deviceResults = append(deviceResults, models.TrustedDeviceResponse{
			ID:                device.ID,
			DeviceFingerprint: device.DeviceFingerprint,
			DeviceName:        device.DeviceName,
			DeviceType:        device.DeviceType,
			UserAgent:         device.UserAgent,
			IPAddress:         device.IPAddress,
			ExpiresAt:         device.ExpiresAt,
			CreatedAt:         device.CreatedAt,
		})
	}
	return deviceResults, nil
}

// NewTrustedDevicesService creates a new trusted devices service
func NewTrustedDevicesService(
	devicesRepo repositories.ITrustedDeviceRepository,
	userRepo repositories.IUserRepository,
	publisher pkgKafka.IEventPublisher,
	redis *redis.Client,
) TrustedDevicesServiceInterface {
	return &TrustedDevicesService{
		devicesRepo: devicesRepo,
		userRepo:    userRepo,
		publisher:   publisher,
		redis:       redis,
	}
}

// UnenrollTrustedDevice removes a trusted device
func (s *TrustedDevicesService) UnenrollTrustedDevice(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID, ipAddress, userAgent string) error {
	device, err := s.devicesRepo.FindByID(deviceID)
	if err != nil {
		return apperrors.Wrap("DEVICE_NOT_FOUND", "Device not found", 404, nil)
	}

	if device.UserID != userID {
		return apperrors.Wrap("FORBIDDEN", "Device does not belong to this user", 404, nil)
	}

	if err := s.devicesRepo.Delete(deviceID); err != nil {
		return apperrors.Wrap("UNENROLL_DEVICE_FAILED", "Failed to unenroll device", 500, nil)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return apperrors.Wrap("USER_NOT_FOUND", "User not found", 404, nil)
	}

	if s.publisher != nil {
		s.publisher.PublishSecurityEvent(ctx, userID.String(), user.Username, pkgKafka.EventTrustedDeviceRemoved, ipAddress, map[string]interface{}{
			"device_id": deviceID.String(),
		})
	}

	return nil
}

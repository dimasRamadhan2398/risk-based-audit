package repositories

import (
	"strings"
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ITrustedDeviceRepository is an alias for the trusted device repository interface
type ITrustedDeviceRepository = TrustedDeviceRepositoryInterface

// TrustedDeviceRepositoryInterface defines the trusted device repository interface
type TrustedDeviceRepositoryInterface interface {
	Create(device *models.TrustedDevice) error
	Delete(id uuid.UUID) error
	DeleteByFingerprint(userID uuid.UUID, fingerprint string) error
	FindByID(id uuid.UUID) (*models.TrustedDevice, error)
	FindByUserID(userID uuid.UUID) ([]*models.TrustedDevice, error)
	FindByFingerprint(userID uuid.UUID, fingerprint string) (*models.TrustedDevice, error)
	IsExpired(device *models.TrustedDevice) bool
	// IsTrusted reports whether the device may skip the MFA step.
	IsTrusted(device *models.TrustedDevice) bool
	// RecordSeen registers the (user, fingerprint) device or refreshes the
	// existing row. It never creates a second row for the same device.
	RecordSeen(s DeviceSighting) (*models.TrustedDevice, error)
	DeleteExpiredDevices(userID uuid.UUID) error
}

// TrustedDeviceTTL is how long "trust this device" lasts.
const TrustedDeviceTTL = 90 * 24 * time.Hour

// DeviceSighting is one sign-in or MFA check from a device.
type DeviceSighting struct {
	UserID      uuid.UUID
	Fingerprint string
	Name        string
	Type        string
	UserAgent   string
	IPAddress   string
	// Trust marks the device trusted for TrustedDeviceTTL. false records the
	// device without changing its trust state (a new row is session-only).
	Trust bool
	Now   time.Time
}

// TrustedDeviceRepository handles trusted device data operations
type TrustedDeviceRepository struct {
	*BaseRepository
}

// NewTrustedDeviceRepository creates a new trusted device repository
func NewTrustedDeviceRepository(db *gorm.DB) ITrustedDeviceRepository {
	return &TrustedDeviceRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new trusted device
func (r *TrustedDeviceRepository) Create(device *models.TrustedDevice) error {
	return r.BaseRepository.Create(device)
}

// Delete deletes a trusted device
func (r *TrustedDeviceRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.TrustedDevice{ID: id})
}

// DeleteByFingerprint deletes a trusted device by fingerprint
func (r *TrustedDeviceRepository) DeleteByFingerprint(userID uuid.UUID, fingerprint string) error {
	return r.GetDB().Where("user_id = ? AND device_fingerprint = ?", userID, fingerprint).Delete(&models.TrustedDevice{}).Error
}

// FindByID finds a trusted device by ID
func (r *TrustedDeviceRepository) FindByID(id uuid.UUID) (*models.TrustedDevice, error) {
	var device models.TrustedDevice
	if err := r.GetDB().First(&device, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &device, nil
}

// FindByUserID finds all trusted devices for a user
func (r *TrustedDeviceRepository) FindByUserID(userID uuid.UUID) ([]*models.TrustedDevice, error) {
	var devices []*models.TrustedDevice
	if err := r.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

// FindByFingerprint finds a trusted device by fingerprint
func (r *TrustedDeviceRepository) FindByFingerprint(userID uuid.UUID, fingerprint string) (*models.TrustedDevice, error) {
	var device models.TrustedDevice
	if err := r.GetDB().Where("user_id = ? AND device_fingerprint = ?", userID, fingerprint).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &device, nil
}

// IsExpired checks if a trusted device is expired
func (r *TrustedDeviceRepository) IsExpired(device *models.TrustedDevice) bool {
	return device.ExpiresAt != nil && device.ExpiresAt.Before(time.Now())
}

// IsTrusted implements TrustedDeviceRepositoryInterface.
func (r *TrustedDeviceRepository) IsTrusted(device *models.TrustedDevice) bool {
	return device != nil && !device.SessionOnly && !r.IsExpired(device)
}

// RecordSeen implements TrustedDeviceRepositoryInterface. When older code left
// several rows for one fingerprint, the newest is refreshed.
func (r *TrustedDeviceRepository) RecordSeen(s DeviceSighting) (*models.TrustedDevice, error) {
	if s.UserID == uuid.Nil || strings.TrimSpace(s.Fingerprint) == "" {
		return nil, apperrors.ErrBadRequest
	}
	now := s.Now
	if now.IsZero() {
		now = time.Now()
	}
	ua := truncate(s.UserAgent, 255)
	ip := normalizeIP(s.IPAddress)

	var device models.TrustedDevice
	err := r.GetDB().Transaction(func(tx *gorm.DB) error {
		err := tx.Where("user_id = ? AND device_fingerprint = ?", s.UserID, s.Fingerprint).
			Order("created_at DESC").Take(&device).Error
		isNew := false
		switch {
		case err == gorm.ErrRecordNotFound:
			isNew = true
			device = models.TrustedDevice{
				ID:                uuid.New(),
				UserID:            s.UserID,
				DeviceFingerprint: truncate(s.Fingerprint, 255),
				SessionOnly:       true,
				CreatedAt:         now,
			}
		case err != nil:
			return err
		}

		if name := strings.TrimSpace(s.Name); name != "" {
			device.DeviceName = truncate(name, 255)
		} else if device.DeviceName == "" {
			device.DeviceName = deviceNameFromUserAgent(ua)
		}
		if typ := strings.TrimSpace(s.Type); typ != "" {
			device.DeviceType = truncate(typ, 50)
		} else if device.DeviceType == "" {
			device.DeviceType = deviceTypeFromUserAgent(ua)
		}
		if ua != "" {
			device.UserAgent = ua
		}
		if ip != "" {
			device.IPAddress = ip
		}
		device.LastSeenAt = &now
		if s.Trust {
			expiresAt := now.Add(TrustedDeviceTTL)
			device.SessionOnly = false
			device.ExpiresAt = &expiresAt
		}
		if isNew {
			return tx.Create(&device).Error
		}
		// Save writes every column, including SessionOnly=false (Updates would
		// skip the zero value).
		return tx.Save(&device).Error
	})
	if err != nil {
		return nil, apperrors.ErrDatabase
	}
	return &device, nil
}

// normalizeIP keeps the client address of an X-Forwarded-For chain
// ("client, proxy1, proxy2") so it fits ip_address varchar(45).
func normalizeIP(ip string) string {
	if i := strings.IndexByte(ip, ','); i >= 0 {
		ip = ip[:i]
	}
	return truncate(strings.TrimSpace(ip), 45)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// deviceNameFromUserAgent is a fallback label ("Chrome on macOS") for callers
// that did not send device_name.
func deviceNameFromUserAgent(ua string) string {
	if ua == "" {
		return "Unknown device"
	}
	l := strings.ToLower(ua)
	browser := "Browser"
	switch {
	case strings.Contains(l, "edg/") || strings.Contains(l, "edge"):
		browser = "Edge"
	case strings.Contains(l, "firefox"):
		browser = "Firefox"
	case strings.Contains(l, "chrome") && !strings.Contains(l, "chromium"):
		browser = "Chrome"
	case strings.Contains(l, "safari"):
		browser = "Safari"
	}
	os := ""
	switch {
	case strings.Contains(l, "windows"):
		os = "Windows"
	case strings.Contains(l, "android"):
		os = "Android"
	case strings.Contains(l, "iphone") || strings.Contains(l, "ipad"):
		os = "iOS"
	case strings.Contains(l, "mac os x"):
		os = "macOS"
	case strings.Contains(l, "linux"):
		os = "Linux"
	}
	if os == "" {
		return browser
	}
	return browser + " on " + os
}

func deviceTypeFromUserAgent(ua string) string {
	l := strings.ToLower(ua)
	switch {
	case strings.Contains(l, "ipad") || strings.Contains(l, "tablet"):
		return "tablet"
	case strings.Contains(l, "mobi") || strings.Contains(l, "android"):
		return "mobile"
	default:
		return "desktop"
	}
}

// DeleteExpiredDevices deletes all expired trusted devices for a user
func (r *TrustedDeviceRepository) DeleteExpiredDevices(userID uuid.UUID) error {
	return r.GetDB().Where("user_id = ? AND expires_at < ?", userID, time.Now()).Delete(&models.TrustedDevice{}).Error
}

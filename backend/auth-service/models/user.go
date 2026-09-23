package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeID string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"employee_id"`
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	PasswordHash  string    `gorm:"type:varchar(255);not null" json:"-"`
	FullName  string    `gorm:"type:varchar(100)" json:"full_name"`
	Phone     string    `gorm:"type:varchar(20)" json:"phone"`
	Department string  `gorm:"type:varchar(100)" json:"department"`
	Position   string  `gorm:"type:varchar(100)" json:"position"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	Roles     []Role    `gorm:"many2many:user_roles;" json:"roles"`
	LockedUntil     *time.Time     `json:"locked_until,omitempty"` 
	LastLoginFingerprint string    `gorm:"type:varchar(255)" json:"last_login_fingerprint"`
	LastLoginAt     *time.Time     `json:"last_login_at"`
	TrustedDevices []TrustedDevice `gorm:"foreignKey:UserID" json:"trusted_devices"`
	MFASetup *MFASetup `gorm:"foreignKey:UserID" json:"mfa_setup"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type TrustedDevice struct {
    ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    UserID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
    User              User       `gorm:"foreignKey:UserID" json:"-"`
    DeviceFingerprint string     `gorm:"type:varchar(255);not null" json:"device_fingerprint"`
	DeviceName        string     `gorm:"type:varchar(255)" json:"device_name"`        // e.g. "Chrome on MacOS"
    DeviceType        string     `gorm:"type:varchar(50)" json:"device_type"`
	UserAgent 		  string     `gorm:"type:varchar(255)" json:"user_agent"`
    IPAddress         string     `gorm:"type:varchar(45)" json:"ip_address"`
    ExpiresAt         *time.Time `json:"expires_at,omitempty"`
    CreatedAt         time.Time  `json:"created_at"`
}

type MFAType string

const (
	MFATypeTOTP MFAType = "TOTP"
	MFATypeEmail MFAType = "Email"
)

type MFASetup struct {
	ID 	uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID 	uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	User 	*User       `gorm:"foreignKey:UserID" json:"-"`
	MFAType 	MFAType  `gorm:"type:varchar(20);not null" json:"mfa_type"`
	SecretKey 	string 	`gorm:"type:varchar(255);not null" json:"-"`
	IsEnabled bool 	`gorm:"default:false" json:"is_enabled"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Role represents a user role
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a permission
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Resource    string    `gorm:"type:varchar(50);not null" json:"resource"`
	Action      string    `gorm:"type:varchar(50);not null" json:"action"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`                                                                                                                                                                                          
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`                                                                                                                                                                                       
    Token     string    `gorm:"type:varchar(255);uniqueIndex;not null"`                                                                                                                                                                         
    ExpiresAt time.Time                                                                                                                                                                                                                         
    RevokedAt *time.Time                                                                                                                                                                                                                        
    CreatedAt time.Time 
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// TableName specifies the table name for TrustedDevice
func (TrustedDevice) TableName() string {
	return "trusted_devices"
}

// TableName specifies the table name for MFASetup
func (MFASetup) TableName() string {
	return "mfa_setups"
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// TableName specifies the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}

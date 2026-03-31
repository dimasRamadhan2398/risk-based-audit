package models

import (
	"time"

	"github.com/google/uuid"
)

// SysAlertAction represents the type of system alert
type SysAlertAction string

const (
	SysAlertLoginLog 	SysAlertAction = "sys_alert_login_log"
	SysAlertLogoutLog 	SysAlertAction = "sys_alert_logout_log"
	SysAlertFailedAttemptLog  SysAlertAction = "sys_alert_failed_attempt_log"
	SysAlertChangePermissionLog SysAlertAction = "sys_alert_change_permission_log"
	SysAlertLockedOut SysAlertAction = "sys_alert_locked_out"
	SysRoleAssigned SysAlertAction = "sys_alert_role_assigned_log"
	SysRoleRevoked SysAlertAction = "sys_alert_role_revoked_log"
	SysRateLimitExceeded SysAlertAction = "sys_alert_rate_limit_exceeded_log"
	SysPasswordResetRequested SysAlertAction = "sys_alert_password_reset_requested_log"
	SysPasswordResetCompleted SysAlertAction = "sys_alert_password_reset_completed_log"	
)

type SysAlertLog struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID *uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action SysAlertAction `gorm:"type:varchar(100);not null" json:"action"`
	IPAddress string `gorm:"type:varchar(45)" json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for SysAlertLog
func (SysAlertLog) TableName() string {
	return "sys_alert_logs"
}
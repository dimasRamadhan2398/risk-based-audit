package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// SysAlertAction represents the type of system alert
type SysAlertAction string

// Legacy action names. Nothing writes them; the activity log uses dotted verbs
// (pkg/activitylog). Kept so old rows, if any, still have a named type.
const (
	SysAlertLoginLog            SysAlertAction = "sys_alert_login_log"
	SysAlertLogoutLog           SysAlertAction = "sys_alert_logout_log"
	SysAlertFailedAttemptLog    SysAlertAction = "sys_alert_failed_attempt_log"
	SysAlertChangePermissionLog SysAlertAction = "sys_alert_change_permission_log"
	SysAlertLockedOut           SysAlertAction = "sys_alert_locked_out"
	SysRoleAssigned             SysAlertAction = "sys_alert_role_assigned_log"
	SysRoleRevoked              SysAlertAction = "sys_alert_role_revoked_log"
	SysRateLimitExceeded        SysAlertAction = "sys_alert_rate_limit_exceeded_log"
	SysPasswordResetRequested   SysAlertAction = "sys_alert_password_reset_requested_log"
	SysPasswordResetCompleted   SysAlertAction = "sys_alert_password_reset_completed_log"
)

// SysAlertLog is one row of the admin user activity log.
//
// The table predates the activity log; the columns below were added
// additively. user_id is the legacy column: it is now nullable and no longer
// written (actor_user_id replaces it without a foreign key, so a log row never
// blocks deleting a user and a failed login for an unknown username can be
// recorded). An existing fk_sys_alert_logs_user constraint, if present, stays
// and is harmless because user_id stays NULL.
type SysAlertLog struct {
	ID     uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid();index:idx_sys_alert_logs_occurred,priority:2,sort:desc;index:idx_sys_alert_logs_created,priority:2" json:"id"`
	UserID *uuid.UUID `gorm:"type:uuid;index" json:"-"`

	// EventID is the producer-assigned id; the unique index makes the Kafka
	// consumer and the direct-write fallback idempotent.
	EventID string `gorm:"type:varchar(64);uniqueIndex:idx_sys_alert_logs_event_id" json:"-"`
	// OccurredAt is when the action happened (producer clock). The cursor
	// index covers (occurred_at DESC, id DESC).
	OccurredAt    time.Time      `gorm:"type:timestamptz;not null;default:now();index:idx_sys_alert_logs_occurred,priority:1,sort:desc;index:idx_sys_alert_logs_actor_occurred,priority:2" json:"occurred_at"`
	ActorUserID   *uuid.UUID     `gorm:"type:uuid;index:idx_sys_alert_logs_actor_occurred,priority:1" json:"actor_user_id"`
	ActorUsername string         `gorm:"type:varchar(255)" json:"actor_username"`
	Action        SysAlertAction `gorm:"type:varchar(100);not null;index:idx_sys_alert_logs_action" json:"action"`
	EntityType    string         `gorm:"type:varchar(64)" json:"entity_type"`
	EntityID      string         `gorm:"type:varchar(128)" json:"entity_id"`
	Result        string         `gorm:"type:varchar(16);not null;default:'success'" json:"result"`
	IPAddress     string         `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent     string         `gorm:"type:varchar(512)" json:"user_agent"`
	Summary       string         `gorm:"type:varchar(500)" json:"summary"`
	Details       JSONMap        `gorm:"type:jsonb" json:"details"`
	// CreatedAt is when this row was stored (consumer clock). Polling
	// (since_cursor) walks this column so late Kafka deliveries are not missed.
	CreatedAt time.Time `gorm:"index:idx_sys_alert_logs_created,priority:1" json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// TableName specifies the table name for SysAlertLog
func (SysAlertLog) TableName() string {
	return "sys_alert_logs"
}

// JSONMap is a JSON object stored in a jsonb column. It is written as text so
// it works the same over the simple and the extended postgres protocol.
type JSONMap map[string]interface{}

// Value implements driver.Valuer.
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(map[string]interface{}(m))
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implements sql.Scanner.
func (m *JSONMap) Scan(value interface{}) error {
	var raw []byte
	switch v := value.(type) {
	case nil:
		*m = nil
		return nil
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return errors.New("JSONMap: unsupported scan type")
	}
	if len(raw) == 0 {
		*m = nil
		return nil
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return err
	}
	*m = out
	return nil
}

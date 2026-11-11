// Package activitylog records the admin user activity log (sys_alert_logs).
//
// Flow: a write point calls Recorder.Record. The recorder normalises and
// redacts the entry, publishes it to this stack's activity topic, and an
// auth-service consumer stores it idempotently on event_id. When the topic is
// unavailable (Kafka disabled, broker down, publish error) the recorder writes
// the row directly, so an event is not lost.
//
// Wire contract on the topic ([<KAFKA_TOPIC_PREFIX>.]activity-logs):
//
//	key    event_id
//	value  Entry as JSON, "version":1
//
// Never put passwords, tokens, MFA secrets/codes or temporary passwords in an
// Entry. Details are passed through Redact anyway as a second line of defence.
package activitylog

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// EnvelopeVersion is bumped only for a breaking change of Entry.
const EnvelopeVersion = 1

// Result values.
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
)

// Actions (dotted verbs) written to sys_alert_logs.action.
const (
	ActionLogin                 = "user.login"
	ActionLogout                = "user.logout"
	ActionLoginFailed           = "user.login_failed"
	ActionPasswordChanged       = "user.password_changed"
	ActionPasswordResetByAdmin  = "user.password_reset_by_admin"
	ActionMFAEnabled            = "mfa.enabled"
	ActionMFADisabled           = "mfa.disabled"
	ActionMFAVerified           = "mfa.verified"
	ActionMFAFailed             = "mfa.failed"
	ActionRoleCreated           = "role.created"
	ActionRoleDeleted           = "role.deleted"
	ActionRolePermissionsAdded  = "role.permissions_assigned"
	ActionRolePermissionsRemove = "role.permissions_revoked"
	ActionRoleAssigned          = "role.assigned"
	ActionRoleRevoked           = "role.revoked"
	ActionDeviceTrustedAdded    = "device.trusted_added"
	ActionDeviceTrustedRemoved  = "device.trusted_removed"
	ActionPermissionDenied      = "permission.denied"
	ActionStepUp                = "security.step_up"
	ActionSystemLogsViewed      = "system_logs.viewed"
)

// Column limits of sys_alert_logs (models.SysAlertLog).
const (
	maxEventID    = 64
	maxUsername   = 255
	maxAction     = 100
	maxEntityType = 64
	maxEntityID   = 128
	maxIP         = 45
	maxUserAgent  = 512
	maxSummary    = 500
)

// Entry is one activity event, both the Kafka message and the input of the
// direct-write fallback.
type Entry struct {
	Version       int                    `json:"version"`
	EventID       string                 `json:"event_id"`
	Source        string                 `json:"source"`
	OccurredAt    time.Time              `json:"occurred_at"`
	ActorUserID   string                 `json:"actor_user_id,omitempty"`
	ActorUsername string                 `json:"actor_username,omitempty"`
	Action        string                 `json:"action"`
	EntityType    string                 `json:"entity_type,omitempty"`
	EntityID      string                 `json:"entity_id,omitempty"`
	Result        string                 `json:"result"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Summary       string                 `json:"summary,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// Normalize fills defaults, bounds every field to its column and redacts
// Details. It is applied by the recorder before publishing and again by the
// consumer before storing, so a message from an older producer is still safe.
func (e *Entry) Normalize(now time.Time) {
	if e.Version == 0 {
		e.Version = EnvelopeVersion
	}
	e.EventID = bound(strings.TrimSpace(e.EventID), maxEventID)
	if e.EventID == "" {
		e.EventID = uuid.NewString()
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = now
	}
	// Postgres keeps microseconds; truncating here keeps cursors exact.
	e.OccurredAt = e.OccurredAt.UTC().Truncate(time.Microsecond)
	e.ActorUserID = strings.TrimSpace(e.ActorUserID)
	if _, err := uuid.Parse(e.ActorUserID); err != nil {
		e.ActorUserID = ""
	}
	e.ActorUsername = bound(strings.TrimSpace(e.ActorUsername), maxUsername)
	e.Action = bound(strings.ToLower(strings.TrimSpace(e.Action)), maxAction)
	e.EntityType = bound(strings.TrimSpace(e.EntityType), maxEntityType)
	e.EntityID = bound(strings.TrimSpace(e.EntityID), maxEntityID)
	if e.Result != ResultFailure {
		e.Result = ResultSuccess
	}
	e.IPAddress = NormalizeIP(e.IPAddress)
	e.UserAgent = bound(strings.TrimSpace(e.UserAgent), maxUserAgent)
	e.Summary = bound(strings.TrimSpace(e.Summary), maxSummary)
	e.Details = Redact(e.Details)
}

// NormalizeIP keeps the first address of an X-Forwarded-For style list and
// bounds it to the column.
func NormalizeIP(ip string) string {
	if i := strings.IndexByte(ip, ','); i >= 0 {
		ip = ip[:i]
	}
	return bound(strings.TrimSpace(ip), maxIP)
}

// bound cuts s to at most n bytes on a rune boundary.
func bound(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

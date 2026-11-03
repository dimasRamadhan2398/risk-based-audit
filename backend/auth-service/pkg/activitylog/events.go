package activitylog

import (
	"context"
	"fmt"

	"auth-service/pkg/kafka"
)

// OperationKey in kafka.Event.Data, when set, overrides the activity action
// derived from the event type (role service: role.created, role.deleted, ...).
const OperationKey = "operation"

// Tap returns a kafka.EventTap that records every published auth event. It is
// how the already-live publishers (MFA, roles, trusted devices,
// permission.denied, step-up) reach the activity log.
func Tap(rec Recorder) kafka.EventTap {
	return func(ctx context.Context, ev *kafka.Event) {
		if rec == nil || ev == nil {
			return
		}
		if e, ok := FromKafkaEvent(ctx, ev); ok {
			rec.Record(ctx, e)
		}
	}
}

// FromKafkaEvent maps an auth event to an activity entry. The actor is the
// authenticated caller in ctx when there is one (an admin changing a role),
// otherwise the event's user (the user enabling their own MFA). ok is false
// for event types the activity log does not record.
func FromKafkaEvent(ctx context.Context, ev *kafka.Event) (Entry, bool) {
	data := copyData(ev.Data)
	e := Entry{
		OccurredAt: ev.Timestamp,
		IPAddress:  ev.IPAddress,
		UserAgent:  ev.UserAgent,
		Result:     ResultSuccess,
	}
	if !ev.Success {
		e.Result = ResultFailure
	}
	if a, ok := ActorFrom(ctx); ok {
		e.ActorUserID, e.ActorUsername = a.UserID, a.Username
	} else {
		e.ActorUserID, e.ActorUsername = ev.UserID, ev.Username
	}
	if ev.ErrorReason != "" {
		data["reason"] = ev.ErrorReason
	}

	who := e.ActorUsername
	if who == "" {
		who = "unknown user"
	}
	role, _ := data["role"].(string)
	op, _ := data[OperationKey].(string)
	delete(data, OperationKey)

	switch ev.Type {
	case kafka.EventUserLogin:
		e.Action, e.EntityType, e.EntityID = ActionLogin, "user", ev.UserID
		e.Summary = fmt.Sprintf("%s signed in", who)
	case kafka.EventUserLogout:
		e.Action, e.EntityType, e.EntityID = ActionLogout, "user", ev.UserID
		e.Summary = fmt.Sprintf("%s signed out", who)
	case kafka.EventUserLoginFailed:
		e.Action, e.EntityType, e.EntityID = ActionLoginFailed, "user", ev.UserID
		e.Result = ResultFailure
		e.Summary = fmt.Sprintf("Failed sign-in for %s", who)
	case kafka.EventPasswordChanged:
		e.Action, e.EntityType, e.EntityID = ActionPasswordChanged, "user", ev.UserID
		e.Summary = fmt.Sprintf("%s changed their password", who)
	case kafka.EventMFAEnabled:
		e.Action, e.EntityType, e.EntityID = ActionMFAEnabled, "user", ev.UserID
		e.Summary = fmt.Sprintf("%s enabled MFA", who)
	case kafka.EventMFADisabled:
		e.Action, e.EntityType, e.EntityID = ActionMFADisabled, "user", ev.UserID
		e.Summary = fmt.Sprintf("%s disabled MFA", who)
	case kafka.EventMFAVerified:
		e.Action, e.EntityType, e.EntityID = ActionMFAVerified, "user", ev.UserID
		e.Summary = fmt.Sprintf("%s verified an MFA code", who)
	case kafka.EventMFAFailed:
		e.Action, e.EntityType, e.EntityID = ActionMFAFailed, "user", ev.UserID
		e.Result = ResultFailure
		e.Summary = fmt.Sprintf("Invalid MFA code for %s", who)
	case kafka.EventRoleAssigned, kafka.EventRoleRevoked:
		e.EntityType, e.EntityID = "role", role
		switch op {
		case ActionRoleCreated:
			e.Action, e.Summary = ActionRoleCreated, fmt.Sprintf("%s created role %s", who, role)
		case ActionRoleDeleted:
			e.Action, e.Summary = ActionRoleDeleted, fmt.Sprintf("%s deleted role %s", who, role)
		case ActionRolePermissionsAdded:
			e.Action, e.Summary = ActionRolePermissionsAdded, fmt.Sprintf("%s assigned permissions to role %s", who, role)
		case ActionRolePermissionsRemove:
			e.Action, e.Summary = ActionRolePermissionsRemove, fmt.Sprintf("%s revoked permissions from role %s", who, role)
		default:
			if ev.Type == kafka.EventRoleAssigned {
				e.Action, e.Summary = ActionRoleAssigned, fmt.Sprintf("%s assigned role %s", who, role)
			} else {
				e.Action, e.Summary = ActionRoleRevoked, fmt.Sprintf("%s revoked role %s", who, role)
			}
		}
	case kafka.EventTrustedDeviceAdded:
		e.Action, e.EntityType = ActionDeviceTrustedAdded, "trusted_device"
		e.EntityID, _ = data["device_id"].(string)
		e.Summary = fmt.Sprintf("%s added a trusted device", who)
	case kafka.EventTrustedDeviceRemoved:
		e.Action, e.EntityType = ActionDeviceTrustedRemoved, "trusted_device"
		e.EntityID, _ = data["device_id"].(string)
		e.Summary = fmt.Sprintf("%s removed a trusted device", who)
	case kafka.EventPermissionDenied:
		e.Action, e.Result = ActionPermissionDenied, ResultFailure
		if role != "" {
			e.EntityType, e.EntityID = "role", role
		}
		e.Summary = fmt.Sprintf("Permission denied for %s", who)
	case kafka.EventStepUp:
		e.Action, e.EntityType, e.EntityID = ActionStepUp, "user", ev.UserID
		purpose, _ := data["purpose"].(string)
		if e.Result == ResultSuccess {
			e.Summary = fmt.Sprintf("%s confirmed their password (%s)", who, purpose)
		} else {
			e.Summary = fmt.Sprintf("Password confirmation rejected for %s (%s)", who, purpose)
		}
	default:
		return Entry{}, false
	}
	if len(data) > 0 {
		e.Details = data
	}
	return e, true
}

func copyData(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in)+1)
	for k, v := range in {
		out[k] = v
	}
	return out
}

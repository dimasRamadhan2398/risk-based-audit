package models

import "time"

// ActivityCodeSequence holds the last sequence number handed out for an
// Activity ID prefix ("ASR-2026"). The counter only goes up, so a deleted
// activity's number is never reused. Both AuditActivity.ProjectCode and
// PlannedActivity.ActivityCode draw from it, so an Activity ID names exactly
// one activity. See pkg/activitycode.
type ActivityCodeSequence struct {
	Prefix    string    `gorm:"type:varchar(30);primaryKey" json:"prefix"`
	LastValue int       `gorm:"type:int;not null;default:0" json:"last_value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ActivityCodeSequence) TableName() string {
	return "activity_code_sequences"
}

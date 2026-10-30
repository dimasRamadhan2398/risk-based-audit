package models

import (
	"time"

	"github.com/google/uuid"
)

// OrgChangeLogActionApply is the action recorded for an applied level change.
const OrgChangeLogActionApply = "org_levels.apply"

// OrgChangeLog is the append-only audit log of org level changes
// (docs/org-structure-model-plan.md 9.4). Created by the gated DDL only, NOT
// auto-migrated. Before, After, Impact and ActorRoles hold JSON documents.
type OrgChangeLog struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	Action        string    `gorm:"column:action" json:"action"`
	ActorUserID   string    `gorm:"column:actor_user_id" json:"actor_user_id"`
	ActorUsername string    `gorm:"column:actor_username" json:"actor_username"`
	ActorRoles    JSONText  `gorm:"column:actor_roles" json:"actor_roles"`
	IP            string    `gorm:"column:ip" json:"ip"`
	UserAgent     string    `gorm:"column:user_agent" json:"user_agent"`
	BaseVersion   string    `gorm:"column:base_version" json:"base_version"`
	Before        JSONText  `gorm:"column:before" json:"before"`
	After         JSONText  `gorm:"column:after" json:"after"`
	Impact        JSONText  `gorm:"column:impact" json:"impact"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (OrgChangeLog) TableName() string { return "org_change_log" }

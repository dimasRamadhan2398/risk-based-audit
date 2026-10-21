package models

import "time"

// Kinds of AnnualSnapshotRef rows.
const (
	// SnapshotRefDepartment: RefID is a master-service department id; Name is its name.
	SnapshotRefDepartment = "department"
	// SnapshotRefEntity: RefID is a corporate Audit Universe id; Name is its name.
	SnapshotRefEntity = "entity"
	// SnapshotRefYearScore: RefID is a corporate Audit Universe id and Year the
	// scored year; RiskLevel is the entity's level for that year.
	SnapshotRefYearScore = "year_score"
)

// AnnualSnapshotRef is audit-service's local copy of the names and levels that
// the PKAT display snapshots (AnnualAuditActivity.Department, RiskName and
// RiskLevel) are filled from. It is fed by the Department and Audit Universe
// events that services/annual_plan_sync already consumes, so it holds the same
// values as master-service and risk-service without a call to either.
//
// It is a cache, not a source of truth: a missing row means "no value known
// yet" and never fails a save. Rows are only written (upserted) by the event
// handler and never deleted, so a deleted source keeps the last name it had,
// like the plan snapshots do.
//
// Links are by business key (RefID is the other service's id), not by foreign
// key, because the source tables live in other databases.
type AnnualSnapshotRef struct {
	// Kind, RefID and Year form one unique key (the upsert target). It is a
	// unique index, not a primary key, so the table needs no surrogate id.
	Kind      string    `gorm:"type:varchar(20);not null;uniqueIndex:idx_annual_snapshot_refs_key,priority:1" json:"kind"`
	RefID     string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_annual_snapshot_refs_key,priority:2" json:"ref_id"`
	Year      int       `gorm:"type:int;not null;uniqueIndex:idx_annual_snapshot_refs_key,priority:3" json:"year"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	RiskLevel string    `gorm:"type:varchar(50)" json:"risk_level"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AnnualSnapshotRef) TableName() string {
	return "annual_snapshot_refs"
}

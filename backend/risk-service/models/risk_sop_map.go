package models

import (
	"time"

	"github.com/google/uuid"
)

type RiskSOPMap struct {
	RiskID    uuid.UUID    `gorm:"type:uuid;primaryKey" json:"risk_id"`
	Risk      RiskRegister `gorm:"foreignKey:RiskID" json:"risk"`
	SOPID     uuid.UUID    `gorm:"type:uuid;primaryKey" json:"sop_id"`
	SOP       SOPDocument  `gorm:"foreignKey:SOPID" json:"sop"`
	CreatedAt time.Time    `json:"created_at"`
}

func (RiskSOPMap) TableName() string {
	return "risk_sop_map"
}

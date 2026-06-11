package models

import (
	"time"

	"github.com/google/uuid"
)

type RiskStrategyMap struct {
	RiskID     uuid.UUID          `gorm:"type:uuid;primaryKey" json:"risk_id"`
	Risk       RiskRegister       `gorm:"foreignKey:RiskID" json:"risk"`
	StrategyID uuid.UUID          `gorm:"type:uuid;primaryKey" json:"strategy_id"`
	Strategy   StrategicObjective `gorm:"foreignKey:StrategyID" json:"strategy"`
	CreatedAt  time.Time          `json:"created_at"`
}

func (RiskStrategyMap) TableName() string {
	return "risk_strategy_map"
}

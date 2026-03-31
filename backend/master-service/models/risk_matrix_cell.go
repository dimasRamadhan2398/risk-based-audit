package master

import (
	"time"

	"github.com/google/uuid"
)

type RiskMatrixCell struct {
    ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

    LikelihoodID uuid.UUID  `gorm:"type:uuid;not null;index" json:"likelihood_id"`
    Likelihood   Likelihood `gorm:"foreignKey:LikelihoodID" json:"likelihood"`

    ImpactID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"impact_id"`
    Impact       Impact     `gorm:"foreignKey:ImpactID" json:"impact"`

    Score        int        `gorm:"not null" json:"score"`           // e.g. 17

    RiskLevelID  uuid.UUID  `gorm:"type:uuid;not null" json:"risk_level_id"`
    RiskLevel    RiskLevel  `gorm:"foreignKey:RiskLevelID" json:"risk_level"`

    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}
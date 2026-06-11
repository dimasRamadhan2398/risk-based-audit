package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BSCPerspective string

const (
	BSCPerspectiveFinance  BSCPerspective = "FINANCE"
	BSCPerspectiveCustomer BSCPerspective = "CUSTOMER"
	BSCPerspectiveInternal BSCPerspective = "INTERNAL"
	BSCPerspectiveLearning BSCPerspective = "LEARNING"
)

type StrategicObjective struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BSCPerspective BSCPerspective `gorm:"type:varchar(20);not null" json:"bsc_perspective"`
	ObjectiveName  string         `gorm:"type:varchar(200);not null" json:"objective_name"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (StrategicObjective) TableName() string {
	return "strategic_objective"
}

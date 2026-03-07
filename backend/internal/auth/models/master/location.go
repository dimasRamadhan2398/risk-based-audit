package master

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Location struct {
    ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    Name        string         `gorm:"type:varchar(100);not null" json:"name"`         // e.g. "HQ Jakarta", "Surabaya Branch"
    Address     string         `gorm:"type:text;not null" json:"address"`              // Street address
    City        string         `gorm:"type:varchar(100);not null" json:"city"`
    Province    string         `gorm:"type:varchar(100)" json:"province"`
    PostalCode  string         `gorm:"type:varchar(20)" json:"postal_code"`
    Country     string         `gorm:"type:varchar(100);default:'Indonesia'" json:"country"`
    Latitude    *float64       `gorm:"type:decimal(10,8)" json:"latitude,omitempty"`   // optional GPS
    Longitude   *float64       `gorm:"type:decimal(11,8)" json:"longitude,omitempty"`  // optional GPS
    IsActive    bool           `gorm:"default:true" json:"is_active"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
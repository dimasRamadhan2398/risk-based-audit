package master

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobRole struct {
	ID 		uuid.UUID 		`gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	JobRoleCode 	string 	`gorm:"type:varchar(50);uniqueIndex;not null" json:"job_role_code"`
	JobRoleName 	string 	`gorm:"type:varchar(100);uniqueIndex;not null" json:"job_role_name"`
	JobRoleDescription string 	`gorm:"type:text" json:"job_role_description"`
	IsActive 	bool 		`gorm:"default:true" json:"is_active"`
	CreatedAt time.Time 	`json:"created_at"`
	UpdatedAt time.Time 	`json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
type RiskLevel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    RiskCode        string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"`    // "LOW", "MEDIUM", "HIGH", "CRITICAL"
    RiskName        string         `gorm:"type:varchar(50);not null" json:"name"`
    RiskDescription string         `gorm:"type:text" json:"description"`
    Level       int            `gorm:"not null" json:"level"`                                // 1=Low, 2=Medium, 3=High, 4=Critical (for sorting/comparison)
    Color       string         `gorm:"type:varchar(20)" json:"color"`                        // UI hint: "green", "yellow", "red"
    IsActive    bool           `gorm:"default:true" json:"is_active"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}
type RiskRegister struct {
    ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    Code            string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`   // "RR-2024-001"
    Name            string         `gorm:"type:varchar(200);not null" json:"name"`              // risk event name
    Description     string         `gorm:"type:text" json:"description"`

    // Ownership
    DepartmentID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"department_id"`
    Department      Department     `gorm:"foreignKey:DepartmentID" json:"department"`

    RiskCategoryID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"risk_category_id"`
    RiskCategory    RiskCategory   `gorm:"foreignKey:RiskCategoryID" json:"risk_category"`

    RiskOwnerID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"risk_owner_id"`
    RiskOwner       Employee       `gorm:"foreignKey:RiskOwnerID" json:"risk_owner"`

    // ── INHERENT RISK (before controls) ──
    InherentLikelihoodID uuid.UUID      `gorm:"type:uuid;not null" json:"inherent_likelihood_id"`
    InherentLikelihood   Likelihood     `gorm:"foreignKey:InherentLikelihoodID" json:"inherent_likelihood"`

    InherentImpactID     uuid.UUID      `gorm:"type:uuid;not null" json:"inherent_impact_id"`
    InherentImpact       Impact         `gorm:"foreignKey:InherentImpactID" json:"inherent_impact"`

    InherentScore        int            `gorm:"not null" json:"inherent_score"`        // from matrix cell
    InherentRiskLevelID  uuid.UUID      `gorm:"type:uuid;not null" json:"inherent_risk_level_id"`
    InherentRiskLevel    RiskLevel      `gorm:"foreignKey:InherentRiskLevelID" json:"inherent_risk_level"`

    // ── RESIDUAL RISK (after controls applied) ──
    ResidualLikelihoodID uuid.UUID      `gorm:"type:uuid;not null" json:"residual_likelihood_id"`
    ResidualLikelihood   Likelihood     `gorm:"foreignKey:ResidualLikelihoodID" json:"residual_likelihood"`

    ResidualImpactID     uuid.UUID      `gorm:"type:uuid;not null" json:"residual_impact_id"`
    ResidualImpact       Impact         `gorm:"foreignKey:ResidualImpactID" json:"residual_impact"`

    ResidualScore        int            `gorm:"not null" json:"residual_score"`        // from matrix cell
    ResidualRiskLevelID  uuid.UUID      `gorm:"type:uuid;not null" json:"residual_risk_level_id"`
    ResidualRiskLevel    RiskLevel      `gorm:"foreignKey:ResidualRiskLevelID" json:"residual_risk_level"`

    AssessmentDate  time.Time      `json:"assessment_date"`
    NextReviewDate  time.Time      `json:"next_review_date"`
    IsActive        bool           `gorm:"default:true" json:"is_active"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
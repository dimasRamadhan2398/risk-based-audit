package repositories

import (
	"risk-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IMitigationRepository interface {
	FindAll(riskID uuid.UUID) ([]models.RiskMitigation, error)
	FindByID(id uuid.UUID) (*models.RiskMitigation, error)
	Create(mitigation *models.RiskMitigation) error
	Save(mitigation *models.RiskMitigation) error
	Delete(mitigation *models.RiskMitigation) error
}

type mitigationRepository struct {
	db *gorm.DB
}

func NewMitigationRepository(db *gorm.DB) IMitigationRepository {
	return &mitigationRepository{db: db}
}

func (r *mitigationRepository) FindAll(riskID uuid.UUID) ([]models.RiskMitigation, error) {
	var mitigations []models.RiskMitigation
	query := r.db
	if riskID != uuid.Nil {
		query = query.Where("risk_id = ?", riskID)
	}
	err := query.Find(&mitigations).Error
	return mitigations, err
}

func (r *mitigationRepository) FindByID(id uuid.UUID) (*models.RiskMitigation, error) {
	var mitigation models.RiskMitigation
	err := r.db.First(&mitigation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &mitigation, nil
}

func (r *mitigationRepository) Create(mitigation *models.RiskMitigation) error {
	return r.db.Create(mitigation).Error
}

func (r *mitigationRepository) Save(mitigation *models.RiskMitigation) error {
	return r.db.Save(mitigation).Error
}

func (r *mitigationRepository) Delete(mitigation *models.RiskMitigation) error {
	return r.db.Delete(mitigation).Error
}

package repositories

import (
	"risk-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IRiskRepository interface {
	FindAll() ([]models.RiskRegister, error)
	FindByID(id uuid.UUID) (*models.RiskRegister, error)
	CreateProfile(profile *models.RiskProfile) error
	CreateRegister(register *models.RiskRegister) error
	CreateAssessment(assessment *models.RiskAssessment) error
	SaveRegister(register *models.RiskRegister) error
	SaveProfile(profile *models.RiskProfile) error
	SaveAssessment(assessment *models.RiskAssessment) error
	DeleteRegister(register *models.RiskRegister) error
	DeleteProfile(id uuid.UUID) error
	FindAssessmentByYear(regID uuid.UUID, year int) (*models.RiskAssessment, error)
	FindAssessmentsByRegisterID(regID uuid.UUID) ([]models.RiskAssessment, error)
}

type riskRepository struct {
	db *gorm.DB
}

func NewRiskRepository(db *gorm.DB) IRiskRepository {
	return &riskRepository{db: db}
}

func (r *riskRepository) FindAll() ([]models.RiskRegister, error) {
	var registers []models.RiskRegister
	err := r.db.Preload("Profile").Preload("Assessments").Find(&registers).Error
	return registers, err
}

func (r *riskRepository) FindByID(id uuid.UUID) (*models.RiskRegister, error) {
	var register models.RiskRegister
	err := r.db.Preload("Profile").First(&register, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &register, nil
}

func (r *riskRepository) CreateProfile(profile *models.RiskProfile) error {
	return r.db.Create(profile).Error
}

func (r *riskRepository) CreateRegister(register *models.RiskRegister) error {
	return r.db.Create(register).Error
}

func (r *riskRepository) CreateAssessment(assessment *models.RiskAssessment) error {
	return r.db.Create(assessment).Error
}

func (r *riskRepository) SaveRegister(register *models.RiskRegister) error {
	return r.db.Save(register).Error
}

func (r *riskRepository) SaveProfile(profile *models.RiskProfile) error {
	return r.db.Save(profile).Error
}

func (r *riskRepository) SaveAssessment(assessment *models.RiskAssessment) error {
	return r.db.Save(assessment).Error
}

func (r *riskRepository) DeleteRegister(register *models.RiskRegister) error {
	return r.db.Delete(register).Error
}

func (r *riskRepository) DeleteProfile(id uuid.UUID) error {
	return r.db.Delete(&models.RiskProfile{}, "id = ?", id).Error
}

func (r *riskRepository) FindAssessmentByYear(regID uuid.UUID, year int) (*models.RiskAssessment, error) {
	var assessment models.RiskAssessment
	err := r.db.First(&assessment, "risk_register_id = ? AND year = ?", regID, year).Error
	if err != nil {
		return nil, err
	}
	return &assessment, nil
}

func (r *riskRepository) FindAssessmentsByRegisterID(regID uuid.UUID) ([]models.RiskAssessment, error) {
	var assessments []models.RiskAssessment
	err := r.db.Find(&assessments, "risk_register_id = ?", regID).Error
	return assessments, err
}

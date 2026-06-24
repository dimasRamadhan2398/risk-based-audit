package risk_matrix_cell
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IRiskMatrixCellRepository interface {
	Create(cell *models.RiskMatrixCell) error
	Update(cell *models.RiskMatrixCell) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.RiskMatrixCell, error)
	FindAll() ([]*models.RiskMatrixCell, error)
	FindByLikelihoodAndImpact(likelihoodID, impactID uuid.UUID) (*models.RiskMatrixCell, error)
}
type RiskMatrixCellRepository struct { *repositories.BaseRepository }
func NewRiskMatrixCellRepository(db *gorm.DB) IRiskMatrixCellRepository { return &RiskMatrixCellRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *RiskMatrixCellRepository) Create(cell *models.RiskMatrixCell) error { return r.BaseRepository.Create(cell) }
func (r *RiskMatrixCellRepository) Update(cell *models.RiskMatrixCell) error { return r.BaseRepository.Update(cell) }
func (r *RiskMatrixCellRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.RiskMatrixCell{ID: id}) }
func (r *RiskMatrixCellRepository) FindByID(id uuid.UUID) (*models.RiskMatrixCell, error) {
	var cell models.RiskMatrixCell
	if err := r.GetDB().Preload("Likelihood").Preload("Impact").Preload("RiskLevel").First(&cell, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &cell, nil
}
func (r *RiskMatrixCellRepository) FindAll() ([]*models.RiskMatrixCell, error) {
	var cells []*models.RiskMatrixCell
	if err := r.GetDB().Preload("Likelihood").Preload("Impact").Preload("RiskLevel").Find(&cells).Error; err != nil { return nil, err }
	return cells, nil
}
func (r *RiskMatrixCellRepository) FindByLikelihoodAndImpact(likelihoodID, impactID uuid.UUID) (*models.RiskMatrixCell, error) {
	var cell models.RiskMatrixCell
	if err := r.GetDB().Where("likelihood_id = ? AND impact_id = ?", likelihoodID, impactID).First(&cell).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &cell, nil
}

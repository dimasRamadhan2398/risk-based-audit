package risk_matrix_cell
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/risk_matrix_cell"
	"github.com/google/uuid"
)
type RiskMatrixCellServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.RiskMatrixCell, error)
	FindById(ctx *base.BaseService, id string) (*models.RiskMatrixCell, error)
	Create(ctx *base.BaseService, cell *models.RiskMatrixCell) (*models.RiskMatrixCell, error)
	Update(ctx *base.BaseService, id string, cell *models.RiskMatrixCell) (*models.RiskMatrixCell, error)
	Delete(ctx *base.BaseService, id string) error
}
type RiskMatrixCellService struct { cellRepo repo.IRiskMatrixCellRepository }
func NewRiskMatrixCellService(cellRepo repo.IRiskMatrixCellRepository) RiskMatrixCellServiceInterface { return &RiskMatrixCellService{cellRepo: cellRepo} }
func (s *RiskMatrixCellService) Create(ctx *base.BaseService, cell *models.RiskMatrixCell) (*models.RiskMatrixCell, error) {
	if _, err := s.cellRepo.FindByLikelihoodAndImpact(cell.LikelihoodID, cell.ImpactID); err == nil {
		return nil, apperrors.Wrap("CELL_ALREADY_EXISTS", "Cell for this Likelihood and Impact already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate cell uniqueness", 500, err)
	}
	if err := s.cellRepo.Create(cell); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create risk matrix cell", 500, err) }
	return s.cellRepo.FindByID(cell.ID)
}
func (s *RiskMatrixCellService) Delete(ctx *base.BaseService, id string) error {
	cellID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_CELL_ID", "Invalid cell ID format", 400, err) }
	if _, err := s.cellRepo.FindByID(cellID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find cell", 500, err)
	}
	if err := s.cellRepo.Delete(cellID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete cell", 500, err) }
	return nil
}
func (s *RiskMatrixCellService) FindAll(ctx *base.BaseService) (*[]models.RiskMatrixCell, error) {
	cells, err := s.cellRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch cells", 500, err) }
	result := make([]models.RiskMatrixCell, 0, len(cells))
	for _, cell := range cells { if cell != nil { result = append(result, *cell) } }
	return &result, nil
}
func (s *RiskMatrixCellService) FindById(ctx *base.BaseService, id string) (*models.RiskMatrixCell, error) {
	cellID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_CELL_ID", "Invalid cell ID format", 400, err) }
	cell, err := s.cellRepo.FindByID(cellID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch cell", 500, err)
	}
	return cell, nil
}
func (s *RiskMatrixCellService) Update(ctx *base.BaseService, id string, cell *models.RiskMatrixCell) (*models.RiskMatrixCell, error) {
	cellID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_CELL_ID", "Invalid cell ID format", 400, err) }
	existingCell, err := s.cellRepo.FindByID(cellID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch cell", 500, err)
	}
	if existingCell.LikelihoodID != cell.LikelihoodID || existingCell.ImpactID != cell.ImpactID {
		if _, err := s.cellRepo.FindByLikelihoodAndImpact(cell.LikelihoodID, cell.ImpactID); err == nil {
			return nil, apperrors.Wrap("CELL_ALREADY_EXISTS", "Cell for this Likelihood and Impact already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate cell uniqueness", 500, err)
		}
	}
	existingCell.LikelihoodID = cell.LikelihoodID
	existingCell.ImpactID = cell.ImpactID
	existingCell.Score = cell.Score
	existingCell.RiskLevelID = cell.RiskLevelID
	if err := s.cellRepo.Update(existingCell); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update cell", 500, err) }
	return s.cellRepo.FindByID(cellID)
}

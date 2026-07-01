package risk_indicator
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/risk_indicator"
	"time"
	"github.com/google/uuid"
)
type RiskIndicatorServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.RiskIndicator, error)
	FindById(ctx *base.BaseService, id string) (*models.RiskIndicator, error)
	Create(ctx *base.BaseService, indicator *models.RiskIndicator) (*models.RiskIndicator, error)
	Update(ctx *base.BaseService, id string, indicator *models.RiskIndicator) (*models.RiskIndicator, error)
	Delete(ctx *base.BaseService, id string) error
	AddLog(ctx *base.BaseService, log *models.RiskIndicatorLog) (*models.RiskIndicatorLog, error)
	GetLogs(ctx *base.BaseService, indicatorID string) (*[]models.RiskIndicatorLog, error)
}
type RiskIndicatorService struct { indicatorRepo repo.IRiskIndicatorRepository }
func NewRiskIndicatorService(indicatorRepo repo.IRiskIndicatorRepository) RiskIndicatorServiceInterface { return &RiskIndicatorService{indicatorRepo: indicatorRepo} }
func (s *RiskIndicatorService) Create(ctx *base.BaseService, indicator *models.RiskIndicator) (*models.RiskIndicator, error) {
	if _, err := s.indicatorRepo.FindByCode(indicator.IndicatorCode); err == nil {
		return nil, apperrors.Wrap("INDICATOR_CODE_ALREADY_EXISTS", "Indicator code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate indicator code", 500, err)
	}
	if err := s.indicatorRepo.Create(indicator); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create risk indicator", 500, err) }
	return s.indicatorRepo.FindByID(indicator.ID)
}
func (s *RiskIndicatorService) Delete(ctx *base.BaseService, id string) error {
	indicatorID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_INDICATOR_ID", "Invalid indicator ID format", 400, err) }
	if _, err := s.indicatorRepo.FindByID(indicatorID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find indicator", 500, err)
	}
	if err := s.indicatorRepo.Delete(indicatorID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete indicator", 500, err) }
	return nil
}
func (s *RiskIndicatorService) FindAll(ctx *base.BaseService) (*[]models.RiskIndicator, error) {
	indicators, err := s.indicatorRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk indicators", 500, err) }
	result := make([]models.RiskIndicator, 0, len(indicators))
	for _, indicator := range indicators { if indicator != nil { result = append(result, *indicator) } }
	return &result, nil
}
func (s *RiskIndicatorService) FindById(ctx *base.BaseService, id string) (*models.RiskIndicator, error) {
	indicatorID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_INDICATOR_ID", "Invalid indicator ID format", 400, err) }
	indicator, err := s.indicatorRepo.FindByID(indicatorID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk indicator", 500, err)
	}
	return indicator, nil
}
func (s *RiskIndicatorService) Update(ctx *base.BaseService, id string, indicator *models.RiskIndicator) (*models.RiskIndicator, error) {
	indicatorID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_INDICATOR_ID", "Invalid indicator ID format", 400, err) }
	existingIndicator, err := s.indicatorRepo.FindByID(indicatorID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk indicator", 500, err)
	}
	if existingIndicator.IndicatorCode != indicator.IndicatorCode {
		if _, err := s.indicatorRepo.FindByCode(indicator.IndicatorCode); err == nil {
			return nil, apperrors.Wrap("INDICATOR_CODE_ALREADY_EXISTS", "Indicator code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate indicator code", 500, err)
		}
	}
	existingIndicator.IndicatorCode = indicator.IndicatorCode
	existingIndicator.IndicatorName = indicator.IndicatorName
	existingIndicator.Description = indicator.Description
	existingIndicator.RiskRegisterID = indicator.RiskRegisterID
	existingIndicator.Metric = indicator.Metric
	existingIndicator.Unit = indicator.Unit
	existingIndicator.Frequency = indicator.Frequency
	existingIndicator.ThresholdMin = indicator.ThresholdMin
	existingIndicator.ThresholdMax = indicator.ThresholdMax
	existingIndicator.ToleranceLevel = indicator.ToleranceLevel
	existingIndicator.CurrentValue = indicator.CurrentValue
	existingIndicator.Trend = indicator.Trend
	existingIndicator.TrendComment = indicator.TrendComment
	existingIndicator.Status = indicator.Status
	existingIndicator.DataSource = indicator.DataSource
	existingIndicator.DataSourceURL = indicator.DataSourceURL
	existingIndicator.OwnerID = indicator.OwnerID
	if err := s.indicatorRepo.Update(existingIndicator); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update risk indicator", 500, err) }
	return s.indicatorRepo.FindByID(indicatorID)
}
func (s *RiskIndicatorService) AddLog(ctx *base.BaseService, log *models.RiskIndicatorLog) (*models.RiskIndicatorLog, error) {
	indicator, err := s.indicatorRepo.FindByID(log.IndicatorID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, apperrors.Wrap("INDICATOR_NOT_FOUND", "Risk indicator not found", 404, nil) }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to find indicator", 500, err)
	}
	if log.RecordedAt.IsZero() { log.RecordedAt = time.Now() }
	if err := s.indicatorRepo.CreateLog(log); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create indicator log", 500, err) }
	indicator.CurrentValue = &log.Value
	now := time.Now(); indicator.LastUpdatedAt = &now
	_ = s.indicatorRepo.Update(indicator)
	return log, nil
}
func (s *RiskIndicatorService) GetLogs(ctx *base.BaseService, indicatorID string) (*[]models.RiskIndicatorLog, error) {
	id, err := uuid.Parse(indicatorID)
	if err != nil { return nil, apperrors.Wrap("INVALID_INDICATOR_ID", "Invalid indicator ID format", 400, err) }
	logs, err := s.indicatorRepo.FindLogsByIndicatorID(id)
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch indicator logs", 500, err) }
	result := make([]models.RiskIndicatorLog, 0, len(logs))
	for _, log := range logs { if log != nil { result = append(result, *log) } }
	return &result, nil
}

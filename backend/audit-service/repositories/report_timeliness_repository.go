package repositories

import (
	"audit-service/models"

	"gorm.io/gorm"
)

type IReportTimelinessRepository interface {
	Create(entity *models.ReportTimeliness) error
	FindAll() ([]models.ReportTimeliness, error)
	FindByYearAndPeriod(year int, period string) (*models.ReportTimeliness, error)
}

type reportTimelinessRepository struct {
	BaseRepository
}

func NewReportTimelinessRepository(db *gorm.DB) IReportTimelinessRepository {
	return &reportTimelinessRepository{
		BaseRepository: BaseRepository{DB: db},
	}
}

func (r *reportTimelinessRepository) Create(entity *models.ReportTimeliness) error {
	return r.BaseRepository.Create(entity)
}

func (r *reportTimelinessRepository) FindAll() ([]models.ReportTimeliness, error) {
	var list []models.ReportTimeliness
	if err := r.DB.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *reportTimelinessRepository) FindByYearAndPeriod(year int, period string) (*models.ReportTimeliness, error) {
	var entity models.ReportTimeliness
	if err := r.DB.Where("year = ? AND period = ?", year, period).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

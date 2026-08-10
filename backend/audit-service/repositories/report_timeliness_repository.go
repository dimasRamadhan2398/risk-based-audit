package repositories

import (
	"audit-service/models"

	"gorm.io/gorm"
)

type IReportTimelinessRepository interface {
	Create(entity *models.ReportTimeliness) error
	FindAll() ([]models.ReportTimeliness, error)
	FindAllByYearAndPeriod(year int, period string) ([]models.ReportTimeliness, error)
	ReplaceByYearAndPeriod(year int, period string, entities []models.ReportTimeliness) error
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

func (r *reportTimelinessRepository) FindAllByYearAndPeriod(year int, period string) ([]models.ReportTimeliness, error) {
	var list []models.ReportTimeliness
	if err := r.DB.Where("year = ? AND period = ?", year, period).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *reportTimelinessRepository) ReplaceByYearAndPeriod(year int, period string, entities []models.ReportTimeliness) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("year = ? AND period = ?", year, period).Delete(&models.ReportTimeliness{}).Error; err != nil {
			return err
		}
		if len(entities) > 0 {
			if err := tx.Create(&entities).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

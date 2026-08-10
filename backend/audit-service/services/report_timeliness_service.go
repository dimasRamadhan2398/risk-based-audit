package services

import (
	"audit-service/models"
	"audit-service/repositories"
)

type ReportTimelinessReq struct {
	Year                   int     `json:"year"`
	Period                 string  `json:"period"`
	QuestionnaireName      string  `json:"questionnaire_name"`
	TotalReportsPlanned    int     `json:"total_reports_planned"`
	TotalReportsCompleted  int     `json:"total_reports_completed"`
	ReportsCompletedOnTime int     `json:"reports_completed_on_time"`
	TimelinessPercentage   float64 `json:"timeliness_percentage"`
	Remarks                string  `json:"remarks"`
}

type IReportTimelinessService interface {
	SaveMultiple(year int, period string, reqs []ReportTimelinessReq) ([]models.ReportTimeliness, error)
	GetByYearAndPeriod(year int, period string) ([]models.ReportTimeliness, error)
}

type reportTimelinessService struct {
	repo repositories.IReportTimelinessRepository
}

func NewReportTimelinessService(repo repositories.IReportTimelinessRepository) IReportTimelinessService {
	return &reportTimelinessService{repo: repo}
}

func (s *reportTimelinessService) SaveMultiple(year int, period string, reqs []ReportTimelinessReq) ([]models.ReportTimeliness, error) {
	var entities []models.ReportTimeliness
	for _, req := range reqs {
		entity := models.ReportTimeliness{
			Year:                   year,
			Period:                 period,
			QuestionnaireName:      req.QuestionnaireName,
			TotalReportsPlanned:    req.TotalReportsPlanned,
			TotalReportsCompleted:  req.TotalReportsCompleted,
			ReportsCompletedOnTime: req.ReportsCompletedOnTime,
			TimelinessPercentage:   req.TimelinessPercentage,
			Remarks:                req.Remarks,
		}
		entities = append(entities, entity)
	}

	err := s.repo.ReplaceByYearAndPeriod(year, period, entities)
	if err != nil {
		return nil, err
	}
	return entities, nil
}

func (s *reportTimelinessService) GetByYearAndPeriod(year int, period string) ([]models.ReportTimeliness, error) {
	return s.repo.FindAllByYearAndPeriod(year, period)
}

package services

import (
	"audit-service/models"
	"audit-service/repositories"
)

type ReportTimelinessReq struct {
	Year                   int     `json:"year"`
	Period                 string  `json:"period"`
	TotalReportsPlanned    int     `json:"total_reports_planned"`
	TotalReportsCompleted  int     `json:"total_reports_completed"`
	ReportsCompletedOnTime int     `json:"reports_completed_on_time"`
	TimelinessPercentage   float64 `json:"timeliness_percentage"`
	Remarks                string  `json:"remarks"`
}

type IReportTimelinessService interface {
	CreateOrUpdate(req *ReportTimelinessReq) (*models.ReportTimeliness, error)
	GetByYearAndPeriod(year int, period string) (*models.ReportTimeliness, error)
}

type reportTimelinessService struct {
	repo repositories.IReportTimelinessRepository
}

func NewReportTimelinessService(repo repositories.IReportTimelinessRepository) IReportTimelinessService {
	return &reportTimelinessService{repo: repo}
}

func (s *reportTimelinessService) CreateOrUpdate(req *ReportTimelinessReq) (*models.ReportTimeliness, error) {
	existing, err := s.repo.FindByYearAndPeriod(req.Year, req.Period)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Update existing
		existing.TotalReportsPlanned = req.TotalReportsPlanned
		existing.TotalReportsCompleted = req.TotalReportsCompleted
		existing.ReportsCompletedOnTime = req.ReportsCompletedOnTime
		existing.TimelinessPercentage = req.TimelinessPercentage
		existing.Remarks = req.Remarks

		// Depending on base repository update
		if br, ok := s.repo.(interface{ Update(interface{}) error }); ok {
			err = br.Update(existing)
		} else {
			// fallback
		}
		
		return existing, err
	}

	entity := &models.ReportTimeliness{
		Year:                   req.Year,
		Period:                 req.Period,
		TotalReportsPlanned:    req.TotalReportsPlanned,
		TotalReportsCompleted:  req.TotalReportsCompleted,
		ReportsCompletedOnTime: req.ReportsCompletedOnTime,
		TimelinessPercentage:   req.TimelinessPercentage,
		Remarks:                req.Remarks,
	}

	err = s.repo.Create(entity)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *reportTimelinessService) GetByYearAndPeriod(year int, period string) (*models.ReportTimeliness, error) {
	return s.repo.FindByYearAndPeriod(year, period)
}

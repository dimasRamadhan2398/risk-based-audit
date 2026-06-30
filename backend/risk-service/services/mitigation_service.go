package services

import (
	"encoding/json"
	"risk-service/models"
	"risk-service/repositories"

	"github.com/google/uuid"
)

type IMitigationService interface {
	GetAll(riskID uuid.UUID) ([]models.RiskMitigation, error)
	Create(mitigation *models.RiskMitigation) (*models.RiskMitigation, error)
	Update(id uuid.UUID, req *models.RiskMitigation) (*models.RiskMitigation, error)
	Delete(id uuid.UUID) error
}

type mitigationService struct {
	repo repositories.IMitigationRepository
}

func NewMitigationService(repo repositories.IMitigationRepository) IMitigationService {
	return &mitigationService{repo: repo}
}

func (s *mitigationService) GetAll(riskID uuid.UUID) ([]models.RiskMitigation, error) {
	mitigations, err := s.repo.FindAll(riskID)
	if err != nil {
		return nil, err
	}

	for i := range mitigations {
		if mitigations[i].MonitoringData != "" {
			json.Unmarshal([]byte(mitigations[i].MonitoringData), &mitigations[i].Monitoring)
		}
		if len(mitigations[i].Monitoring) == 0 {
			mitigations[i].Monitoring = models.GenerateMonitoringChecks(mitigations[i].StartDate, mitigations[i].EndDate)
		}
	}

	return mitigations, nil
}

func (s *mitigationService) Create(mitigation *models.RiskMitigation) (*models.RiskMitigation, error) {
	if mitigation.ID == uuid.Nil {
		mitigation.ID = uuid.New()
	}

	if len(mitigation.Monitoring) == 0 {
		mitigation.Monitoring = models.GenerateMonitoringChecks(mitigation.StartDate, mitigation.EndDate)
	}

	if data, err := json.Marshal(mitigation.Monitoring); err == nil {
		mitigation.MonitoringData = string(data)
	}

	if err := s.repo.Create(mitigation); err != nil {
		return nil, err
	}

	return mitigation, nil
}

func (s *mitigationService) Update(id uuid.UUID, req *models.RiskMitigation) (*models.RiskMitigation, error) {
	mitigation, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	mitigation.RiskEvent = req.RiskEvent
	mitigation.MitigationPlan = req.MitigationPlan
	mitigation.Supervisor = req.Supervisor
	mitigation.PIC = req.PIC
	mitigation.UnitInCharge = req.UnitInCharge

	dateChanged := !mitigation.StartDate.Equal(req.StartDate) || !mitigation.EndDate.Equal(req.EndDate)
	mitigation.StartDate = req.StartDate
	mitigation.EndDate = req.EndDate
	mitigation.Notes = req.Notes

	if len(req.Monitoring) > 0 {
		mitigation.Monitoring = req.Monitoring
	} else if dateChanged {
		mitigation.Monitoring = models.GenerateMonitoringChecks(req.StartDate, req.EndDate)
	} else {
		if mitigation.MonitoringData != "" {
			json.Unmarshal([]byte(mitigation.MonitoringData), &mitigation.Monitoring)
		}
	}

	if data, err := json.Marshal(mitigation.Monitoring); err == nil {
		mitigation.MonitoringData = string(data)
	}

	if err := s.repo.Save(mitigation); err != nil {
		return nil, err
	}

	return mitigation, nil
}

func (s *mitigationService) Delete(id uuid.UUID) error {
	mitigation, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	return s.repo.Delete(mitigation)
}

package confidentiality

import (
	"auth-service/models"
	"auth-service/repositories"
	"time"

	"github.com/google/uuid"
)

// ConfidentialityServiceInterface defines the confidentiality service interface
type ConfidentialityServiceInterface interface {
	GetStatus(userID uuid.UUID) (bool, *time.Time, error)
	Accept(userID uuid.UUID, req *models.AcceptConfidentialityRequest, ipAddress, userAgent string) error
}

// ConfidentialityService handles confidentiality agreement business logic
type ConfidentialityService struct {
	repo repositories.IConfidentialityAgreementRepository
}

// NewConfidentialityService creates a new confidentiality service
func NewConfidentialityService(repo repositories.IConfidentialityAgreementRepository) ConfidentialityServiceInterface {
	return &ConfidentialityService{repo: repo}
}

// GetStatus checks if the user has already accepted the latest confidentiality agreement
func (s *ConfidentialityService) GetStatus(userID uuid.UUID) (bool, *time.Time, error) {
	accepted, err := s.repo.HasAcceptedLatest(userID)
	if err != nil {
		return false, nil, err
	}
	if !accepted {
		return false, nil, nil
	}
	// Fetch the latest accepted agreement to get its timestamp
	agreement, err := s.repo.FindLatestByUserID(userID, "SYSTEM")
	if err != nil {
		return true, nil, nil
	}
	return true, agreement.AcceptedAt, nil
}

// Accept records the user's acceptance of the confidentiality agreement
func (s *ConfidentialityService) Accept(userID uuid.UUID, req *models.AcceptConfidentialityRequest, ipAddress, userAgent string) error {
	now := time.Now()
	agreement := &models.ConfidentialityAgreement{
		UserID:        userID,
		AgreementType: models.AgreementType(req.AgreementType),
		Title:         req.Title,
		Content:       req.Content,
		Version:       req.Version,
		IsAccepted:    true,
		AcceptedAt:    &now,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}
	return s.repo.Create(agreement)
}

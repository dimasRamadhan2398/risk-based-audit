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
	AcceptanceSummary(employeeCodes []string) ([]models.ConfidentialityAcceptanceItem, error)
}

// ConfidentialityService handles confidentiality agreement business logic
type ConfidentialityService struct {
	repo repositories.IConfidentialityAgreementRepository
	now  func() time.Time
}

// NewConfidentialityService creates a new confidentiality service
func NewConfidentialityService(repo repositories.IConfidentialityAgreementRepository) ConfidentialityServiceInterface {
	return &ConfidentialityService{repo: repo, now: time.Now}
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

// AcceptanceSummary reports, per employee code and in the order given, whether
// the code has an active login account and whether that account accepted the
// SYSTEM agreement at models.CurrentSystemAgreementVersion (unexpired). ROLE
// and PROJECT agreements and older SYSTEM versions do not count (plan Q11).
// The caller has already trimmed, de-duplicated and capped the codes.
func (s *ConfidentialityService) AcceptanceSummary(employeeCodes []string) ([]models.ConfidentialityAcceptanceItem, error) {
	now := time.Now
	if s.now != nil {
		now = s.now
	}
	found, err := s.repo.FindAcceptanceByEmployeeCodes(employeeCodes, models.AgreementTypeSystem, models.CurrentSystemAgreementVersion, now())
	if err != nil {
		return nil, err
	}
	items := make([]models.ConfidentialityAcceptanceItem, 0, len(employeeCodes))
	for _, code := range employeeCodes {
		item := models.ConfidentialityAcceptanceItem{EmployeeCode: code, HasUser: found.Linked[code]}
		if at, ok := found.AcceptedAt[code]; ok && item.HasUser {
			item.Accepted = true
			version := models.CurrentSystemAgreementVersion
			item.Version = &version
			if !at.IsZero() {
				t := at.UTC()
				item.AcceptedAt = &t
			}
		}
		items = append(items, item)
	}
	return items, nil
}

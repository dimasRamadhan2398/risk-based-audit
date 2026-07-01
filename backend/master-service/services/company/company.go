package company
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/company"
	"strings"

	"github.com/google/uuid"
)
type CompanyServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Company, error)
	FindById(ctx *base.BaseService, id string) (*models.Company, error)
	Create(ctx *base.BaseService, company *models.Company) (*models.Company, error)
	Update(ctx *base.BaseService, id string, company *models.Company) (*models.Company, error)
	Delete(ctx *base.BaseService, id string) error
}
type CompanyService struct { companyRepo repo.ICompanyRepository }
func NewCompanyService(companyRepo repo.ICompanyRepository) CompanyServiceInterface { return &CompanyService{companyRepo: companyRepo} }
func (s *CompanyService) Create(ctx *base.BaseService, company *models.Company) (*models.Company, error) {
	if s.companyRepo == nil {
		return nil, apperrors.Wrap("SERVICE_UNAVAILABLE", "Company repository is unavailable", 500, nil)
	}
	if company == nil {
		return nil, apperrors.Wrap("INVALID_REQUEST", "Company payload is required", 400, nil)
	}

	company.CompanyCode = strings.TrimSpace(company.CompanyCode)
	company.CompanyName = strings.TrimSpace(company.CompanyName)
	if company.CompanyCode == "" {
		return nil, apperrors.Wrap("VALIDATION_ERROR", "Company code is required", 400, nil)
	}

	if _, err := s.companyRepo.FindByCode(company.CompanyCode); err == nil {
		return nil, apperrors.Wrap("COMPANY_CODE_ALREADY_EXISTS", "Company code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate company code", 500, err)
	}
	if err := s.companyRepo.Create(company); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create company", 500, err) }
	return company, nil
}
func (s *CompanyService) Delete(ctx *base.BaseService, id string) error {
	companyID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_COMPANY_ID", "Invalid company ID format", 400, err) }
	if _, err := s.companyRepo.FindByID(companyID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find company", 500, err)
	}
	if err := s.companyRepo.Delete(companyID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete company", 500, err) }
	return nil
}
func (s *CompanyService) FindAll(ctx *base.BaseService) (*[]models.Company, error) {
	companies, err := s.companyRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch companies", 500, err) }
	result := make([]models.Company, 0, len(companies))
	for _, company := range companies { if company != nil { result = append(result, *company) } }
	return &result, nil
}
func (s *CompanyService) FindById(ctx *base.BaseService, id string) (*models.Company, error) {
	companyID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_COMPANY_ID", "Invalid company ID format", 400, err) }
	company, err := s.companyRepo.FindByID(companyID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch company", 500, err)
	}
	return company, nil
}
func (s *CompanyService) Update(ctx *base.BaseService, id string, company *models.Company) (*models.Company, error) {
	if s.companyRepo == nil {
		return nil, apperrors.Wrap("SERVICE_UNAVAILABLE", "Company repository is unavailable", 500, nil)
	}
	if company == nil {
		return nil, apperrors.Wrap("INVALID_REQUEST", "Company payload is required", 400, nil)
	}

	companyID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_COMPANY_ID", "Invalid company ID format", 400, err) }
	existingCompany, err := s.companyRepo.FindByID(companyID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch company", 500, err)
	}
	if existingCompany.CompanyCode != company.CompanyCode {
		if _, err := s.companyRepo.FindByCode(company.CompanyCode); err == nil {
			return nil, apperrors.Wrap("COMPANY_CODE_ALREADY_EXISTS", "Company code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate company code", 500, err)
		}
	}
	existingCompany.CompanyCode = company.CompanyCode
	existingCompany.CompanyName = company.CompanyName
	existingCompany.LegalName = company.LegalName
	existingCompany.TaxID = company.TaxID
	existingCompany.CompanyType = company.CompanyType
	existingCompany.ParentID = company.ParentID
	existingCompany.LocationID = company.LocationID
	existingCompany.Phone = company.Phone
	existingCompany.Email = company.Email
	existingCompany.Website = company.Website
	existingCompany.IsActive = company.IsActive
	existingCompany.EstablishedAt = company.EstablishedAt
	if err := s.companyRepo.Update(existingCompany); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update company", 500, err) }
	return existingCompany, nil
}

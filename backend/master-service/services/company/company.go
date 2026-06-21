package company

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	companyRepo "master-service/repositories/company"

	"github.com/google/uuid"
)

type CompanyServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Company, error)
	FindById(ctx *base.BaseService, id string) (*models.Company, error)
	FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.Company, int64, error)
	Create(ctx *base.BaseService, company *models.Company) (*models.Company, error)
	Update(ctx *base.BaseService, id string, company *models.Company) (*models.Company, error)
	Delete(ctx *base.BaseService, id string) error
}

type CompanyService struct {
	companyRepo companyRepo.ICompanyRepository
}

// Create implements CompanyServiceInterface.
func (s *CompanyService) Create(ctx *base.BaseService, company *models.Company) (*models.Company, error) {
	if err := s.validateReferences(company); err != nil {
		return nil, err
	}

	if _, err := s.companyRepo.FindByName(company.CompanyName); err == nil {
		return nil, apperrors.Wrap("COMPANY_NAME_ALREADY_EXISTS", "Company name already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate company name", 500, err)
	}

	if err := s.companyRepo.Create(company); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create company", 500, err)
	}

	return company, nil
}

// Delete implements CompanyServiceInterface.
func (s *CompanyService) Delete(ctx *base.BaseService, id string) error {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.Wrap("INVALID_COMPANY_ID", "Invalid company ID format", 400, err)
	}

	if _, err := s.companyRepo.FindByID(companyID); err != nil {
		if err == apperrors.ErrNotFound {
			return err
		}
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find company", 500, err)
	}

	if err := s.companyRepo.Delete(companyID); err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to delete company", 500, err)
	}

	return nil
}

// FindAll implements CompanyServiceInterface.
func (s *CompanyService) FindAll(ctx *base.BaseService) (*[]models.Company, error) {
	companies, err := s.companyRepo.FindAll()
	if err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch companies", 500, err)
	}

	result := make([]models.Company, 0, len(companies))
	for _, company := range companies {
		if company != nil {
			result = append(result, *company)
		}
	}

	return &result, nil
}

// FindById implements CompanyServiceInterface.
func (s *CompanyService) FindById(ctx *base.BaseService, id string) (*models.Company, error) {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_COMPANY_ID", "Invalid company ID format", 400, err)
	}

	company, err := s.companyRepo.FindByID(companyID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch company", 500, err)
	}

	return company, nil
}

// FindMany finds companies with pagination
func (s *CompanyService) FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.Company, int64, error) {
	companies, err := s.companyRepo.FindMany(offset, limit, search)
	if err != nil {
		return nil, 0, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch companies", 500, err)
	}

	count, err := s.companyRepo.Count(search)
	if err != nil {
		return nil, 0, apperrors.Wrap("DATABASE_ERROR", "Failed to count companies", 500, err)
	}

	result := make([]models.Company, 0, len(companies))
	for _, company := range companies {
		if company != nil {
			result = append(result, *company)
		}
	}

	return &result, count, nil
}

// Update implements CompanyServiceInterface.
func (s *CompanyService) Update(ctx *base.BaseService, id string, company *models.Company) (*models.Company, error) {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_COMPANY_ID", "Invalid company ID format", 400, err)
	}

	existingCompany, err := s.companyRepo.FindByID(companyID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch company", 500, err)
	}

	if err := s.validateReferences(company); err != nil {
		return nil, err
	}

	if existingCompany.CompanyName != company.CompanyName {
		if _, err := s.companyRepo.FindByName(company.CompanyName); err == nil {
			return nil, apperrors.Wrap("COMPANY_NAME_ALREADY_EXISTS", "Company name already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate company name", 500, err)
		}
	}

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

	if err := s.companyRepo.Update(existingCompany); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update company", 500, err)
	}

	return existingCompany, nil
}

func (s *CompanyService) validateReferences(company *models.Company) error {
	// Validate parent company if provided
	if company.ParentID != nil && *company.ParentID != uuid.Nil {
		parentExists, err := s.companyRepo.FindByID(*company.ParentID)
		if err != nil {
			return apperrors.Wrap("DATABASE_ERROR", "Failed to validate parent company", 500, err)
		}
		if parentExists == nil {
			return apperrors.Wrap("PARENT_COMPANY_NOT_FOUND", "Parent company not found", 404, nil)
		}
		// Prevent circular reference
		if *company.ParentID == company.ID {
			return apperrors.Wrap("CIRCULAR_REFERENCE", "Company cannot be its own parent", 400, nil)
		}
	}

	// Validate location if provided
	if company.LocationID != nil && *company.LocationID != uuid.Nil {
		locationExists, err := s.companyRepo.LocationExists(*company.LocationID)
		if err != nil {
			return apperrors.Wrap("DATABASE_ERROR", "Failed to validate location", 500, err)
		}
		if !locationExists {
			return apperrors.Wrap("LOCATION_NOT_FOUND", "Location not found", 404, nil)
		}
	}

	return nil
}

func NewCompanyService(companyRepo companyRepo.ICompanyRepository) CompanyServiceInterface {
	return &CompanyService{
		companyRepo: companyRepo,
	}
}

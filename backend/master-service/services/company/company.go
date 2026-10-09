package company

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/company"
	"net/http"
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
type CompanyService struct{ companyRepo repo.ICompanyRepository }

var errCompanyCodeExists = apperrors.New(apperrors.CodeCompanyCodeAlreadyExists,
	"A company with this code already exists.", http.StatusConflict).
	WithFields(map[string]string{"code": apperrors.FieldAlreadyExists})

var errInvalidCompanyBody = apperrors.New(apperrors.CodeInvalidRequestBody, apperrors.MsgInvalidRequestBody, http.StatusBadRequest)

func NewCompanyService(companyRepo repo.ICompanyRepository) CompanyServiceInterface {
	return &CompanyService{companyRepo: companyRepo}
}
func (s *CompanyService) Create(ctx *base.BaseService, company *models.Company) (*models.Company, error) {
	if s.companyRepo == nil {
		return nil, apperrors.Internal("", nil)
	}
	if company == nil {
		return nil, errInvalidCompanyBody
	}

	company.CompanyCode = strings.TrimSpace(company.CompanyCode)
	company.CompanyName = strings.TrimSpace(company.CompanyName)
	if company.CompanyCode == "" {
		return nil, apperrors.ValidationFailed("", map[string]string{"code": apperrors.FieldRequired})
	}

	if _, err := s.companyRepo.FindByCode(company.CompanyCode); err == nil {
		return nil, errCompanyCodeExists
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the company code.")
	}
	if err := s.companyRepo.Create(company); err != nil {
		return nil, apperrors.DB(err, apperrors.OpCreate, "Failed to create the company.")
	}
	return company, nil
}
func (s *CompanyService) Delete(ctx *base.BaseService, id string) error {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.InvalidID(err)
	}
	if _, err := s.companyRepo.FindByID(companyID); err != nil {
		if err == apperrors.ErrNotFound {
			return apperrors.ErrCompanyNotFound
		}
		return apperrors.DB(err, apperrors.OpRead, "Failed to load the company.")
	}
	if err := s.companyRepo.Delete(companyID); err != nil {
		return apperrors.DB(err, apperrors.OpDelete, "Failed to delete the company.")
	}
	return nil
}
func (s *CompanyService) FindAll(ctx *base.BaseService) (*[]models.Company, error) {
	companies, err := s.companyRepo.FindAll()
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load companies.")
	}
	result := make([]models.Company, 0, len(companies))
	for _, company := range companies {
		if company != nil {
			result = append(result, *company)
		}
	}
	return &result, nil
}
func (s *CompanyService) FindById(ctx *base.BaseService, id string) (*models.Company, error) {
	companyID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.InvalidID(err)
	}
	company, err := s.companyRepo.FindByID(companyID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, apperrors.ErrCompanyNotFound
		}
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load the company.")
	}
	return company, nil
}
func (s *CompanyService) Update(ctx *base.BaseService, id string, company *models.Company) (*models.Company, error) {
	if s.companyRepo == nil {
		return nil, apperrors.Internal("", nil)
	}
	if company == nil {
		return nil, errInvalidCompanyBody
	}

	companyID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.InvalidID(err)
	}
	existingCompany, err := s.companyRepo.FindByID(companyID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, apperrors.ErrCompanyNotFound
		}
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load the company.")
	}
	if existingCompany.CompanyCode != company.CompanyCode {
		if _, err := s.companyRepo.FindByCode(company.CompanyCode); err == nil {
			return nil, errCompanyCodeExists
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the company code.")
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
	if err := s.companyRepo.Update(existingCompany); err != nil {
		return nil, apperrors.DB(err, apperrors.OpUpdate, "Failed to update the company.")
	}
	return existingCompany, nil
}

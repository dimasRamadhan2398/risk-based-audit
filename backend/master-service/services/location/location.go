package location

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/reference"
	repo "master-service/repositories/location"

	"github.com/google/uuid"
)

type LocationServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Location, error)
	FindById(ctx *base.BaseService, id string) (*models.Location, error)
	Create(ctx *base.BaseService, location *models.Location) (*models.Location, error)
	Update(ctx *base.BaseService, id string, location *models.Location) (*models.Location, error)
	Delete(ctx *base.BaseService, id string) error
}

type LocationService struct {
	locationRepo repo.ILocationRepository
}

func NewLocationService(locationRepo repo.ILocationRepository) LocationServiceInterface {
	return &LocationService{
		locationRepo: locationRepo,
	}
}

const msgLocationRegionInvalid = "Country, province and city must be chosen from the reference list (GET /api/v1/reference/...), and the city must be in the chosen province."

// checkLocationNames validates country, province and city against the embedded
// reference data and replaces them with the canonical names. existing is the
// stored record on update (nil on create): its unchanged values are accepted
// as they are. Province/city are only checked for Indonesia (or an empty
// country, which the column defaults to Indonesia). See pkg/reference/validate.go.
func checkLocationNames(location, existing *models.Location) error {
	ref := reference.MustDefault()
	country := reference.NameField{Key: "country", Value: &location.Country}
	province := reference.NameField{Key: "province", Value: &location.Province}
	city := reference.NameField{Key: "city", Value: &location.City}
	if existing != nil {
		country.Stored = &existing.Country
		province.Stored = &existing.Province
		city.Stored = &existing.City
	}
	fields := map[string]string{}
	if isIndonesia, _ := ref.CheckCountry(country, fields); isIndonesia {
		ref.CheckRegion(province, city, fields)
	}
	if len(fields) > 0 {
		return apperrors.ValidationFailed(msgLocationRegionInvalid, fields)
	}
	return nil
}

func (s *LocationService) Create(ctx *base.BaseService, location *models.Location) (*models.Location, error) {
	if err := checkLocationNames(location, nil); err != nil {
		return nil, err
	}
	if err := s.locationRepo.Create(location); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create location", 500, err)
	}

	return location, nil
}

func (s *LocationService) Delete(ctx *base.BaseService, id string) error {
	locationID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.Wrap("INVALID_LOCATION_ID", "Invalid location ID format", 400, err)
	}

	if _, err := s.locationRepo.FindByID(locationID); err != nil {
		if err == apperrors.ErrNotFound {
			return err
		}
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find location", 500, err)
	}

	if err := s.locationRepo.Delete(locationID); err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to delete location", 500, err)
	}

	return nil
}

func (s *LocationService) FindAll(ctx *base.BaseService) (*[]models.Location, error) {
	locations, err := s.locationRepo.FindAll()
	if err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch locations", 500, err)
	}

	result := make([]models.Location, 0, len(locations))
	for _, location := range locations {
		if location != nil {
			result = append(result, *location)
		}
	}

	return &result, nil
}

func (s *LocationService) FindById(ctx *base.BaseService, id string) (*models.Location, error) {
	locationID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_LOCATION_ID", "Invalid location ID format", 400, err)
	}

	location, err := s.locationRepo.FindByID(locationID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch location", 500, err)
	}

	return location, nil
}

func (s *LocationService) Update(ctx *base.BaseService, id string, location *models.Location) (*models.Location, error) {
	locationID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_LOCATION_ID", "Invalid location ID format", 400, err)
	}

	existingLocation, err := s.locationRepo.FindByID(locationID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch location", 500, err)
	}

	if err := checkLocationNames(location, existingLocation); err != nil {
		return nil, err
	}

	existingLocation.Name = location.Name
	existingLocation.Address = location.Address
	existingLocation.City = location.City
	existingLocation.Province = location.Province
	existingLocation.PostalCode = location.PostalCode
	existingLocation.Country = location.Country
	existingLocation.Latitude = location.Latitude
	existingLocation.Longitude = location.Longitude
	existingLocation.IsActive = location.IsActive

	if err := s.locationRepo.Update(existingLocation); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update location", 500, err)
	}

	return existingLocation, nil
}

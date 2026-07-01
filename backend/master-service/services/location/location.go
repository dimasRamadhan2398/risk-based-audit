package location

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
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

func (s *LocationService) Create(ctx *base.BaseService, location *models.Location) (*models.Location, error) {
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

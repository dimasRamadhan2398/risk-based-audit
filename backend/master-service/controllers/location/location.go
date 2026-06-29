package location

import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	locationSvc "master-service/services/location"

	"github.com/gin-gonic/gin"
)

type LocationControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type LocationController struct {
	*base.BaseController
	locationSvc locationSvc.LocationServiceInterface
}

func NewLocationController(locationSvc locationSvc.LocationServiceInterface, validator *validations.Validator) LocationControllerInterface {
	return &LocationController{
		BaseController: base.NewBaseController(validator),
		locationSvc:    locationSvc,
	}
}

func (d *LocationController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	locations, err := d.locationSvc.FindAll(baseService)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Locations fetched successfully", locations)
}

func (d *LocationController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	location, err := d.locationSvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Location fetched successfully", location)
}

func (d *LocationController) Create(ctx *gin.Context) {
	var req models.Location
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	location, err := d.locationSvc.Create(baseService, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.Created(ctx, "Location created successfully", location)
}

func (d *LocationController) Update(ctx *gin.Context) {
	var req models.Location
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	location, err := d.locationSvc.Update(baseService, id, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Location updated successfully", location)
}

func (d *LocationController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	if err := d.locationSvc.Delete(baseService, id); err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Location deleted successfully", nil)
}

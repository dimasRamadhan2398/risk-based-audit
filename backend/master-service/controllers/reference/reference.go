// Package reference serves the embedded reference data (pkg/reference):
// provinces, kabupaten/kota and countries. Read-only, full lists, no
// pagination; responses carry the dataset version as ETag.
package reference

import (
	"net/http"
	"strconv"
	"strings"

	"master-service/pkg/reference"
	"master-service/pkg/response"

	"github.com/gin-gonic/gin"
)

// CacheControl: the data only changes with a deploy, and the ETag lets a
// client revalidate cheaply after max-age. private because the routes sit
// behind authentication.
const CacheControl = "private, max-age=86400"

type Controller struct {
	data *reference.Dataset
}

func New(data *reference.Dataset) *Controller {
	return &Controller{data: data}
}

// notModified sets ETag/Cache-Control and answers 304 when the client already
// has this version. It returns true when the response is complete.
func notModified(c *gin.Context, version string) bool {
	etag := `"` + version + `"`
	c.Header("ETag", etag)
	c.Header("Cache-Control", CacheControl)
	for _, tag := range strings.Split(c.GetHeader("If-None-Match"), ",") {
		tag = strings.TrimPrefix(strings.TrimSpace(tag), "W/")
		if tag == etag || tag == "*" {
			c.Status(http.StatusNotModified)
			return true
		}
	}
	return false
}

func includeInactive(c *gin.Context) (bool, bool) {
	raw := strings.TrimSpace(c.Query("include_inactive"))
	if raw == "" {
		return false, true
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		response.BadRequest(c, "include_inactive must be true or false")
		return false, false
	}
	return v, true
}

// Provinces: GET /reference/provinces[?include_inactive=true]
func (h *Controller) Provinces(c *gin.Context) {
	inactive, ok := includeInactive(c)
	if !ok || notModified(c, h.data.RegionsVersion()) {
		return
	}
	response.OK(c, "Provinces fetched successfully", h.data.Provinces(inactive))
}

// ProvinceRegencies: GET /reference/provinces/:code/regencies
func (h *Controller) ProvinceRegencies(c *gin.Context) {
	inactive, ok := includeInactive(c)
	if !ok {
		return
	}
	code := strings.TrimSpace(c.Param("code"))
	if _, found := h.data.Province(code); !found {
		response.NotFound(c, "Province not found")
		return
	}
	if notModified(c, h.data.RegionsVersion()) {
		return
	}
	response.OK(c, "Regencies fetched successfully", h.data.SearchRegencies("", code, inactive))
}

// Regencies: GET /reference/regencies?q=&province=
func (h *Controller) Regencies(c *gin.Context) {
	inactive, ok := includeInactive(c)
	if !ok {
		return
	}
	province := strings.TrimSpace(c.Query("province"))
	if province != "" {
		if _, found := h.data.Province(province); !found {
			response.NotFound(c, "Province not found")
			return
		}
	}
	if notModified(c, h.data.RegionsVersion()) {
		return
	}
	response.OK(c, "Regencies fetched successfully", h.data.SearchRegencies(c.Query("q"), province, inactive))
}

// Countries: GET /reference/countries?q=
func (h *Controller) Countries(c *gin.Context) {
	inactive, ok := includeInactive(c)
	if !ok || notModified(c, h.data.CountriesVersion()) {
		return
	}
	response.OK(c, "Countries fetched successfully", h.data.SearchCountries(c.Query("q"), inactive))
}

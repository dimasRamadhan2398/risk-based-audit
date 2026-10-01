package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	// Code mirrors Error.Code on error responses so clients can read a stable
	// machine-readable code at the top level. Omitted on success.
	Code  string       `json:"code,omitempty"`
	Data  interface{}  `json:"data,omitempty"`
	Error *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	// Fields maps request field names (JSON) to field-level codes
	// (e.g. REQUIRED, INVALID_FORMAT, ALREADY_EXISTS). Validation errors only.
	Fields map[string]string `json:"fields,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, code string, message string, details string) {
	ErrorWithFields(c, statusCode, code, message, details, nil)
}

// ErrorWithFields sends an error response including field-level codes.
// details must never contain raw DB/driver output.
func ErrorWithFields(c *gin.Context, statusCode int, code string, message string, details string, fields map[string]string) {
	c.JSON(statusCode, Response{
		Success: false,
		Code:    code,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
			Fields:  fields,
		},
	})
}

// OK sends a 200 OK response
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// Created sends a 201 Created response
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message, "")
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message, "")
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message, "")
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message, "")
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, "")
}

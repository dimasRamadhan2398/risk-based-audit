package errors

import "net/http"

// Stable error codes returned in API error responses. The frontend translates
// these under errors.codes.<CODE>; do not rename existing values.
const (
	// Generic
	CodeValidationFailed    = "VALIDATION_FAILED"
	CodeInvalidRequestBody  = "INVALID_REQUEST_BODY"
	CodeInvalidID           = "INVALID_ID"
	CodeValueTooLong        = "VALUE_TOO_LONG"
	CodeNotFound            = "NOT_FOUND"
	CodeDuplicateEntry      = "DUPLICATE_ENTRY"
	CodeReferenceNotFound   = "REFERENCE_NOT_FOUND"
	CodeForeignKeyViolation = "FOREIGN_KEY_VIOLATION"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeInternal            = "INTERNAL_ERROR"

	// Resource specific
	CodeEmployeeNotFound           = "EMPLOYEE_NOT_FOUND"
	CodeEmployeeCodeAlreadyExists  = "EMPLOYEE_CODE_ALREADY_EXISTS"
	CodeEmployeeEmailAlreadyExists = "EMPLOYEE_EMAIL_ALREADY_EXISTS"

	CodeDepartmentNotFound          = "DEPARTMENT_NOT_FOUND"
	CodeDepartmentCodeAlreadyExists = "DEPARTMENT_CODE_ALREADY_EXISTS"
	CodeDepartmentNameAlreadyExists = "DEPARTMENT_NAME_ALREADY_EXISTS"

	CodeCompanyNotFound           = "COMPANY_NOT_FOUND"
	CodeCompanyCodeAlreadyExists  = "COMPANY_CODE_ALREADY_EXISTS"
	CodeCompanyTaxIDAlreadyExists = "COMPANY_TAX_ID_ALREADY_EXISTS"

	CodeBusinessUnitNotFound        = "BUSINESS_UNIT_NOT_FOUND"
	CodeBusinessUnitCompanyMismatch = "BUSINESS_UNIT_COMPANY_MISMATCH"
	CodePICNotFound                 = "PIC_NOT_FOUND"
	CodePICCompanyMismatch          = "PIC_COMPANY_MISMATCH"
)

// Field-level codes used in the `fields` map of VALIDATION_FAILED responses.
const (
	FieldRequired      = "REQUIRED"
	FieldInvalidFormat = "INVALID_FORMAT"
	FieldInvalidType   = "INVALID_TYPE"
	FieldInvalidOption = "INVALID_OPTION"
	FieldOutOfRange    = "OUT_OF_RANGE"
	FieldInvalid       = "INVALID"
	FieldAlreadyExists = "ALREADY_EXISTS"
	FieldNotFound      = "NOT_FOUND"
	FieldMismatch      = "MISMATCH"
)

// Default English messages for generic codes.
const (
	MsgValidationFailed    = "Some fields are missing or invalid. Please check the form and try again."
	MsgInvalidRequestBody  = "The submitted data could not be read. Please check the form and try again."
	MsgInvalidID           = "The provided ID is not valid."
	MsgValueTooLong        = "One of the values is longer than allowed."
	MsgNotFound            = "The requested record was not found."
	MsgDuplicateEntry      = "A record with the same value already exists."
	MsgReferenceNotFound   = "A selected related record does not exist or has been removed."
	MsgForeignKeyViolation = "This record cannot be deleted because it is still used by other records."
	MsgInternal            = "An unexpected error occurred. Please try again later."
)

// Resource-specific sentinels. Callers should return these (optionally via
// WithErr/WithFields, which copy) rather than the generic ErrNotFound.
var (
	ErrEmployeeNotFound   = New(CodeEmployeeNotFound, "Employee not found.", http.StatusNotFound)
	ErrDepartmentNotFound = New(CodeDepartmentNotFound, "Department not found.", http.StatusNotFound)
	ErrCompanyNotFound    = New(CodeCompanyNotFound, "Company not found.", http.StatusNotFound)
)

// InvalidID builds a 400 for a malformed path ID.
func InvalidID(err error) *AppError {
	return Wrap(CodeInvalidID, MsgInvalidID, http.StatusBadRequest, err)
}

// ValidationFailed builds a 400 VALIDATION_FAILED with optional field codes.
// message may be empty to use the default.
func ValidationFailed(message string, fields map[string]string) *AppError {
	if message == "" {
		message = MsgValidationFailed
	}
	e := New(CodeValidationFailed, message, http.StatusBadRequest)
	e.Fields = fields
	return e
}

// Internal builds a 500 INTERNAL_ERROR with a human-readable message and the
// technical cause kept for logging only.
func Internal(message string, err error) *AppError {
	if message == "" {
		message = MsgInternal
	}
	return Wrap(CodeInternal, message, http.StatusInternalServerError, err)
}

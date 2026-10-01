package errors

import (
	stderrors "errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// This file has no master-service imports so it can be lifted into a shared
// module later. Service-specific constraint names live in constraints.go.

// Op is the kind of operation that produced an error. It decides how a
// foreign-key violation is reported: on create/update it means a referenced
// record is missing; on delete it means the record is still in use.
type Op string

const (
	OpRead   Op = "read"
	OpCreate Op = "create"
	OpUpdate Op = "update"
	OpDelete Op = "delete"
)

// OpFromHTTPMethod derives an Op from an HTTP method.
func OpFromHTTPMethod(method string) Op {
	switch strings.ToUpper(method) {
	case http.MethodPost:
		return OpCreate
	case http.MethodPut, http.MethodPatch:
		return OpUpdate
	case http.MethodDelete:
		return OpDelete
	default:
		return OpRead
	}
}

// Postgres SQLSTATE codes we map. See
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation          = "23505"
	pgForeignKeyViolation      = "23503"
	pgNotNullViolation         = "23502"
	pgCheckViolation           = "23514"
	pgStringDataRightTruncated = "22001"
	pgInvalidTextRepresent     = "22P02"
	pgInvalidDatetimeFormat    = "22007"
	pgDatetimeFieldOverflow    = "22008"
	pgNumericValueOutOfRange   = "22003"
)

// ConstraintInfo describes how a named DB constraint should surface to clients.
type ConstraintInfo struct {
	// Code overrides the generic code (e.g. EMPLOYEE_CODE_ALREADY_EXISTS).
	// Leave empty to keep the generic code but still attach Field.
	Code    string
	Message string
	// Field is the JSON request field the constraint relates to.
	Field string
}

// FromDB maps a persistence error to an AppError with a stable code and a
// safe message. It returns nil when err is not a recognised DB condition.
// The original error is kept in AppError.Err for server-side logging only.
func FromDB(err error, op Op) *AppError {
	if err == nil {
		return nil
	}

	if stderrors.Is(err, gorm.ErrRecordNotFound) || stderrors.Is(err, ErrNotFound) {
		return Wrap(CodeNotFound, MsgNotFound, http.StatusNotFound, err)
	}

	var pgErr *pgconn.PgError
	if stderrors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolation:
			return uniqueViolation(pgErr.ConstraintName, err)
		case pgForeignKeyViolation:
			return foreignKeyViolation(pgErr.ConstraintName, op, err)
		case pgNotNullViolation:
			var fields map[string]string
			if pgErr.ColumnName != "" {
				fields = map[string]string{pgErr.ColumnName: FieldRequired}
			}
			return ValidationFailed("", fields).WithErr(err)
		case pgStringDataRightTruncated:
			return Wrap(CodeValueTooLong, MsgValueTooLong, http.StatusBadRequest, err)
		case pgInvalidTextRepresent, pgInvalidDatetimeFormat, pgDatetimeFieldOverflow,
			pgNumericValueOutOfRange, pgCheckViolation:
			return ValidationFailed("", nil).WithErr(err)
		}
		return nil
	}

	// Drivers opened with gorm.Config{TranslateError: true} (e.g. sqlite).
	if stderrors.Is(err, gorm.ErrDuplicatedKey) {
		return uniqueViolation("", err)
	}
	if stderrors.Is(err, gorm.ErrForeignKeyViolated) {
		return foreignKeyViolation("", op, err)
	}
	return nil
}

func uniqueViolation(constraint string, err error) *AppError {
	appErr := Wrap(CodeDuplicateEntry, MsgDuplicateEntry, http.StatusConflict, err)
	if info, ok := lookupConstraint(constraint); ok {
		applyConstraint(appErr, info, FieldAlreadyExists)
	}
	return appErr
}

func foreignKeyViolation(constraint string, op Op, err error) *AppError {
	if op == OpDelete {
		// The deleted row is referenced elsewhere; the constraint belongs to
		// the referencing table, so there is no request field to point at.
		return Wrap(CodeForeignKeyViolation, MsgForeignKeyViolation, http.StatusConflict, err)
	}
	appErr := Wrap(CodeReferenceNotFound, MsgReferenceNotFound, http.StatusUnprocessableEntity, err)
	if info, ok := lookupConstraint(constraint); ok {
		applyConstraint(appErr, info, FieldNotFound)
	}
	return appErr
}

func applyConstraint(appErr *AppError, info ConstraintInfo, fieldCode string) {
	if info.Code != "" {
		appErr.Code = info.Code
	}
	if info.Message != "" {
		appErr.Message = info.Message
	}
	if info.Field != "" {
		appErr.Fields = map[string]string{info.Field: fieldCode}
	}
}

func lookupConstraint(name string) (ConstraintInfo, bool) {
	if name == "" {
		return ConstraintInfo{}, false
	}
	info, ok := constraintRegistry[name]
	return info, ok
}

// Normalize turns any error into a client-safe AppError:
//   - a 4xx AppError is returned as-is (legacy VALIDATION_ERROR is renamed to
//     VALIDATION_FAILED);
//   - DB conditions (not found, unique, FK, bad input) anywhere in the chain
//     are mapped via FromDB, even if a service wrapped them in a 500;
//   - everything else becomes INTERNAL_ERROR, keeping a service-provided
//     human message when there is one.
//
// The returned AppError always carries the original error in Err.
func Normalize(err error, op Op) *AppError {
	if err == nil {
		return nil
	}

	appErr, isApp := As(err)
	if isApp && appErr.StatusCode > 0 && appErr.StatusCode < http.StatusInternalServerError {
		if appErr.Code == ErrValidation.Code {
			cp := *appErr
			cp.Code = CodeValidationFailed
			return &cp
		}
		return appErr
	}

	if mapped := FromDB(err, op); mapped != nil {
		mapped.Err = err
		return mapped
	}

	message := MsgInternal
	if isApp && appErr.Message != "" && appErr.Message != ErrDatabase.Message {
		message = appErr.Message
	}
	return Internal(message, err)
}

// DB is the service-layer helper for repository failures: known DB conditions
// are mapped via FromDB, anything else becomes INTERNAL_ERROR with message.
func DB(err error, op Op, message string) *AppError {
	if mapped := FromDB(err, op); mapped != nil {
		return mapped
	}
	return Internal(message, err)
}

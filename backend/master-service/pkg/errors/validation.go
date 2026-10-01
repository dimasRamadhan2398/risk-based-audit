package errors

import (
	"encoding/json"
	stderrors "errors"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FromBinding converts a request binding/validation error (gin ShouldBind*,
// go-playground/validator, encoding/json) into a client-safe AppError.
// req is the struct that was bound; it is used to report JSON field names.
func FromBinding(err error, req interface{}) *AppError {
	if err == nil {
		return nil
	}

	if stderrors.Is(err, io.EOF) || stderrors.Is(err, io.ErrUnexpectedEOF) {
		return Wrap(CodeInvalidRequestBody, MsgInvalidRequestBody, http.StatusBadRequest, err)
	}

	var syntaxErr *json.SyntaxError
	if stderrors.As(err, &syntaxErr) {
		return Wrap(CodeInvalidRequestBody, MsgInvalidRequestBody, http.StatusBadRequest, err)
	}

	var typeErr *json.UnmarshalTypeError
	if stderrors.As(err, &typeErr) {
		var fields map[string]string
		if typeErr.Field != "" {
			fields = map[string]string{typeErr.Field: FieldInvalidType}
		}
		return ValidationFailed("", fields).WithErr(err)
	}

	var verrs validator.ValidationErrors
	if stderrors.As(err, &verrs) {
		fields := make(map[string]string, len(verrs))
		for _, fe := range verrs {
			fields[jsonFieldPath(req, fe)] = fieldCodeForTag(fe.Tag())
		}
		return ValidationFailed("", fields).WithErr(err)
	}

	// Anything else (e.g. a TextUnmarshaler such as uuid.UUID rejecting "")
	// has no reliable field information.
	return ValidationFailed("", nil).WithErr(err)
}

func fieldCodeForTag(tag string) string {
	switch {
	case tag == "required" || strings.HasPrefix(tag, "required_"):
		return FieldRequired
	case tag == "oneof":
		return FieldInvalidOption
	case tag == "min" || tag == "max" || tag == "len" || tag == "gt" || tag == "gte" || tag == "lt" || tag == "lte":
		return FieldOutOfRange
	case tag == "email" || tag == "url" || tag == "uri" || tag == "datetime" || tag == "numeric" ||
		tag == "number" || tag == "alphanum" || tag == "alpha" || strings.HasPrefix(tag, "uuid") ||
		tag == "e164" || tag == "ip" || tag == "hexadecimal":
		return FieldInvalidFormat
	default:
		return FieldInvalid
	}
}

// jsonFieldPath resolves a validator field error to its JSON path
// (e.g. "department_code" or "items.name"), falling back to the Go name.
func jsonFieldPath(req interface{}, fe validator.FieldError) string {
	parts := strings.Split(fe.StructNamespace(), ".")
	if len(parts) <= 1 || req == nil {
		return fe.Field()
	}

	t := reflect.TypeOf(req)
	out := make([]string, 0, len(parts)-1)
	for _, part := range parts[1:] {
		name := part
		if i := strings.IndexByte(name, '['); i >= 0 {
			name = name[:i]
		}
		for t != nil && (t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map) {
			t = t.Elem()
		}
		if t == nil || t.Kind() != reflect.Struct {
			out = append(out, name)
			t = nil
			continue
		}
		sf, ok := t.FieldByName(name)
		if !ok {
			out = append(out, name)
			t = nil
			continue
		}
		out = append(out, jsonName(sf))
		t = sf.Type
	}
	return strings.Join(out, ".")
}

func jsonName(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag == "" || tag == "-" {
		return sf.Name
	}
	if name := strings.Split(tag, ",")[0]; name != "" {
		return name
	}
	return sf.Name
}

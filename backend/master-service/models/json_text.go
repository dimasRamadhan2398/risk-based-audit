package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONText is a JSON document stored in a json/jsonb (Postgres) or text
// (sqlite) column. It is sent to the driver as a string so Postgres casts it to
// the column type, and it marshals to API clients as the raw document.
type JSONText string

// Value implements driver.Valuer.
func (j JSONText) Value() (driver.Value, error) {
	if j == "" {
		return nil, nil
	}
	return string(j), nil
}

// Scan implements sql.Scanner.
func (j *JSONText) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*j = ""
	case string:
		*j = JSONText(v)
	case []byte:
		*j = JSONText(string(v))
	default:
		return fmt.Errorf("JSONText: unsupported type %T", src)
	}
	return nil
}

// MarshalJSON emits the stored document as-is (null when empty).
func (j JSONText) MarshalJSON() ([]byte, error) {
	if j == "" {
		return []byte("null"), nil
	}
	if !json.Valid([]byte(j)) {
		return json.Marshal(string(j))
	}
	return []byte(j), nil
}

// ToJSONText marshals v into a JSONText.
func ToJSONText(v any) (JSONText, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return JSONText(b), nil
}

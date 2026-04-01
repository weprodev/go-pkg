package uuidutil

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// ID is a UUID-based identifier type that implements sql.Scanner and driver.Valuer.
// Use this as a base for domain-specific ID types that need database persistence.
type ID uuid.UUID

// Scan implements sql.Scanner interface for database reads
func (id *ID) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		parsed, err := uuid.ParseBytes(v)
		if err != nil {
			return err
		}
		*id = ID(parsed)
		return nil
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		*id = ID(parsed)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into ID", src)
	}
}

// Value implements driver.Valuer interface for database writes
func (id ID) Value() (driver.Value, error) {
	return uuid.UUID(id).String(), nil
}

// String returns the string representation of the ID
func (id ID) String() string {
	return uuid.UUID(id).String()
}

// IsNil returns true if the ID is the zero value (nil UUID)
func (id ID) IsNil() bool {
	return uuid.UUID(id) == uuid.Nil
}

// NewID generates a new random ID
func NewID() ID {
	return ID(uuid.New())
}

// ParseID parses a string into an ID
func ParseID(s string) (ID, error) {
	parsed, err := uuid.Parse(s)
	if err != nil {
		return ID{}, err
	}
	return ID(parsed), nil
}

// MustParseID parses a string into an ID, panicking on error
func MustParseID(s string) ID {
	return ID(uuid.MustParse(s))
}

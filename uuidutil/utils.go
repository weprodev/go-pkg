package uuidutil

import (
	"fmt"
	"github.com/weprodev/go-pkg/httperr"

	"github.com/google/uuid"
)

// UUIDValueOrNil safely dereferences a UUID pointer, returning the UUID value or uuid.Nil if nil
func UUIDValueOrNil(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}

// UUIDPointerOrNil converts a UUID value to pointer, returning nil if the UUID is uuid.Nil
func UUIDPointerOrNil(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

// ValidateAndParseUUID validates that a string is not empty and is a valid UUID format
func ValidateAndParseUUID(idStr string, fieldName string) (uuid.UUID, error) {
	name := fieldName
	if name == "" {
		name = "ID"
	}

	if idStr == "" {
		return uuid.Nil, httperr.NewServiceError(httperr.StatusBadRequest, fmt.Sprintf("missing %s", name))
	}

	parsed, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, httperr.NewServiceError(httperr.StatusBadRequest, fmt.Sprintf("invalid %s format", name))
	}

	return parsed, nil
}

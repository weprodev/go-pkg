package uuidutil_test

import (
	"database/sql/driver"
	"testing"

	"github.com/google/uuid"
	"github.com/weprodev/go-pkg/uuidutil"
)

// --- utils.go tests ---

func TestUUIDValueOrNil(t *testing.T) {
	if got := uuidutil.UUIDValueOrNil(nil); got != uuid.Nil {
		t.Errorf("got %v, want uuid.Nil", got)
	}
	id := uuid.New()
	if got := uuidutil.UUIDValueOrNil(&id); got != id {
		t.Errorf("got %v, want %v", got, id)
	}
}

func TestUUIDPointerOrNil(t *testing.T) {
	if got := uuidutil.UUIDPointerOrNil(uuid.Nil); got != nil {
		t.Errorf("got %v, want nil", got)
	}
	id := uuid.New()
	if got := uuidutil.UUIDPointerOrNil(id); got == nil || *got != id {
		t.Errorf("got %v, want pointer to %v", got, id)
	}
}

func TestValidateAndParseUUID(t *testing.T) {
	validStr := uuid.New().String()

	parsed, err := uuidutil.ValidateAndParseUUID(validStr, "MyID")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if parsed.String() != validStr {
		t.Errorf("got %s, want %s", parsed, validStr)
	}

	_, err = uuidutil.ValidateAndParseUUID("", "")
	if err == nil || err.Error() != "missing ID" {
		t.Errorf("expected missing ID error, got %v", err)
	}

	_, err = uuidutil.ValidateAndParseUUID("not-a-uuid", "UserID")
	if err == nil || err.Error() != "invalid UserID format" {
		t.Errorf("expected invalid format error, got %v", err)
	}
}

// --- id.go tests ---

func TestID_String_IsNil(t *testing.T) {
	zero := uuidutil.ID{}
	if !zero.IsNil() {
		t.Error("expected zero ID to be nil")
	}

	id := uuidutil.NewID()
	if id.IsNil() {
		t.Error("expected new ID to not be nil")
	}
	if len(id.String()) != 36 {
		t.Errorf("expected string length 36, got %d", len(id.String()))
	}
}

func TestID_Parse(t *testing.T) {
	idStr := uuid.New().String()

	parsed, err := uuidutil.ParseID(idStr)
	if err != nil {
		t.Errorf("unexpected parsing error: %v", err)
	}
	if parsed.String() != idStr {
		t.Errorf("got %s, want %s", parsed, idStr)
	}

	_, err = uuidutil.ParseID("invalid")
	if err == nil {
		t.Error("expected error parsing invalid ID")
	}

	mustParsed := uuidutil.MustParseID(idStr)
	if mustParsed.String() != idStr {
		t.Errorf("got %s, want %s", mustParsed, idStr)
	}
}

func TestID_MustParseID_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid MustParseID call")
		}
	}()
	uuidutil.MustParseID("invalid")
}

func TestID_Value_Scan(t *testing.T) {
	id := uuidutil.NewID()

	// Valuer
	val, err := id.Value()
	if err != nil {
		t.Errorf("unexpected Valuer error: %v", err)
	}
	strVal, ok := val.(string)
	if !ok || strVal != id.String() {
		t.Errorf("got %v, want %s", val, id.String())
	}

	// Scanner
	var scanned uuidutil.ID

	// string scan
	err = scanned.Scan(id.String())
	if err != nil {
		t.Errorf("unexpected Scan error: %v", err)
	}
	if scanned != id {
		t.Errorf("got %v, want %v", scanned, id)
	}

	// bytes scan
	err = scanned.Scan([]byte(id.String()))
	if err != nil {
		t.Errorf("unexpected Scan error: %v", err)
	}
	if scanned != id {
		t.Errorf("got %v, want %v", scanned, id)
	}

	// errors
	err = scanned.Scan("not-uuid")
	if err == nil {
		t.Error("expected error scanning invalid string")
	}
	err = scanned.Scan([]byte("not-uuid"))
	if err == nil {
		t.Error("expected error scanning invalid bytes")
	}
	err = scanned.Scan(123)
	if err == nil {
		t.Error("expected error scanning invalid type")
	}

	// Make sure we satisfied Valuer and Scanner interfaces
	var _ driver.Valuer = id
	var _ = (&id).Scan(nil)
}

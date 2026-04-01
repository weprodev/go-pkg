package validator_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/weprodev/go-pkg/httperr"
	myvalidator "github.com/weprodev/go-pkg/validator"
	echoValidator "github.com/weprodev/go-pkg/validator/echo"
)

type TestDTO struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}

type NestedDTO struct {
	Items []TestDTO `json:"items" validate:"required,dive"`
}

type UnknownFieldsDTO struct {
	ID string `json:"id"`
}

// Implements ValidatableDTO
func (d TestDTO) Validate() error {
	v := myvalidator.NewValidator()
	return v.Validate(d)
}

// Implements StrictValidatableDTO
func (d *TestDTO) ValidateAndUnmarshal(data []byte) error {
	return myvalidator.StrictUnmarshalAndValidate(data, d, func() error {
		return d.Validate()
	})
}

func TestValidator_Validate(t *testing.T) {
	v := myvalidator.NewValidator()

	dto := TestDTO{Name: ""}
	err := v.Validate(dto)
	if err == nil {
		t.Fatal("expected validation error")
	}

	se := httperr.AsServiceError(err)
	if se == nil {
		t.Fatal("expected ServiceError")
	}
	if se.Code != 422 {
		t.Errorf("expected 422 Unprocessable Entity, got %d", se.Code)
	}

	dto = TestDTO{Name: "John"}
	err = v.Validate(dto)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidator_RegisterValidation(t *testing.T) {
	v := myvalidator.NewValidator()

	err := v.RegisterValidation("is_awesome", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "awesome"
	})
	if err != nil {
		t.Fatalf("unexpected error registering validation: %v", err)
	}

	type Awesomeness struct {
		Trait string `json:"trait" validate:"required,is_awesome"`
	}

	if err := v.Validate(Awesomeness{Trait: "boring"}); err == nil {
		t.Fatal("expected validation error for 'boring'")
	}
	if err := v.Validate(Awesomeness{Trait: "awesome"}); err != nil {
		t.Fatal("expected no error for 'awesome'")
	}
}

func TestStrictUnmarshalAndValidate(t *testing.T) {
	t.Run("Valid JSON", func(t *testing.T) {
		data := []byte(`{"name":"Jane", "email":"jane@example.com"}`)
		var dto TestDTO
		err := myvalidator.StrictUnmarshalValidatable(data, &dto)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Unknown Field", func(t *testing.T) {
		data := []byte(`{"id":"123", "unknown":"field"}`)
		var dto UnknownFieldsDTO
		err := myvalidator.StrictUnmarshalAndValidate(data, &dto, nil)
		if err == nil {
			t.Fatal("expected error for unknown field")
		}
		se := httperr.AsServiceError(err)
		if se.Message != "unknown field in request" {
			t.Errorf("expected unknown field message, got %s", se.Message)
		}
	})

	t.Run("Syntax Error", func(t *testing.T) {
		data := []byte(`{"id":"123"`)
		var dto UnknownFieldsDTO
		err := myvalidator.StrictUnmarshalAndValidate(data, &dto, nil)
		if err == nil {
			t.Fatal("expected error for syntax err")
		}
		se := httperr.AsServiceError(err)
		// Usually json Decoder emits "unexpected EOF" which falls into "invalid JSON format"
		if se.Message != "invalid JSON syntax" && se.Message != "invalid JSON format" {
			t.Errorf("expected syntax error message, got %q", se.Message)
		}
	})

	t.Run("Type Error", func(t *testing.T) {
		data := []byte(`{"name": 123}`)
		var dto TestDTO
		err := myvalidator.StrictUnmarshalAndValidate(data, &dto, nil)
		if err == nil {
			t.Fatal("expected error for type err")
		}
		se := httperr.AsServiceError(err)
		if se.Message != "invalid field type" {
			t.Errorf("expected invalid field type message, got %q", se.Message)
		}
	})
}

func TestBindAndValidateStrict(t *testing.T) {
	e := echo.New()

	// Valid request
	body := bytes.NewBufferString(`{"name":"Jane"}`)
	req := httptest.NewRequest("POST", "/", body)
	c := e.NewContext(req, httptest.NewRecorder())

	var dto TestDTO
	err := echoValidator.BindAndValidateStrict(c, &dto)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Invalid body
	body = bytes.NewBufferString(`{"name":"Jane", "unknown": true}`)
	req = httptest.NewRequest("POST", "/", body)
	c = e.NewContext(req, httptest.NewRecorder())

	err = echoValidator.BindAndValidateStrict(c, &dto)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

// FieldErrorToString logic validation
func TestFieldErrorToString(t *testing.T) {
	// Directly test the unexported FieldErrorToString or test it via validate output
	v := myvalidator.NewValidator()
	err := v.Validate(TestDTO{Name: ""})

	se := httperr.AsServiceError(err)
	if len(se.Errors) == 0 {
		t.Fatal("expected error messages")
	}

	expectedMsgs := []string{
		"name is required",
	}

	for _, expected := range expectedMsgs {
		found := false
		for _, got := range se.Errors {
			if got == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("did not find expected message %q in %v", expected, se.Errors)
		}
	}
}

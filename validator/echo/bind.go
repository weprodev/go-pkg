package echo

import (
	"io"

	"github.com/labstack/echo/v4"

	"github.com/weprodev/go-pkg/httperr"
	"github.com/weprodev/go-pkg/validator"
)

// BindAndValidateStrict reads the request body and invokes dto.ValidateAndUnmarshal
// (which should perform strict JSON decoding and validation).
func BindAndValidateStrict(c echo.Context, dto validator.StrictValidatableDTO) error {
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return httperr.ErrInvalidRequestBody
	}
	return dto.ValidateAndUnmarshal(bodyBytes)
}

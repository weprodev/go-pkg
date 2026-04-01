// Package api provides HTTP framework utilities for building HTTP APIs.
//
// It includes route registration helpers, a JSON serializer that preserves
// raw HTML in responses, a generic ErrorHandler that maps ServiceError values
// to structured JSON responses, and a SanitizeBody helper for XSS protection.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/labstack/echo/v4"

	"github.com/weprodev/go-pkg/httperr"
	"github.com/weprodev/go-pkg/sanitizer"
)

// EchoRouteConfig describes a single Echo route with optional middleware.
type EchoRouteConfig struct {
	Method     string
	Path       string
	Handler    echo.HandlerFunc
	Middleware []echo.MiddlewareFunc
}

// RegisterEchoRoutes registers a slice of EchoRouteConfig under the given group prefix.
func RegisterEchoRoutes(e *echo.Echo, groupPrefix string, routes []EchoRouteConfig) {
	group := e.Group(groupPrefix)
	for _, r := range routes {
		group.Add(r.Method, r.Path, r.Handler, r.Middleware...)
	}
}

// EchoErrorHandler processes errors in Echo handlers and returns structured JSON.
// It recognises httperr.ServiceError and maps unknown errors to 500.
type EchoErrorHandler struct {
	Logger *slog.Logger
}

// NewEchoErrorHandler creates a new EchoErrorHandler with the given logger.
func NewEchoErrorHandler(logger *slog.Logger) *EchoErrorHandler {
	return &EchoErrorHandler{Logger: logger}
}

// HandleError writes a JSON error response. If err is a ServiceError its code
// and message are used directly; otherwise a generic 500 is returned and the
// original error is logged.
func (h *EchoErrorHandler) HandleError(c echo.Context, err error, operation string) error {
	var se httperr.ServiceError
	if errors.As(err, &se) {
		return c.JSON(se.Code, se)
	}

	// Unknown error — log it and return 500.
	if h.Logger != nil {
		h.Logger.Error("unhandled error",
			"error", err,
			"operation", operation,
			"path", c.Request().URL.Path,
			"method", c.Request().Method,
		)
	}
	return c.JSON(httperr.StatusInternalServer, httperr.ErrInternalServer)
}

// EchoHandler adapts EchoErrorHandler to echo.HTTPErrorHandler for global registration.
// It uses "http_request" as the default operation context.
func (h *EchoErrorHandler) EchoHandler(err error, c echo.Context) {
	// If the response was already committed, we shouldn't send another JSON
	if c.Response().Committed {
		return
	}
	// Echo's default error maps like 404 and 405 are returned as *echo.HTTPError
	var he *echo.HTTPError
	if errors.As(err, &he) {
		_ = c.JSON(he.Code, httperr.NewServiceError(he.Code, fmt.Sprintf("%v", he.Message)))
		return
	}

	_ = h.HandleError(c, err, "http_request")
}

// EchoUnescapedHTMLJSONSerializer is a JSON serializer that disables HTML escaping.
// Use this when returning pre-sanitized HTML content in JSON responses.
type EchoUnescapedHTMLJSONSerializer struct{}

// Serialize encodes the value as JSON without HTML escaping.
func (s *EchoUnescapedHTMLJSONSerializer) Serialize(ctx echo.Context, i interface{}, indent string) error {
	enc := json.NewEncoder(ctx.Response())
	enc.SetEscapeHTML(false)
	if indent != "" {
		enc.SetIndent("", indent)
	}
	return enc.Encode(i)
}

// Deserialize decodes JSON from the request body.
func (s *EchoUnescapedHTMLJSONSerializer) Deserialize(ctx echo.Context, i interface{}) error {
	return json.NewDecoder(ctx.Request().Body).Decode(i)
}

// EchoSanitizeBody reads the entire request body as a string, sanitizes it using
// the default HTML sanitizer, and returns the clean result.
//
// Use this in handlers that receive raw HTML from clients (e.g. rich-text
// editors) to strip XSS before further processing.
//
// Example:
//
//	func myHandler(c echo.Context) error {
//	    clean, err := api.EchoSanitizeBody(c)
//	    if err != nil { return err }
//	    // use clean
//	}
func EchoSanitizeBody(c echo.Context) (string, error) {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return "", httperr.NewServiceError(httperr.StatusBadRequest, "failed to read request body")
	}
	return sanitizer.SanitizeHTML(string(body)), nil
}

// EchoSanitizeBodyWith reads the request body and sanitizes it using the provided
// Sanitizer, allowing callers to apply a custom sanitization policy.
func EchoSanitizeBodyWith(c echo.Context, s *sanitizer.Sanitizer) (string, error) {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return "", httperr.NewServiceError(httperr.StatusBadRequest, "failed to read request body")
	}
	return s.Sanitize(string(body)), nil
}

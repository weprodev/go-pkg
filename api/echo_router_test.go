package api_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/weprodev/go-pkg/api"
	"github.com/weprodev/go-pkg/httperr"
	"github.com/weprodev/go-pkg/sanitizer"
)

func TestRegisterRoutes(t *testing.T) {
	e := echo.New()
	handler := func(c echo.Context) error { return c.String(200, "ok") }

	routes := []api.EchoRouteConfig{
		{Method: http.MethodGet, Path: "/test", Handler: handler},
	}

	api.RegisterEchoRoutes(e, "/api", routes)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != 200 || rec.Body.String() != "ok" {
		t.Errorf("expected 200 ok, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestErrorHandler(t *testing.T) {
	e := echo.New()
	handler := api.NewEchoErrorHandler(nil) // nil logger is allowed

	t.Run("ServiceError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := httperr.NewServiceError(400, "bad input")
		if hErr := handler.HandleError(c, err, "test"); hErr != nil {
			t.Errorf("HandleError shouldn't return error on success, got: %v", hErr)
		}

		if rec.Code != 400 {
			t.Errorf("expected 400, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "bad input") {
			t.Errorf("expected body to contain bad input, got %s", rec.Body.String())
		}
	})

	t.Run("UnknownError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := errors.New("something explosive")
		if hErr := handler.HandleError(c, err, "test"); hErr != nil {
			t.Errorf("HandleError shouldn't return error on success, got: %v", hErr)
		}

		if rec.Code != 500 {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("EchoHandler Adapter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Test that basic errors fall back to 500
		err := errors.New("fallback error")
		handler.EchoHandler(err, c)

		if rec.Code != 500 {
			t.Errorf("expected 500, got %d", rec.Code)
		}
	})

	t.Run("EchoHandler Native HTTPError", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Test that native echo errors map accurately
		err := echo.NewHTTPError(http.StatusNotFound, "Not Found Route")
		handler.EchoHandler(err, c)

		if rec.Code != 404 {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestSanitizeBody(t *testing.T) {
	e := echo.New()

	body := bytes.NewBufferString(`<b>hello</b><script>alert(1)</script>`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	clean, err := api.EchoSanitizeBody(c)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if clean != `<b>hello</b>` {
		t.Errorf("expected <b>hello</b>, got %q", clean)
	}
}

func TestSanitizeBodyWith(t *testing.T) {
	e := echo.New()
	strict := sanitizer.NewStrict()

	body := bytes.NewBufferString(`<b>hello</b>`)
	req := httptest.NewRequest(http.MethodPost, "/", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	clean, err := api.EchoSanitizeBodyWith(c, strict)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if clean != `hello` {
		t.Errorf("expected hello, got %q", clean)
	}
}

func TestUnescapedHTMLJSONSerializer(t *testing.T) {
	e := echo.New()
	serializer := &api.EchoUnescapedHTMLJSONSerializer{}

	// Test Serialize
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	c := e.NewContext(req, rec)

	data := map[string]string{"html": "<b>Hello</b>"}
	if err := serializer.Serialize(c, data, ""); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	// Ensure < and > are not escaped to \u003c and \u003e
	expected := `{"html":"<b>Hello</b>"}` + "\n"
	if rec.Body.String() != expected {
		t.Errorf("got %q, want %q", rec.Body.String(), expected)
	}

	// Test Deserialize
	body := bytes.NewBufferString(`{"html":"<b>Hello</b>"}`)
	req2 := httptest.NewRequest(http.MethodPost, "/", body)
	c2 := e.NewContext(req2, httptest.NewRecorder())

	var out map[string]string
	if err := serializer.Deserialize(c2, &out); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out["html"] != "<b>Hello</b>" {
		t.Errorf("got %q, want %q", out["html"], "<b>Hello</b>")
	}
}

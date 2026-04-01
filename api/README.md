# api

The `api` package provides generic tools to process HTTP requests. 
Currently, it holds `github.com/labstack/echo/v4` specific implementations (`echo_router.go`) but can be easily extended with bindings for `gin`, `fiber` or `chi` by introducing new feature files like `gin_router.go`.

## Usage Example (Echo)

```go
package main

import (
    "log/slog"
    "github.com/labstack/echo/v4"
    "github.com/weprodev/go-pkg/api"
    "github.com/weprodev/go-pkg/httperr"
)

func main() {
    e := echo.New()
    
    // Set our standard error handler
    errorHandler := api.NewEchoErrorHandler(slog.Default())
    e.HTTPErrorHandler = func(err error, c echo.Context) {
        errorHandler.HandleError(c, err, "request")
    }

    // Register routes safely
    routes := []api.EchoRouteConfig{
        {
            Method: "POST",
            Path:   "/bio",
            Handler: func(c echo.Context) error {
                // Automatically sanitizes XSS inside the body!
                cleanHTML, err := api.EchoSanitizeBody(c)
                if err != nil {
                    return err
                }
                return c.String(200, cleanHTML)
            },
        },
    }
    
    api.RegisterEchoRoutes(e, "/v1", routes)
    e.Start(":8080")
}
```

# httperr

The `httperr` package provides structured HTTP error types that can be returned from any domain or handler logic. It is purely dependent on the standard library.

## Usage Example

```go
package main

import (
	"errors"
	"fmt"
	"github.com/weprodev/go-pkg/httperr"
)

func DoSomething() error {
	// Create an error with HTTP context
	return httperr.NewServiceError(400, "Invalid Request", "name is required", "email is invalid")
}

func main() {
	err := DoSomething()
	
	// Safely extract the ServiceError type
	if se := httperr.AsServiceError(err); se != nil {
		fmt.Printf("HTTP Code: %d, Message: %s\n", se.Code, se.Message)
		// HTTP Code: 400, Message: Invalid Request
	}
	
	// Error comparison works standard
	if errors.Is(err, httperr.ErrBadRequest) {
		fmt.Println("It was a bad request!")
	}
}
```

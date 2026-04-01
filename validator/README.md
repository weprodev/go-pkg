# validator

Provides binding to JSON tags seamlessly on top of `go-playground/validator`. Supports strict unmarshalling configurations.

## Usage Example

```go
package main

import (
	"fmt"
	"github.com/weprodev/go-pkg/validator"
)

type MyDTO struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"gte=18"`
}

// Attach a receiver func to implement ValidatableDTO
func (d MyDTO) Validate() error {
	v := validator.NewValidator()
	return v.Validate(d)
}

func main() {
	payload := []byte(`{"name": "", "age": 16, "unknown_field": "error!"}`)
	var dto MyDTO
	
	// This will enforce:
	// 1. JSON has exactly the expected fields (fails due to unknown_field)
	// 2. Validation constraints apply (`name` is required, `age` is gte 18)
	err := validator.StrictUnmarshalValidatable(payload, dto)
	if err != nil {
		fmt.Println("Error:", err.Error())
	}
}
```

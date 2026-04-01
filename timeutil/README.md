# timeutil

The `timeutil` package provides simple semantic wrappers for timestamps.

## Usage Example

```go
package main

import (
	"fmt"
	"time"
	"github.com/weprodev/go-pkg/timeutil"
)

func main() {
	now := timeutil.Now()
	fmt.Println("Current Time:", now)

	// Avoid ugly nil checks in DTO conversions:
	ptr := timeutil.ToPointer(now)
	fmt.Println("Is pointer nil?", ptr == nil)
}
```

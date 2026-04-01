# uuidutil

The `uuidutil` package contains an ID entity mapped to `database/sql` driver standards in addition to standard HTTP parameter parsing utilities.

## Usage Example

```go
package main

import (
	"fmt"
	"github.com/weprodev/go-pkg/uuidutil"
)

func main() {
	// Generates a fully DB compatible ID 
	// (satisfies driver.Valuer and sql.Scanner)
	id := uuidutil.NewID()
	fmt.Println("New ID:", id.String())

	// Safe API parsing - Returns a Service Error (400) if invalid
	parsedID, err := uuidutil.ValidateAndParseUUID(id.String(), "user_id")
	if err != nil {
		panic(err)
	}

	fmt.Println("Validated UUID:", parsedID)
}
```

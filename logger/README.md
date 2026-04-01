# logger

The `logger` package wraps `log/slog` to provide a generic JSON or Text logger. It natively supports context-based key extraction (e.g., getting a `trace_id` from the context) and fanning out to multiple external logging destinations.

## Usage Example

```go
package main

import (
	"context"
	"github.com/weprodev/go-pkg/logger"
)

func main() {
	// 1. Define how to extract elements from context
	traceExtractor := func(ctx context.Context) []any {
		if traceID, ok := ctx.Value("trace_id").(string); ok {
			return []any{"trace_id", traceID}
		}
		return nil
	}

	// 2. Configure Logger
	cfg := logger.Config{
		Level:             logger.LevelInfo,
		Format:            logger.FormatJSON,
		ContextExtractors: []logger.ContextExtractor{traceExtractor},
	}
	
	l, err := logger.New(cfg)
	if err != nil {
		panic(err)
	}

	// 3. Log with context
	ctx := context.WithValue(context.Background(), "trace_id", "abc-12345")
	
	l.WithContext(ctx).Info("Request started", "method", "GET")
	// {"level":"INFO","msg":"Request started","method":"GET","trace_id":"abc-12345"}
}
```

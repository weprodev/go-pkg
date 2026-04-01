# config

The `config` package provides generic utilities to load configurations from `.yaml` or `.json` files, merging them with environment variables. It has zero domain-specific logic.

## Usage Example

```go
package main

import (
	"fmt"
	"github.com/weprodev/go-pkg/config"
)

type MyServiceConfig struct {
	config.Config // Embeds standard Server, Logging, DB, etc.
	MyCustomKey   string `yaml:"custom_key"`
}

func main() {
	var cfg MyServiceConfig
	
	// Load structure from custom YAML
	if err := config.LoadConfigFile("config.yml", &cfg); err != nil {
		panic(err)
	}

	// Important: Merge environment overrides for the standard embedded struct!
	config.ApplyEnvironmentOverrides(&cfg.Config)

	// Validate universal defaults
	if errs := config.ValidateConfig(&cfg.Config); len(errs) > 0 {
		panic(fmt.Sprintf("Invalid Config: %v", errs))
	}

	fmt.Printf("Started on port %d", cfg.Server.Port)
}
```

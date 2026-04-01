# sanitizer

The `sanitizer` package guarantees safe XSS checking. By default, it preserves basic structure elements (`<p>`, `<b>`, `<mark>`, etc.) while eliminating scripts and bad CSS properties.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/weprodev/go-pkg/sanitizer"
)

func main() {
    input := `<b>Hello</b> <script>alert(1)</script>`
    
    // 1. Maintain formatting safely
    safeHTML := sanitizer.SanitizeHTML(input)
    fmt.Println(safeHTML) // <b>Hello</b> 
    
    // 2. Erase everything entirely
    plainText := sanitizer.StripAllHTML(input)
    fmt.Println(plainText) // Hello
    
    // 3. Customize policy rules
    s := sanitizer.New(
        sanitizer.WithElements("a"),
        sanitizer.WithAttrs("href").OnElements("a"),
    )
    customSafe := s.Sanitize(`<a href="https://example.com" onclick="steal()">Click</a>`)
    fmt.Println(customSafe) // <a href="https://example.com">Click</a>
}
```

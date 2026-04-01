# crypto

The `crypto` package abstracts common cryptographic operations used frequently in web services like structured AES-256-GCM authenticated encryption and BCrypt credential hashing.

## Usage Example

### AEAD Secret Cryptography

```go
package main

import (
    "fmt"
    "github.com/weprodev/go-pkg/crypto"
)

func main() {
    key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes required
    
    svc, err := crypto.NewAESService(key)
    if err != nil {
        panic(err)
    }

    raw := []byte("highly sensitive internal data")
    
    cipher, _ := svc.Encrypt(raw)
    fmt.Printf("Encrypted safely: %x\n", cipher)

    // And reverse to use it...
    plain, _ := svc.Decrypt(cipher)
    fmt.Println(string(plain))
}
```

### Credential Hashing

```go
package main

import (
    "fmt"
    "github.com/weprodev/go-pkg/crypto"
)

func main() {
    password := "my_admin_password!"
    
    // Automatic bcrypt salt & hash (Cost: 14)
    hash, _ := crypto.HashSecret(password)

    if crypto.CheckSecretHash("wrong", hash) {
        fmt.Println("This string is never printed.")
    }

    if crypto.CheckSecretHash(password, hash) {
        fmt.Println("Success!")
    }
}
```

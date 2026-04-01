# go-pkg: WeProDev Standard Library

`go-pkg` is a zero-business-logic, highly modular set of utilities for Go backend services. It is designed to be fully open-source compatible, robust, and extensible.

## 📦 Packages

Each package is designed to be tightly scoped with minimal dependencies across packages.

- **[config](./config/)**: Environment and File-based generic configuration loading.
- **[api](./api/)**: Helper structs and utilities for the [Echo web framework](https://echo.labstack.com) (extensible to other frameworks like Chi/Gin).
- **[crypto](./crypto/)**: Generic AES-256-GCM authenticated encryption and BCrypt credential hashing.
- **[httperr](./httperr/)**: Structured and actionable error types for web APIs (`ServiceError`).
- **[logger](./logger/)**: Generic JSON/text logger wrapping `log/slog` with Context extractors and multi-handler fan out.
- **[pgsql](./pgsql/)**: Configurable PostgreSQL connection helpers with transaction context wrappers.
- **[sanitizer](./sanitizer/)**: Extensible, configuration-driven HTML sanitization (prevents XSS).
- **[timeutil](./timeutil/)**: Timestamp and pointer time utilities.
- **[uuidutil](./uuidutil/)**: Common identifier types, mapping safely to `database/sql` driver implementations.
- **[validator](./validator/)**: Framework-agnostic validation helpers over `go-playground/validator` with strict JSON decoding utilities.
- **[validator/echo](./validator/echo/)**: Echo integration helpers (binding + strict validation) without coupling the core `validator` package to Echo.

## 🛠 Usage & Audit

Validate the library with:
```bash
make audit
```

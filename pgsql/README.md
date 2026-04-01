# pgsql

The `pgsql` package provides a configured `*sql.DB` wrapper utilizing standard driver conventions (`github.com/lib/pq`). It makes managing transactions across multiple repositories explicit but simple.

## Usage Example

```go
package repository

import (
	"context"
	"github.com/weprodev/go-pkg/pgsql"
)

type UserRepo struct {
	client *pgsql.PgClient
}

func (r *UserRepo) CreateUser(ctx context.Context, name string) error {
	// `GetDB` returns a `*sql.Tx` if in a transaction block, else `*sql.DB` 
	db := r.client.GetDB(ctx)
	
	_, err := db.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", name)
	return err
}

func (r *UserRepo) CreateUserWithProfile(ctx context.Context, name, bio string) error {
	// Using RunInTransaction handles begin, commit, and rollback logic
	return r.client.RunInTransaction(ctx, func(txCtx context.Context) error {
		// Both of these calls will now share the same atomic transaction
		if err := r.CreateUser(txCtx, name); err != nil {
			return err
		}
		if err := r.createProfile(txCtx, bio); err != nil {
			return err
		}
		return nil
	})
}
```

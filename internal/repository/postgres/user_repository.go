package postgres

import (
	"context"
)

// UserRepository wraps the Store interface
type UserRepository struct {
	store Store
}

// NewUserRepository initializes UserRepository
func NewUserRepository(store Store) *UserRepository {
	return &UserRepository{
		store: store,
	}
}

// CreateUser executes a single query using the embedded Querier
func (r *UserRepository) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
	return r.store.CreateUser(ctx, arg)
}

// GetUsers retrieves all users
func (r *UserRepository) GetUsers(ctx context.Context) ([]GetUsersRow, error) {
	return r.store.GetUsers(ctx)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	return r.store.GetUserByEmail(ctx, email)
}

// ExampleTx demonstrates executing multiple queries within a single transaction
func (r *UserRepository) ExampleTx(ctx context.Context, arg CreateUserParams) error {
	return r.store.ExecTx(ctx, func(q Querier) error {
		_, err := q.CreateUser(ctx, arg)
		if err != nil {
			return err
		}

		// Call other methods on 'q' here to keep them in the same transaction
		// _, err = q.UpdateSomething(ctx, ...)

		return nil
	})
}

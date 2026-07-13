package postgres

import (
	"context"

	"github.com/google/uuid"
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
func (r *UserRepository) GetUsers(ctx context.Context, arg GetUsersParams) ([]GetUsersRow, error) {
	return r.store.GetUsers(ctx, arg)
}

func (r *UserRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.store.CountUsers(ctx)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	return r.store.GetUserByEmail(ctx, email)
}

func (r *UserRepository) CreateAdmin(ctx context.Context, arg CreateAdminParams) (CreateAdminRow, error) {
	return r.store.CreateAdmin(ctx, arg)
}

func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (GetUserByIDRow, error) {
	return r.store.GetUserByID(ctx, id)
}

func (r *UserRepository) UpdateUser(ctx context.Context, arg UpdateUserParams) (UpdateUserRow, error) {
	return r.store.UpdateUser(ctx, arg)
}

func (r UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.store.DeleteUser(ctx, id)
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

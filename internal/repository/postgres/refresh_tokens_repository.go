package postgres

import "context"

type RefreshTokensRepository struct {
	store Store
}

func NewRefreshTokensRepository(store Store) *RefreshTokensRepository {
	return &RefreshTokensRepository{store: store}
}

func (r *RefreshTokensRepository) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	return r.store.CreateRefreshToken(ctx, arg)
}

func (r *RefreshTokensRepository) GetRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	return r.store.GetRefreshToken(ctx, token)
}

func (r *RefreshTokensRepository) RevokeToken(ctx context.Context, token string) error {
	return r.store.RevokeToken(ctx, token)
}

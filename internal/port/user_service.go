package port

import (
	"context"

	"github.com/meteoradev/BlogApi/internal/domain"
)

type UserService interface {
	Create(ctx context.Context, username, email, password string) (*domain.User, error)
	GetByID(ctx context.Context, userID int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, currUserID, userID int64, username, email, password string) error
	Delete(ctx context.Context, currUserID, userID int64) error
}

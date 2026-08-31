package port

import (
	"context"

	"github.com/meteoradev/BlogApi/internal/domain"
)

type UserCreater interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
}

type UserReaderByID interface {
	GetByID(ctx context.Context, userID int64) (*domain.User, error)
}
type UserReaderByEmail interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type UserUpdater interface {
	Update(ctx context.Context, user *domain.User) error
}
type UserDeleter interface {
	Delete(ctx context.Context, id int64) error
}

type Hasher interface {
	Hash(password string) (string, error)
	Check(hashed, plain string) error
}

type UserRepository interface {
	UserReaderByID
	UserReaderByEmail
	UserCreater
	UserUpdater
	UserDeleter
}

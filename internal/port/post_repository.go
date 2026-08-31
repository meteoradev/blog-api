package port

import (
	"context"

	"github.com/meteoradev/BlogApi/internal/domain"
)

type PostCreater interface {
	Create(ctx context.Context, post *domain.Post) (*domain.Post, error)
}
type PostReaderByID interface {
	GetByID(ctx context.Context, ID int64) (*domain.Post, error)
}
type PostReaderByTitle interface {
	GetByTitle(ctx context.Context, title string) (*domain.Post, error)
}

type PostUpdater interface {
	Update(ctx context.Context, post *domain.Post) error
}

type PostUpdaterWithValidation interface {
	UpdateWithValidate(ctx context.Context, currUserID int64, post *domain.Post) error
}

type PostDeleter interface {
	Delete(ctx context.Context, ID int64) (string, error)
}
type PostDeleterWithValidate interface {
	DeleteWithValidate(ctx context.Context, currUserID, ID int64) (string, error)
}

type PostRepository interface {
	PostReaderByID
	PostReaderByTitle
	PostCreater
	PostUpdater
	PostDeleter
	PostUpdaterWithValidation
	PostDeleterWithValidate
}

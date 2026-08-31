package port

import (
	"context"

	"github.com/meteoradev/BlogApi/internal/domain"
)

type PostService interface {
	Create(ctx context.Context, userID int64, title, content string) (*domain.Post, error)
	GetByID(ctx context.Context, postID int64) (*domain.Post, error)
	GetByTitle(ctx context.Context, title string) (*domain.Post, error)
	Update(ctx context.Context, postID int64, title, content string) error
	UpdateWithValidate(ctx context.Context, currUserID, postID int64, title, content string) error
	Delete(ctx context.Context, postID int64) error
	DeleteWithValidate(ctx context.Context, currUserID, postID int64) error
}

package port

import (
	"context"

	"github.com/meteoradev/BlogApi/internal/domain"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (*domain.AuthResult, error)
}

type TokenProvider interface {
	GenerateToken(_ context.Context, userID int64, email string) (string, error)
	ValidateToken(tokenString string) (*domain.TokenPayload, error)
	// RefreshToken(oldToken string) (string, error)
}

package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/service"
)

type mockTokenProvider struct {
	generateTokenFunc func(ctx context.Context, userID int64, email string) (string, error)
	validateTokenFunc func(token string) (*domain.TokenPayload, error)
}

func (m *mockTokenProvider) GenerateToken(ctx context.Context, userID int64, email string) (string, error) {
	return m.generateTokenFunc(ctx, userID, email)
}

func (m *mockTokenProvider) ValidateToken(token string) (*domain.TokenPayload, error) {
	return m.validateTokenFunc(token)
}

func TestAuthService_Login(t *testing.T) {
	now := time.Now()
	validToken := "valid_jwt_token"
	tokenPayload := &domain.TokenPayload{
		Claims: domain.Claims{
			UserID: 1,
			Email:  "test@mail.com",
		},
		ExpireAt: now.Add(24 * time.Hour),
	}

	mockRepo := &mockUserRepo{
		getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:           1,
				Username:     "testuser",
				Email:        "test@mail.com",
				PasswordHash: "hashed_password",
			}, nil
		},
	}
	hasher := &mockHasher{
		checkFunc: func(hashed, plain string) error {
			return nil
		},
	}
	tokenProvider := &mockTokenProvider{
		generateTokenFunc: func(ctx context.Context, userID int64, email string) (string, error) {
			return validToken, nil
		},
		validateTokenFunc: func(token string) (*domain.TokenPayload, error) {
			return tokenPayload, nil
		},
	}

	svc := service.NewAuthService(mockRepo, hasher, tokenProvider)

	t.Run("happy path", func(t *testing.T) {
		result, err := svc.Login(context.Background(), "test@mail.com", "password123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Token != validToken {
			t.Errorf("got token %s, want %s", result.Token, validToken)
		}
		if result.TokenPayload == nil {
			t.Fatalf("expected TokenPayload, got nil")
		}
		if result.TokenPayload.Claims.UserID != 1 {
			t.Errorf("got userID %d, want 1", result.TokenPayload.Claims.UserID)
		}
		if result.TokenPayload.Claims.Email != "test@mail.com" {
			t.Errorf("got email %s, want test@mail.com", result.TokenPayload.Claims.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, service.ErrUserNotFound
			},
		}
		h := &mockHasher{checkFunc: func(hashed, plain string) error { return nil }}
		tp := &mockTokenProvider{
			generateTokenFunc: func(ctx context.Context, userID int64, email string) (string, error) {
				return validToken, nil
			},
			validateTokenFunc: func(token string) (*domain.TokenPayload, error) {
				return tokenPayload, nil
			},
		}
		svc := service.NewAuthService(repo, h, tp)

		_, err := svc.Login(context.Background(), "notfound@mail.com", "password123")
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Errorf("got error %v, want %v", err, service.ErrUserNotFound)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@mail.com",
					PasswordHash: "hashed_password",
				}, nil
			},
		}
		hasher := &mockHasher{
			checkFunc: func(hashed, plain string) error {
				return errors.New("wrong password")
			},
		}
		token := &mockTokenProvider{
			generateTokenFunc: func(ctx context.Context, userID int64, email string) (string, error) {
				return validToken, nil
			},
			validateTokenFunc: func(token string) (*domain.TokenPayload, error) {
				return tokenPayload, nil
			},
		}
		svc := service.NewAuthService(repo, hasher, token)

		_, err := svc.Login(context.Background(), "test@mail.com", "wrongpassword")
		if !errors.Is(err, service.ErrWrongPassword) {
			t.Errorf("got error %v, want %v", err, service.ErrWrongPassword)
		}
	})

	t.Run("token generation failed", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@mail.com",
					PasswordHash: "hashed_password",
				}, nil
			},
		}
		hasher := &mockHasher{
			checkFunc: func(hashed, plain string) error { return nil },
		}
		token := &mockTokenProvider{
			generateTokenFunc: func(ctx context.Context, userID int64, email string) (string, error) {
				return "", service.ErrCanNotLogin
			},
			validateTokenFunc: func(token string) (*domain.TokenPayload, error) {
				return tokenPayload, nil
			},
		}
		svc := service.NewAuthService(repo, hasher, token)

		_, err := svc.Login(context.Background(), "test@mail.com", "password123")
		if !errors.Is(err, service.ErrCanNotLogin) {
			t.Errorf("got error %v, want %v", err, service.ErrCanNotLogin)
		}
	})
}

package service

import (
	"context"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/port"
	"github.com/rs/zerolog/log"
)

type AuthService struct {
	userRepo      port.UserRepository
	hasher        port.Hasher
	tokenProvider port.TokenProvider
}

func NewAuthService(
	userRepo port.UserRepository,
	hasher port.Hasher,
	tokenProvider port.TokenProvider,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		hasher:        hasher,
		tokenProvider: tokenProvider,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.AuthResult, error) {
	logger := log.Ctx(ctx)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := s.hasher.Check(user.PasswordHash, password); err != nil {
		return nil, ErrWrongPassword
	}
	token, err := s.tokenProvider.GenerateToken(ctx, user.ID, user.Email)
	if err != nil {
		return nil, ErrCanNotLogin
	}
	payload, _ := s.tokenProvider.ValidateToken(token)
	trace_id, _ := ctx.Value("trace_id").(string)
	logger.Info().
		Str("trace_id", trace_id).
		Int64("user_id", user.ID).
		Msg("User loggined")
	return domain.NewAuthResult(token, payload), nil
}

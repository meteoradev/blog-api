package jwt

import (
	"context"
	"time"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type Provider struct {
	secret  []byte
	signing jwt.SigningMethod
	expiry  time.Duration
}

func NewProvider(secret string, expiry time.Duration) *Provider {
	return &Provider{
		secret:  []byte(secret),
		signing: jwt.SigningMethodHS256,
		expiry:  expiry,
	}
}

func (p *Provider) GenerateToken(_ context.Context, userID int64, email string) (string, error) {
	claims := domain.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		}}
	token := jwt.NewWithClaims(p.signing, claims)
	return token.SignedString(p.secret)
}

func (p *Provider) ValidateToken(tokenString string) (*domain.TokenPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.Claims{}, func(t *jwt.Token) (any, error) {
		if t.Method != p.signing {
			return nil, ErrInvalidToken
		}
		return p.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*domain.Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return domain.NewPayload(claims, time.Now().Add(p.expiry)), nil
}

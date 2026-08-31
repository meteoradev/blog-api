package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/meteoradev/BlogApi/internal/domain"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		testName     string
		expErr       error
		username     string
		email        string
		passwordHash string
	}{
		// negative
		{"empty username", domain.ErrInvalidUserName, "", "test@example.com", "hashed_password"},
		{"too long username", domain.ErrInvalidUserName, strings.Repeat("username", 30), "test@example.com", "hashed_password"},
		{"empty email", domain.ErrInvalidEmail, "testuser", "", "hashed_password"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := domain.NewUser(tt.username, tt.email, tt.passwordHash)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("NewUser() negative cases got error = %e, want = %e", err, tt.expErr)
			}
		})
	}
	t.Run("happy path", func(t *testing.T) {
		user, err := domain.NewUser("testuser", "test@example.com", "hashed_password")
		if err != nil {
			t.Errorf("NewUser() positive cases got error = %e, want = %v", err, nil)
		}
		if user.Username != "testuser" || user.Email != "test@example.com" {
			t.Errorf("NewUser() positive cases got username = %s, email = %s, want = %s, %s", user.Username, user.Email, "testuser", "test@example.com")
		}
	})
}

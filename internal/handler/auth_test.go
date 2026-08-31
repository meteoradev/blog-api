package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/service"
)

type mockAuthService struct {
	mockLoginFunc func(ctx context.Context, email, password string) (*domain.AuthResult, error)
}

func (m *mockAuthService) Login(ctx context.Context, email, password string) (*domain.AuthResult, error) {
	return m.mockLoginFunc(ctx, email, password)
}

func TestAuthController_Login(t *testing.T) {
	authResult := &domain.AuthResult{
		Token: "sample.jwt.token",
		TokenPayload: &domain.TokenPayload{
			Claims: domain.Claims{
				UserID: 1,
				Email:  "user@example.com",
			},
		},
	}
	resp := AuthResponse{Token: authResult.Token, TokenPayload: authResult.TokenPayload}
	respBytes, _ := json.Marshal(resp)
	expected := string(respBytes)

	tests := []struct {
		name           string
		service        mockAuthService
		input          AuthRequest
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockAuthService{mockLoginFunc: func(ctx context.Context, email, password string) (*domain.AuthResult, error) {
				return authResult, nil
			}},
			AuthRequest{Email: "user@example.com", Password: "password123"},
			http.StatusOK,
			expected,
		},
		{
			"invalid JSON",
			mockAuthService{},
			AuthRequest{},
			http.StatusBadRequest,
			`{"error": "invalid JSON"}`,
		},
		{
			"user not found",
			mockAuthService{mockLoginFunc: func(ctx context.Context, email, password string) (*domain.AuthResult, error) {
				return nil, service.ErrUserNotFound
			}},
			AuthRequest{Email: "unknown@example.com", Password: "password123"},
			http.StatusNotFound,
			`{"error": "user not found"}`,
		},
		{
			"wrong password",
			mockAuthService{mockLoginFunc: func(ctx context.Context, email, password string) (*domain.AuthResult, error) {
				return nil, service.ErrWrongPassword
			}},
			AuthRequest{Email: "user@example.com", Password: "wrongpassword"},
			http.StatusUnauthorized,
			`{"error": "wrong password"}`,
		},
		{
			"can not login",
			mockAuthService{mockLoginFunc: func(ctx context.Context, email, password string) (*domain.AuthResult, error) {
				return nil, service.ErrCanNotLogin
			}},
			AuthRequest{Email: "user@example.com", Password: "password123"},
			http.StatusUnauthorized,
			`{"error": "can not login"}`,
		},
		{
			"unexpected error",
			mockAuthService{mockLoginFunc: func(ctx context.Context, email, password string) (*domain.AuthResult, error) {
				return nil, service.ErrUnexpected
			}},
			AuthRequest{Email: "user@example.com", Password: "password123"},
			http.StatusInternalServerError,
			`{"error": "internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewAuthController(&tt.service)

			var bodyReader *strings.Reader
			if tt.expectedBody == `{"error": "invalid JSON"}` {
				bodyReader = strings.NewReader("not valid json{")
			} else {
				bodyBytes, _ := json.Marshal(tt.input)
				bodyReader = strings.NewReader(string(bodyBytes))
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctrl.Login(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedBody != "" {
				var got, want interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if err := json.Unmarshal([]byte(tt.expectedBody), &want); err != nil {
					t.Errorf("failed to unmarshal expected: %v", err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("expected body %v, got %v", want, got)
				}
			}
		})
	}
}

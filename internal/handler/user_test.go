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
	"github.com/go-chi/chi/v5"
)

type mockUserService struct {
	mockCreateFunc     func(ctx context.Context, username, email, password string) (*domain.User, error)
	mockGetByIDFunc    func(ctx context.Context, userID int64) (*domain.User, error)
	mockGetByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	mockUpdateFunc     func(ctx context.Context, currUserID, userID int64, username, email, password string) error
	mockDeleteFunc     func(ctx context.Context, currUserID, userID int64) error
}

func (m *mockUserService) Create(ctx context.Context, username, email, password string) (*domain.User, error) {
	return m.mockCreateFunc(ctx, username, email, password)
}
func (m *mockUserService) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	return m.mockGetByIDFunc(ctx, userID)
}
func (m *mockUserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.mockGetByEmailFunc(ctx, email)
}
func (m *mockUserService) Update(ctx context.Context, currUserID, userID int64, username, email, password string) error {
	return m.mockUpdateFunc(ctx, currUserID, userID, username, email, password)
}
func (m *mockUserService) Delete(ctx context.Context, currUserID, userID int64) error {
	return m.mockDeleteFunc(ctx, currUserID, userID)
}

func TestUserController_Create(t *testing.T) {
	createdUser := userResponse{ID: 67, Username: "johndoe", Email: "user@example.com"}
	user, _ := json.Marshal(createdUser)
	resp := string(user)
	tests := []struct {
		name           string
		service        mockUserService
		input          createUserRequest
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return &domain.User{ID: 67, Username: username, Email: email, PasswordHash: password}, nil
			}},
			createUserRequest{Username: "johndoe", Email: "user@example.com", Password: "password123"},
			http.StatusCreated,
			resp,
		},
		{
			"empty username",
			mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return nil, domain.ErrInvalidUserName
			}},
			createUserRequest{Username: "", Email: "user@example.com", Password: "password123"},
			http.StatusBadRequest,
			"",
		},
		{
			"empty email", mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return nil, domain.ErrInvalidEmail
			}},
			createUserRequest{Username: "johndoe", Email: "", Password: "password123"},
			http.StatusBadRequest,
			"",
		},
		{
			"empty password", mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return nil, service.ErrInvalidPasswordLength
			}},
			createUserRequest{Username: "johndoe", Email: "user@example.com", Password: ""},
			http.StatusBadRequest,
			"",
		},
		{
			"invalid email", mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return nil, domain.ErrInvalidEmail
			}},
			createUserRequest{Username: "johndoe", Email: "invalidemail", Password: "password123"},
			http.StatusBadRequest,
			"",
		},
		{
			"too long username", mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return nil, domain.ErrInvalidUserName
			}},
			createUserRequest{Username: strings.Repeat("a", 31), Email: "user@example.com", Password: "password123"},
			http.StatusBadRequest,
			"",
		},
		{
			"too long password", mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return nil, service.ErrInvalidPasswordLength
			}},
			createUserRequest{Username: "johndoe", Email: "user@example.com", Password: strings.Repeat("a", 61)},
			http.StatusBadRequest,
			"",
		},
		{
			"too short password", mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return nil, service.ErrInvalidPasswordLength
			}},
			createUserRequest{Username: "johndoe", Email: "user@example.com", Password: "short"},
			http.StatusBadRequest,
			"",
		},
		{
			"user already exists", mockUserService{mockCreateFunc: func(ctx context.Context, username, email, password string) (*domain.User, error) {
				return nil, service.ErrUserAlreadyExists
			}},
			createUserRequest{Username: "johndoe", Email: "user@example.com", Password: "short"},
			http.StatusConflict,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewUserController(&tt.service)
			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctrl.Create(w, req)
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

func TestUserController_GetByID(t *testing.T) {
	foundUser := userResponse{ID: 67, Username: "johndoe", Email: "user@example.com"}
	user, _ := json.Marshal(foundUser)
	resp := string(user)

	tests := []struct {
		name           string
		service        mockUserService
		userID         string
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockUserService{mockGetByIDFunc: func(ctx context.Context, userID int64) (*domain.User, error) {
				return &domain.User{ID: 67, Username: "johndoe", Email: "user@example.com"}, nil
			}},
			"67",
			http.StatusOK,
			resp,
		},
		{
			"invalid userID",
			mockUserService{},
			"abc",
			http.StatusBadRequest,
			`{"error": "invalid user ID"}`,
		},
		{
			"user not found",
			mockUserService{mockGetByIDFunc: func(ctx context.Context, userID int64) (*domain.User, error) {
				return nil, service.ErrUserNotFound
			}},
			"999",
			http.StatusNotFound,
			`{"error": "user not found"}`,
		},
		{
			"unexpected error",
			mockUserService{mockGetByIDFunc: func(ctx context.Context, userID int64) (*domain.User, error) {
				return nil, service.ErrUnexpected
			}},
			"67",
			http.StatusBadRequest,
			`{"error": "failed to get user"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewUserController(&tt.service)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userID", tt.userID)
			ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
			req := httptest.NewRequest(http.MethodGet, "/users/id/"+tt.userID, nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			ctrl.GetByID(w, req)
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

func TestUserController_GetByEmail(t *testing.T) {
	foundUser := userResponse{ID: 67, Username: "johndoe", Email: "user@example.com"}
	user, _ := json.Marshal(foundUser)
	resp := string(user)

	tests := []struct {
		name           string
		service        mockUserService
		email          string
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockUserService{mockGetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: 67, Username: "johndoe", Email: email}, nil
			}},
			"user@example.com",
			http.StatusOK,
			resp,
		},
		{
			"email with URL encoding",
			mockUserService{mockGetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: 67, Username: "johndoe", Email: email}, nil
			}},
			"user%40example.com",
			http.StatusOK,
			resp,
		},
		{
			"user not found",
			mockUserService{mockGetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, service.ErrUserNotFound
			}},
			"unknown@example.com",
			http.StatusNotFound,
			`{"error": "user not found"}`,
		},
		{
			"unexpected error",
			mockUserService{mockGetByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, service.ErrUnexpected
			}},
			"user@example.com",
			http.StatusBadRequest,
			`{"error": "failed to get user"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewUserController(&tt.service)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("email", tt.email)
			ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
			req := httptest.NewRequest(http.MethodGet, "/users/email/"+tt.email, nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			ctrl.GetByEmail(w, req)
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

func TestUserController_Update(t *testing.T) {
	tests := []struct {
		name           string
		service        mockUserService
		userID         string
		input          updateUserRequest
		currUserID     string
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockUserService{mockUpdateFunc: func(ctx context.Context, currUserID, userID int64, username, email, password string) error {
				return nil
			}},
			"67",
			updateUserRequest{Username: "newname", Email: "new@example.com", Password: "newpassword123"},
			"67",
			http.StatusOK,
			"",
		},
		{
			"invalid userID in path",
			mockUserService{},
			"abc",
			updateUserRequest{Username: "newname", Email: "new@example.com", Password: "newpassword123"},
			"67",
			http.StatusBadRequest,
			`{"error": "invalid user ID"}`,
		},
		{
			"invalid currUserID in context",
			mockUserService{},
			"67",
			updateUserRequest{Username: "newname", Email: "new@example.com", Password: "newpassword123"},
			"abc",
			http.StatusBadRequest,
			`{"error": "invalid user ID"}`,
		},
		{
			"invalid JSON body",
			mockUserService{},
			"67",
			updateUserRequest{},
			"67",
			http.StatusBadRequest,
			`{"error": "invalid JSON"}`,
		},
		{
			"user not found",
			mockUserService{mockUpdateFunc: func(ctx context.Context, currUserID, userID int64, username, email, password string) error {
				return service.ErrUserNotFound
			}},
			"67",
			updateUserRequest{Username: "newname", Email: "new@example.com", Password: "newpassword123"},
			"67",
			http.StatusNotFound,
			`{"error": "user not found"}`,
		},
		{
			"unexpected error",
			mockUserService{mockUpdateFunc: func(ctx context.Context, currUserID, userID int64, username, email, password string) error {
				return service.ErrUnexpected
			}},
			"67",
			updateUserRequest{Username: "newname", Email: "new@example.com", Password: "newpassword123"},
			"67",
			http.StatusBadRequest,
			`{"error": "failed to get user"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewUserController(&tt.service)

			var bodyReader *strings.Reader
			if tt.expectedBody == `{"error": "invalid JSON"}` {
				bodyReader = strings.NewReader("not valid json{")
			} else {
				bodyBytes, _ := json.Marshal(tt.input)
				bodyReader = strings.NewReader(string(bodyBytes))
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userID", tt.userID)
			ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, "userID", tt.currUserID)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.userID, bodyReader)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctrl.Update(w, req)
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

func TestUserController_Delete(t *testing.T) {
	tests := []struct {
		name           string
		service        mockUserService
		userID         string
		currUserID     string
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockUserService{mockDeleteFunc: func(ctx context.Context, currUserID, userID int64) error {
				return nil
			}},
			"67",
			"67",
			http.StatusNoContent,
			"",
		},
		{
			"invalid userID in path",
			mockUserService{},
			"abc",
			"67",
			http.StatusBadRequest,
			`{"error": "invalid user ID"}`,
		},
		{
			"invalid currUserID in context",
			mockUserService{},
			"67",
			"abc",
			http.StatusBadRequest,
			`{"error": "invalid user ID"}`,
		},
		{
			"user not found",
			mockUserService{mockDeleteFunc: func(ctx context.Context, currUserID, userID int64) error {
				return service.ErrUserNotFound
			}},
			"999",
			"67",
			http.StatusNotFound,
			`{"error": "user not found"}`,
		},
		{
			"unexpected error",
			mockUserService{mockDeleteFunc: func(ctx context.Context, currUserID, userID int64) error {
				return service.ErrUnexpected
			}},
			"67",
			"67",
			http.StatusBadRequest,
			`{"error": "failed to get user"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewUserController(&tt.service)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userID", tt.userID)
			ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, "userID", tt.currUserID)
			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			ctrl.Delete(w, req)
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

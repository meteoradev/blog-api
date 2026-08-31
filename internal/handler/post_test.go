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

type mockPostService struct {
	mockCreateFunc         func(ctx context.Context, userID int64, title, content string) (*domain.Post, error)
	mockGetByIDFunc        func(ctx context.Context, postID int64) (*domain.Post, error)
	mockGetByTitleFunc     func(ctx context.Context, title string) (*domain.Post, error)
	mockUpdateFunc         func(ctx context.Context, postID int64, title, content string) error
	mockUpdateWithValidate func(ctx context.Context, currUserID, postID int64, title, content string) error
	mockDeleteFunc         func(ctx context.Context, postID int64) error
	mockDeleteWithValidate func(ctx context.Context, currUserID, postID int64) error
}

func (m *mockPostService) Create(ctx context.Context, userID int64, title, content string) (*domain.Post, error) {
	return m.mockCreateFunc(ctx, userID, title, content)
}
func (m *mockPostService) GetByID(ctx context.Context, postID int64) (*domain.Post, error) {
	return m.mockGetByIDFunc(ctx, postID)
}
func (m *mockPostService) GetByTitle(ctx context.Context, title string) (*domain.Post, error) {
	return m.mockGetByTitleFunc(ctx, title)
}
func (m *mockPostService) Update(ctx context.Context, postID int64, title, content string) error {
	return m.mockUpdateFunc(ctx, postID, title, content)
}
func (m *mockPostService) UpdateWithValidate(ctx context.Context, currUserID, postID int64, title, content string) error {
	return m.mockUpdateWithValidate(ctx, currUserID, postID, title, content)
}
func (m *mockPostService) Delete(ctx context.Context, postID int64) error {
	return m.mockDeleteFunc(ctx, postID)
}
func (m *mockPostService) DeleteWithValidate(ctx context.Context, currUserID, postID int64) error {
	return m.mockDeleteWithValidate(ctx, currUserID, postID)
}

func TestPostController_Create(t *testing.T) {
	tests := []struct {
		name           string
		service        mockPostService
		input          createPostRequest
		userID         int64
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockPostService{mockCreateFunc: func(ctx context.Context, userID int64, title, content string) (*domain.Post, error) {
				return &domain.Post{ID: 1, UserID: userID, Title: title, Content: content}, nil
			}},
			createPostRequest{Title: "Hello World", Content: "Some content"},
			10,
			http.StatusCreated,
			"",
		},
		{
			"empty title",
			mockPostService{mockCreateFunc: func(ctx context.Context, userID int64, title, content string) (*domain.Post, error) {
				return nil, domain.ErrInvalidTitle
			}},
			createPostRequest{Title: "", Content: "Some content"},
			10,
			http.StatusBadRequest,
			`{"error": "title must contain at least 1 character"`,
		},
		{
			"empty content",
			mockPostService{mockCreateFunc: func(ctx context.Context, userID int64, title, content string) (*domain.Post, error) {
				return nil, domain.ErrInvalidContent
			}},
			createPostRequest{Title: "Hello", Content: ""},
			10,
			http.StatusBadRequest,
			`{"error": "content must contain at least 1 character"`,
		},
		{
			"linked user not found",
			mockPostService{mockCreateFunc: func(ctx context.Context, userID int64, title, content string) (*domain.Post, error) {
				return nil, service.ErrLinkedUserNotFound
			}},
			createPostRequest{Title: "Hello", Content: "Content"},
			999,
			http.StatusConflict,
			`{"error": "linked user with this id doesnt exists."}`,
		},
		{
			"unexpected error",
			mockPostService{mockCreateFunc: func(ctx context.Context, userID int64, title, content string) (*domain.Post, error) {
				return nil, service.ErrUnexpected
			}},
			createPostRequest{Title: "Hello", Content: "Content"},
			10,
			http.StatusInternalServerError,
			`{"error": "failed to create post"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewPostController(&tt.service)
			body, _ := json.Marshal(tt.input)
			ctx := context.WithValue(context.Background(), "userID", tt.userID)
			req := httptest.NewRequest(http.MethodPost, "/posts/", strings.NewReader(string(body)))
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctrl.Create(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedBody != "" {
				bodyStr := strings.TrimSpace(w.Body.String())
				if bodyStr != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, bodyStr)
				}
			} else {
				var got postResponse
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if got.ID != 1 || got.Title != "Hello World" || got.UserID != 10 {
					t.Errorf("unexpected response body: %v", got)
				}
			}
		})
	}
}

func TestPostController_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		service        mockPostService
		postID         string
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockPostService{mockGetByIDFunc: func(ctx context.Context, postID int64) (*domain.Post, error) {
				return &domain.Post{ID: 42, UserID: 10, Title: "Hello World", Content: "Some content"}, nil
			}},
			"42",
			http.StatusOK,
			"",
		},
		{
			"invalid postID",
			mockPostService{},
			"abc",
			http.StatusBadRequest,
			`{"error": "invalid post ID"}`,
		},
		{
			"post not found",
			mockPostService{mockGetByIDFunc: func(ctx context.Context, postID int64) (*domain.Post, error) {
				return nil, service.ErrPostNotFound
			}},
			"999",
			http.StatusNotFound,
			`{"error": "post not found"}`,
		},
		{
			"unexpected error",
			mockPostService{mockGetByIDFunc: func(ctx context.Context, postID int64) (*domain.Post, error) {
				return nil, service.ErrUnexpected
			}},
			"42",
			http.StatusBadRequest,
			`{"error": "failed to get post"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewPostController(&tt.service)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("postID", tt.postID)
			ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
			req := httptest.NewRequest(http.MethodGet, "/posts/id/"+tt.postID, nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			ctrl.GetByID(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedBody != "" {
				bodyStr := strings.TrimSpace(w.Body.String())
				if bodyStr != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, bodyStr)
				}
			} else {
				var got postResponse
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if got.ID != 42 || got.Title != "Hello World" || got.UserID != 10 {
					t.Errorf("unexpected response body: %v", got)
				}
			}
		})
	}
}

func TestPostController_GetByTitle(t *testing.T) {
	tests := []struct {
		name           string
		service        mockPostService
		title          string
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockPostService{mockGetByTitleFunc: func(ctx context.Context, title string) (*domain.Post, error) {
				return &domain.Post{ID: 42, UserID: 10, Title: title, Content: "Some content"}, nil
			}},
			"Hello%20World",
			http.StatusOK,
			"",
		},
		{
			"empty title",
			mockPostService{},
			"",
			http.StatusBadRequest,
			`{"error": "invalid post title"}`,
		},
		{
			"post not found",
			mockPostService{mockGetByTitleFunc: func(ctx context.Context, title string) (*domain.Post, error) {
				return nil, service.ErrPostNotFound
			}},
			"unknown-title",
			http.StatusNotFound,
			`{"error": "post not found"}`,
		},
		{
			"unexpected error",
			mockPostService{mockGetByTitleFunc: func(ctx context.Context, title string) (*domain.Post, error) {
				return nil, service.ErrUnexpected
			}},
			"Hello%20World",
			http.StatusBadRequest,
			`{"error": "failed to get post"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewPostController(&tt.service)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("title", tt.title)
			ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
			req := httptest.NewRequest(http.MethodGet, "/posts/title/"+tt.title, nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			ctrl.GetByTitle(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedBody != "" {
				bodyStr := strings.TrimSpace(w.Body.String())
				if bodyStr != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, bodyStr)
				}
			} else {
				var got postResponse
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if got.ID != 42 || got.UserID != 10 {
					t.Errorf("unexpected response body: %v", got)
				}
			}
		})
	}
}

func TestPostController_Update(t *testing.T) {
	tests := []struct {
		name           string
		service        mockPostService
		postID         string
		input          updatePostRequest
		currUserID     any
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockPostService{mockUpdateWithValidate: func(ctx context.Context, currUserID, postID int64, title, content string) error {
				return nil
			}},
			"42",
			updatePostRequest{ID: 42, Title: "Updated Title", Content: "Updated content"},
			"10",
			http.StatusOK,
			"",
		},
		{
			"invalid JSON body",
			mockPostService{},
			"42",
			updatePostRequest{},
			"10",
			http.StatusBadRequest,
			`{"error": "invalid JSON"}`,
		},
		{
			"invalid currUserID in context",
			mockPostService{},
			"42",
			updatePostRequest{ID: 42, Title: "Updated", Content: "Content"},
			"not_a_number",
			http.StatusBadRequest,
			`{"error": "invalid user ID"}`,
		},
		{
			"update failed",
			mockPostService{mockUpdateWithValidate: func(ctx context.Context, currUserID, postID int64, title, content string) error {
				return service.ErrUpdatePostFailed
			}},
			"42",
			updatePostRequest{ID: 42, Title: "Updated", Content: "Content"},
			"10",
			http.StatusBadRequest,
			`{"error": "failed to get post"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewPostController(&tt.service)

			var bodyReader *strings.Reader
			if tt.expectedBody == `{"error": "invalid JSON"}` {
				bodyReader = strings.NewReader("not valid json{")
			} else {
				bodyBytes, _ := json.Marshal(tt.input)
				bodyReader = strings.NewReader(string(bodyBytes))
			}

			ctx := context.WithValue(context.Background(), "userID", tt.currUserID)
			req := httptest.NewRequest(http.MethodPut, "/posts/42", bodyReader)
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			ctrl.Update(w, req)
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if tt.expectedBody != "" {
				bodyStr := strings.TrimSpace(w.Body.String())
				if bodyStr != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, bodyStr)
				}
			}
		})
	}
}

func TestPostController_Delete(t *testing.T) {
	tests := []struct {
		name           string
		service        mockPostService
		postID         string
		currUserID     any
		expectedStatus int
		expectedBody   string
	}{
		{
			"happy path",
			mockPostService{mockDeleteWithValidate: func(ctx context.Context, currUserID, postID int64) error {
				return nil
			}},
			"42",
			int64(10),
			http.StatusNoContent,
			"",
		},
		{
			"invalid postID",
			mockPostService{},
			"abc",
			int64(10),
			http.StatusBadRequest,
			`{"error": "invalid post ID"}`,
		},
		{
			"invalid currUserID in context",
			mockPostService{},
			"42",
			"not_a_number",
			http.StatusBadRequest,
			`{"error": "invalid user ID"}`,
		},
		{
			"post not found",
			mockPostService{mockDeleteWithValidate: func(ctx context.Context, currUserID, postID int64) error {
				return service.ErrPostNotFound
			}},
			"999",
			int64(10),
			http.StatusNotFound,
			`{"error": "post not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewPostController(&tt.service)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("postID", tt.postID)
			ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
			ctx = context.WithValue(ctx, "userID", tt.currUserID)
			req := httptest.NewRequest(http.MethodDelete, "/posts/"+tt.postID, nil)
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

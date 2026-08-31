package service_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/service"
)

type mockPostRepo struct {
	createPostFunc             func(ctx context.Context, post *domain.Post) (*domain.Post, error)
	getPostByIDFunc            func(ctx context.Context, ID int64) (*domain.Post, error)
	getPostByTitleFunc         func(ctx context.Context, title string) (*domain.Post, error)
	updatePostFunc             func(ctx context.Context, post *domain.Post) error
	updateWithValidatePostFunc func(ctx context.Context, currUserID int64, post *domain.Post) error
	deletePostFunc             func(ctx context.Context, id int64) (string, error)
	deletePostWithValidateFunc func(ctx context.Context, currUserId, id int64) (string, error)
}

func (m *mockPostRepo) Create(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	return m.createPostFunc(ctx, post)
}
func (m *mockPostRepo) GetByID(ctx context.Context, ID int64) (*domain.Post, error) {
	return m.getPostByIDFunc(ctx, ID)
}
func (m *mockPostRepo) GetByTitle(ctx context.Context, title string) (*domain.Post, error) {
	return m.getPostByTitleFunc(ctx, title)
}
func (m *mockPostRepo) Update(ctx context.Context, post *domain.Post) error {
	return m.updatePostFunc(ctx, post)
}
func (m *mockPostRepo) Delete(ctx context.Context, ID int64) (string, error) {
	return m.deletePostFunc(ctx, ID)
}
func (m *mockPostRepo) UpdateWithValidate(ctx context.Context, currUserID int64, post *domain.Post) error {
	return m.updateWithValidatePostFunc(ctx, currUserID, post)
}

func (m *mockPostRepo) DeleteWithValidate(ctx context.Context, currUserID, ID int64) (string, error) {
	return m.deletePostWithValidateFunc(ctx, currUserID, ID)
}

type mockCache struct {
	getFunc  func(ctx context.Context, key string) (string, bool)
	setFunc  func(ctx context.Context, key string, value any, ttl time.Duration) bool
	delFunc  func(ctx context.Context, key string) bool
	GetCalls int
	SetCalls int
	DelCalls int
}

func (m *mockCache) Get(ctx context.Context, key string) (string, bool) {
	m.GetCalls++
	return m.getFunc(ctx, key)
}

func (m *mockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) bool {
	m.SetCalls++
	return m.setFunc(ctx, key, value, ttl)
}
func (m *mockCache) Del(ctx context.Context, key string) bool {
	m.DelCalls++
	return m.delFunc(ctx, key)
}

func TestPostService_Create(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cache := &mockCache{SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
			return true
		},
		}
		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{createPostFunc: func(ctx context.Context, post *domain.Post) (*domain.Post, error) {
			return prepPost, nil
		},
		}
		svc := service.NewPostService(repo, cache)
		postFromSVC, err := svc.Create(context.Background(), prepPost.UserID, prepPost.Title, prepPost.Content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prepPost.ID != postFromSVC.ID || prepPost.UserID != postFromSVC.UserID || prepPost.Title != postFromSVC.Title || prepPost.Content != postFromSVC.Content {
			t.Errorf("got post %v, want %v", postFromSVC, prepPost)
		}
		if cache.SetCalls != 1 {
			t.Errorf("expected 1 set call, got %d", cache.SetCalls)
		}
	})

	t.Run("cache error", func(t *testing.T) {
		cache := &mockCache{SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
			return false
		},
		}
		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{createPostFunc: func(ctx context.Context, post *domain.Post) (*domain.Post, error) {
			return prepPost, nil
		},
		}
		svc := service.NewPostService(repo, cache)
		postFromSVC, err := svc.Create(context.Background(), prepPost.UserID, prepPost.Title, prepPost.Content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prepPost.ID != postFromSVC.ID || prepPost.UserID != postFromSVC.UserID || prepPost.Title != postFromSVC.Title || prepPost.Content != postFromSVC.Content {
			t.Errorf("got post %v, want %v", postFromSVC, prepPost)
		}
		if cache.SetCalls != 1 {
			t.Errorf("expected 1 set call, got %d", cache.SetCalls)
		}
	})
	t.Run("no linked user", func(t *testing.T) {
		cache := &mockCache{SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
			return true
		},
		}
		prepPost := &domain.Post{UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{createPostFunc: func(ctx context.Context, post *domain.Post) (*domain.Post, error) {
			return nil, sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		_, err := svc.Create(context.Background(), prepPost.UserID, prepPost.Title, prepPost.Content)
		if err == nil {
			t.Errorf("expected %v, got nil", service.ErrLinkedUserNotFound)
		}
		if cache.SetCalls != 0 {
			t.Errorf("expected 0 set call, got %d", cache.SetCalls)
		}
	})
}

func TestPostService_GetByID(t *testing.T) {
	t.Run("happy path from db", func(t *testing.T) {
		cache := &mockCache{GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		},
			setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
				return true
			},
		}
		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{getPostByIDFunc: func(ctx context.Context, postID int64) (*domain.Post, error) {
			return prepPost, nil
		},
		}
		svc := service.NewPostService(repo, cache)
		postFromSVC, err := svc.GetByID(context.Background(), prepPost.UserID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prepPost.ID != postFromSVC.ID || prepPost.UserID != postFromSVC.UserID || prepPost.Title != postFromSVC.Title || prepPost.Content != postFromSVC.Content {
			t.Errorf("got post %v, want %v", postFromSVC, prepPost)
		}
		if cache.GetCalls != 1 {
			t.Errorf("expected 1 get call, got %d", cache.GetCalls)
		}
	})

	t.Run("happy path from cache", func(t *testing.T) {
		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		cache := &mockCache{GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			data, _ := json.Marshal(prepPost)
			return string(data), true
		},
			setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
				return true
			},
		}
		repo := &mockPostRepo{getPostByIDFunc: func(ctx context.Context, postID int64) (*domain.Post, error) {
			t.Fatalf("Post from db, wanted from cache.")
			return nil, nil
		},
		}
		svc := service.NewPostService(repo, cache)
		postFromSVC, err := svc.GetByID(context.Background(), prepPost.UserID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prepPost.ID != postFromSVC.ID || prepPost.UserID != postFromSVC.UserID || prepPost.Title != postFromSVC.Title || prepPost.Content != postFromSVC.Content {
			t.Errorf("got post %v, want %v", postFromSVC, prepPost)
		}
		if cache.GetCalls != 1 {
			t.Errorf("expected 1 get call, got %d", cache.GetCalls)
		}
	})
	t.Run("from db, cache unavailable", func(t *testing.T) {
		cache := &mockCache{GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		},
			SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
				return false
			},
		}
		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{getPostByIDFunc: func(ctx context.Context, postID int64) (*domain.Post, error) {
			return prepPost, nil
		},
		}
		svc := service.NewPostService(repo, cache)
		postFromSVC, err := svc.GetByID(context.Background(), prepPost.UserID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prepPost.ID != postFromSVC.ID || prepPost.UserID != postFromSVC.UserID || prepPost.Title != postFromSVC.Title || prepPost.Content != postFromSVC.Content {
			t.Errorf("got post %v, want %v", postFromSVC, prepPost)
		}
		if cache.GetCalls != 1 {
			t.Errorf("expected 1 get call, got %d", cache.GetCalls)
		}
		if cache.SetCalls != 1 {
			t.Errorf("expected 1 set call, got %d", cache.SetCalls)
		}

	})
	t.Run("not found", func(t *testing.T) {
		cache := &mockCache{GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		},
		}
		repo := &mockPostRepo{getPostByIDFunc: func(ctx context.Context, postID int64) (*domain.Post, error) {
			return nil, sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		_, err := svc.GetByID(context.Background(), 67)
		if !errors.Is(err, service.ErrPostNotFound) {
			t.Fatalf("got wrong err: %v, wanted %v", err, service.ErrPostNotFound)
		}
		if cache.GetCalls != 1 {
			t.Errorf("expected 1 get call, got %d", cache.GetCalls)
		}

	})
}

func TestPostService_GetByTitle(t *testing.T) {
	t.Run("happy path from db", func(t *testing.T) {
		cache := &mockCache{GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		},
			SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
				return true
			},
		}
		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{getPostByTitleFunc: func(ctx context.Context, title string) (*domain.Post, error) {
			return prepPost, nil
		},
		}
		svc := service.NewPostService(repo, cache)
		postFromSVC, err := svc.GetByTitle(context.Background(), prepPost.Title)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prepPost.ID != postFromSVC.ID || prepPost.UserID != postFromSVC.UserID || prepPost.Title != postFromSVC.Title || prepPost.Content != postFromSVC.Content {
			t.Errorf("got post %v, want %v", postFromSVC, prepPost)
		}
		if cache.GetCalls != 1 {
			t.Errorf("expected 1 get call, got %d", cache.GetCalls)
		}
		if cache.SetCalls != 1 {
			t.Errorf("expected 1 set call, got %d", cache.SetCalls)
		}
	})

	t.Run("happy path from cache", func(t *testing.T) {
		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		cache := &mockCache{GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			data, _ := json.Marshal(prepPost)
			return string(data), true
		},
			SetCalls: 1, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
				return true
			},
		}
		repo := &mockPostRepo{getPostByTitleFunc: func(ctx context.Context, title string) (*domain.Post, error) {
			t.Fatalf("Post from db, wanted from cache.")
			return nil, nil
		},
		}
		svc := service.NewPostService(repo, cache)
		postFromSVC, err := svc.GetByTitle(context.Background(), prepPost.Title)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prepPost.ID != postFromSVC.ID || prepPost.UserID != postFromSVC.UserID || prepPost.Title != postFromSVC.Title || prepPost.Content != postFromSVC.Content {
			t.Errorf("got post %v, want %v", postFromSVC, prepPost)
		}
		if cache.GetCalls != 1 {
			t.Errorf("expected 1 get call, got %d", cache.GetCalls)
		}
		if cache.SetCalls != 1 {
			t.Errorf("expected 1 set call, got %d", cache.SetCalls)
		}
	})
	t.Run("from db, cache unavailable", func(t *testing.T) {
		cache := &mockCache{GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		},
			SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
				return false
			},
		}
		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{getPostByTitleFunc: func(ctx context.Context, title string) (*domain.Post, error) {
			return prepPost, nil
		},
		}
		svc := service.NewPostService(repo, cache)
		postFromSVC, err := svc.GetByTitle(context.Background(), prepPost.Title)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prepPost.ID != postFromSVC.ID || prepPost.UserID != postFromSVC.UserID || prepPost.Title != postFromSVC.Title || prepPost.Content != postFromSVC.Content {
			t.Errorf("got post %v, want %v", postFromSVC, prepPost)
		}
		if cache.GetCalls != 1 {
			t.Errorf("expected 1 get call, got %d", cache.GetCalls)
		}
		if cache.SetCalls != 1 {
			t.Errorf("expected 1 set call, got %d", cache.SetCalls)
		}

	})
	t.Run("not found", func(t *testing.T) {
		cache := &mockCache{GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		},
		}
		repo := &mockPostRepo{getPostByTitleFunc: func(ctx context.Context, title string) (*domain.Post, error) {
			return nil, sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		_, err := svc.GetByTitle(context.Background(), "aaaa")
		if !errors.Is(err, service.ErrPostNotFound) {
			t.Fatalf("got wrong err: %v, wanted %v", err, service.ErrPostNotFound)
		}
		if cache.GetCalls != 1 {
			t.Errorf("expected 1 get call, got %d", cache.GetCalls)
		}

	})
}

func TestPostService_UpdateWithValidate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		}, GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		}, SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
			return true
		}}

		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{updateWithValidatePostFunc: func(ctx context.Context, currUserID int64, post *domain.Post) error {
			return nil
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.UpdateWithValidate(context.Background(), prepPost.UserID, prepPost.ID, prepPost.Title, prepPost.Content); err != nil {
			t.Fatalf("got err %v, wanted %v", err, nil)
			if cache.DelCalls != 2 {
				t.Errorf("expected 2 del call, got %d", cache.DelCalls)
			}
		}
		if cache.DelCalls != 2 {
			t.Errorf("expected 2 get call, got %d", cache.GetCalls)
		}
	})
	t.Run("not found", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		}, GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		}, SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
			return true
		}}

		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{updateWithValidatePostFunc: func(ctx context.Context, currUserID int64, post *domain.Post) error {
			return sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.UpdateWithValidate(context.Background(), prepPost.UserID, prepPost.ID, prepPost.Title, prepPost.Content); !errors.Is(err, service.ErrUpdatePostFailed) {
			t.Fatalf("got err %v, wanted %v", err, nil)
			if cache.DelCalls != 2 {
				t.Errorf("expected 2 del call, got %d", cache.DelCalls)
			}
		}
		if cache.DelCalls != 0 {
			t.Errorf("expected 0 get call, got %d", cache.GetCalls)
		}
	})
	t.Run("user validation failed", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		}, GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		}, SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
			return true
		}}

		prepPost := &domain.Post{ID: 6767, UserID: 5252, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{updateWithValidatePostFunc: func(ctx context.Context, currUserID int64, post *domain.Post) error {
			return sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.UpdateWithValidate(context.Background(), prepPost.UserID, prepPost.ID, prepPost.Title, prepPost.Content); !errors.Is(err, service.ErrUpdatePostFailed) {
			t.Fatalf("got err %v, wanted %v", err, nil)
			if cache.DelCalls != 2 {
				t.Errorf("expected 2 del call, got %d", cache.DelCalls)
			}
		}
		if cache.DelCalls != 0 {
			t.Errorf("expected 2 get call, got %d", cache.GetCalls)
		}
	})
}

func TestPostService_Update(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		}, GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		}, SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
			return true
		}}

		prepPost := &domain.Post{ID: 6767, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{updatePostFunc: func(ctx context.Context, post *domain.Post) error {
			return nil
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.Update(context.Background(), prepPost.ID, prepPost.Title, prepPost.Content); err != nil {
			t.Fatalf("got err %v, wanted %v", err, nil)
			if cache.DelCalls != 2 {
				t.Errorf("expected 2 del call, got %d", cache.DelCalls)
			}
		}
		if cache.DelCalls != 2 {
			t.Errorf("expected 0 get call, got %d", cache.GetCalls)
		}
	})
	t.Run("not found", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		}, GetCalls: 0, getFunc: func(ctx context.Context, key string) (string, bool) {
			return "", false
		}, SetCalls: 0, setFunc: func(ctx context.Context, key string, value any, ttl time.Duration) bool {
			return true
		}}

		prepPost := &domain.Post{ID: 6767, Title: "Test Post", Content: "Test Content"}
		repo := &mockPostRepo{updatePostFunc: func(ctx context.Context, post *domain.Post) error {
			return sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.Update(context.Background(), prepPost.ID, prepPost.Title, prepPost.Content); !errors.Is(err, service.ErrPostNotFound) {
			t.Fatalf("got err %v, wanted %v", err, nil)
			if cache.DelCalls != 2 {
				t.Errorf("expected 2 del call, got %d", cache.DelCalls)
			}
		}
		if cache.DelCalls != 0 {
			t.Errorf("expected 0 get call, got %d", cache.GetCalls)
		}
	})

}

func TestPostService_DeleteWithValidate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		},
		}
		repo := &mockPostRepo{deletePostWithValidateFunc: func(ctx context.Context, currUserID, postID int64) (string, error) {
			return "Test Post", nil
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.DeleteWithValidate(context.Background(), 6767, 5252); err != nil {
			t.Fatalf("got err %v, wanted %v", err, nil)
		}
		if cache.DelCalls != 2 {
			t.Errorf("expected 2 del call, got %d", cache.DelCalls)
		}
	})

	t.Run("user validation failed", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		},
		}
		repo := &mockPostRepo{deletePostWithValidateFunc: func(ctx context.Context, currUserID, postID int64) (string, error) {
			return "", sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.DeleteWithValidate(context.Background(), 6767, 5252); !errors.Is(err, service.ErrDeletePostFailed) {
			t.Fatalf("got err %v, wanted %v", err, nil)
		}
		if cache.DelCalls != 0 {
			t.Errorf("expected 0 del call, got %d", cache.DelCalls)
		}
	})
	t.Run("post not found", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		},
		}
		repo := &mockPostRepo{deletePostWithValidateFunc: func(ctx context.Context, currUserID, postID int64) (string, error) {
			return "", sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.DeleteWithValidate(context.Background(), 6767, 5252); !errors.Is(err, service.ErrDeletePostFailed) {
			t.Fatalf("got err %v, wanted %v", err, nil)
		}
		if cache.DelCalls != 0 {
			t.Errorf("expected 0 del call, got %d", cache.DelCalls)
		}
	})
}

func TestPostService_Delete(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		},
		}
		repo := &mockPostRepo{deletePostFunc: func(ctx context.Context, postID int64) (string, error) {
			return "Test Post", nil
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.Delete(context.Background(), 6767); err != nil {
			t.Fatalf("got err %v, wanted %v", err, nil)
		}
		if cache.DelCalls != 2 {
			t.Errorf("expected 2 del call, got %d", cache.DelCalls)
		}
	})

	t.Run("post not found", func(t *testing.T) {
		cache := &mockCache{DelCalls: 0, delFunc: func(ctx context.Context, key string) bool {
			return true
		},
		}
		repo := &mockPostRepo{deletePostFunc: func(ctx context.Context, postID int64) (string, error) {
			return "", sql.ErrNoRows
		},
		}
		svc := service.NewPostService(repo, cache)
		if err := svc.Delete(context.Background(), 6767); !errors.Is(err, service.ErrPostNotFound) {
			t.Fatalf("got err %v, wanted %v", err, nil)
		}
		if cache.DelCalls != 0 {
			t.Errorf("expected 0 del call, got %d", cache.DelCalls)
		}
	})
}

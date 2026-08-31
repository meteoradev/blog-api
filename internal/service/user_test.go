package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/service"
)

type mockUserRepo struct {
	getUserByIDFunc    func(ctx context.Context, id int64) (*domain.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	createUserFunc     func(ctx context.Context, user *domain.User) (*domain.User, error)
	updateUserFunc     func(ctx context.Context, user *domain.User) error
	deleteUserFunc     func(ctx context.Context, id int64) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	return m.createUserFunc(ctx, user)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return m.getUserByIDFunc(ctx, id)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.getUserByEmailFunc(ctx, email)
}
func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return m.updateUserFunc(ctx, user)
}
func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	return m.deleteUserFunc(ctx, id)
}

type mockHasher struct {
	hashFunc  func(s string) (string, error)
	checkFunc func(hashed, plain string) error
}

func (m *mockHasher) Hash(s string) (string, error) {
	return m.hashFunc(s)
}
func (m *mockHasher) Check(hashed, plain string) error {
	return m.checkFunc(hashed, plain)
}

func TestUserService_Create(t *testing.T) {
	hasher := &mockHasher{
		hashFunc: func(s string) (string, error) { return "hashed_" + s, nil },
	}

	t.Run("happy path", func(t *testing.T) {
		repo := &mockUserRepo{
			createUserFunc: func(ctx context.Context, user *domain.User) (*domain.User, error) {
				return &domain.User{ID: 67, Username: user.Username, Email: user.Email, PasswordHash: user.PasswordHash}, nil
			},
		}
		svc := service.NewUserService(repo, hasher)

		user, err := svc.Create(context.Background(), "testuser", "test@mail.com", "password123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != 67 {
			t.Errorf("got ID %d, want 67", user.ID)
		}
		if user.Username != "testuser" {
			t.Errorf("got username %s, want 'testuser'", user.Username)
		}
		if user.Email != "test@mail.com" {
			t.Errorf("got email %s, want 'test@mail.com'", user.Email)
		}
		if user.PasswordHash != "hashed_password123" {
			t.Errorf("got password %s, want 'hashed_password123'", user.PasswordHash)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		repo := &mockUserRepo{
			createUserFunc: func(ctx context.Context, user *domain.User) (*domain.User, error) {
				return nil, sql.ErrNoRows
			},
		}
		svc := service.NewUserService(repo, hasher)
		_, err := svc.Create(context.Background(), "testuser", "test@mail.com", "password123")
		if !errors.Is(err, service.ErrUserAlreadyExists) {
			t.Errorf("got error %v, want %v", err, service.ErrUserAlreadyExists)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		repo := &mockUserRepo{
			createUserFunc: func(ctx context.Context, user *domain.User) (*domain.User, error) {
				return nil, sql.ErrNoRows
			},
		}
		svc := service.NewUserService(repo, hasher)
		_, err := svc.Create(context.Background(), "testuser", "test@mail.com", "password123")
		if !errors.Is(err, service.ErrUserAlreadyExists) {
			t.Errorf("got error %v, want %v", err, service.ErrUserAlreadyExists)
		}
	})
}

func TestUserService_GetByID(t *testing.T) {
	hasher := &mockHasher{
		hashFunc: func(s string) (string, error) { return "hashed_" + s, nil },
	}
	user := &domain.User{ID: 67, Username: "testuser", Email: "test@mail.com", PasswordHash: "hashed_password123"}
	t.Run("happy path", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
				return &domain.User{ID: user.ID, Username: user.Username, Email: user.Email, PasswordHash: user.PasswordHash}, nil
			},
		}

		svc := service.NewUserService(repo, hasher)
		userFromSVC, err := svc.GetByID(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if userFromSVC.ID != user.ID || userFromSVC.Username != user.Username || userFromSVC.Email != user.Email || userFromSVC.PasswordHash != user.PasswordHash {
			t.Errorf("got user %v, want %v", userFromSVC, user)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
				return nil, sql.ErrNoRows
			},
		}
		svc := service.NewUserService(repo, hasher)
		_, err := svc.GetByID(context.Background(), user.ID)
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Errorf("got error %v, want %v", err, service.ErrUserNotFound)
		}
	})
}

func TestUserService_GetByEmail(t *testing.T) {
	hasher := &mockHasher{
		hashFunc: func(s string) (string, error) { return "hashed_" + s, nil },
	}
	user := &domain.User{ID: 67, Username: "testuser", Email: "test@mail.com", PasswordHash: "hashed_password123"}
	t.Run("happy path", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: user.ID, Username: user.Username, Email: user.Email, PasswordHash: user.PasswordHash}, nil
			},
		}
		svc := service.NewUserService(repo, hasher)
		userFromSVC, err := svc.GetByEmail(context.Background(), user.Email)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if userFromSVC.ID != user.ID || userFromSVC.Username != user.Username || userFromSVC.Email != user.Email || userFromSVC.PasswordHash != user.PasswordHash {
			t.Errorf("got user %v, want %v", userFromSVC, user)
		}
	})
	t.Run("user not found", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, sql.ErrNoRows
			},
		}
		svc := service.NewUserService(repo, hasher)
		_, err := svc.GetByEmail(context.Background(), user.Email)
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Errorf("got error %v, want %v", err, service.ErrUserNotFound)
		}
	})
}

func TestUserService_Update(t *testing.T) {
	hasher := &mockHasher{
		hashFunc: func(s string) (string, error) { return "hashed_" + s, nil },
	}
	t.Run("happy path", func(t *testing.T) {
		repo := &mockUserRepo{
			updateUserFunc: func(ctx context.Context, user *domain.User) error {
				return nil
			},
		}
		svc := service.NewUserService(repo, hasher)
		err := svc.Update(context.Background(), 67, 67, "testuser", "test@mail.com", "hashed_password123")
		if err != nil {
			t.Errorf("got error %v, want %v", err, nil)
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("user not found", func(t *testing.T) {
		repo := &mockUserRepo{
			updateUserFunc: func(ctx context.Context, user *domain.User) error {
				return sql.ErrNoRows
			},
		}
		svc := service.NewUserService(repo, hasher)
		err := svc.Update(context.Background(), 67, 67, "testuser", "test@mail.com", "hashed_password123")
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Errorf("got error %v, want %v", err, service.ErrUserNotFound)
		}
	})
	t.Run("validation failed", func(t *testing.T) {
		repo := &mockUserRepo{
			updateUserFunc: func(ctx context.Context, user *domain.User) error {
				return nil
			},
		}
		svc := service.NewUserService(repo, hasher)
		err := svc.Update(context.Background(), 52, 67, "testuser", "test@mail.com", "")
		if !errors.Is(err, service.ErrMethodNotAllowed) {
			t.Errorf("got error %v, want %v", err, service.ErrMethodNotAllowed)
		}
	})
}

func TestUserService_Delete(t *testing.T) {
	hasher := &mockHasher{
		hashFunc: func(s string) (string, error) { return "hashed_" + s, nil },
	}
	t.Run("happy path", func(t *testing.T) {
		repo := &mockUserRepo{
			deleteUserFunc: func(ctx context.Context, id int64) error {
				return nil
			},
		}
		svc := service.NewUserService(repo, hasher)
		err := svc.Delete(context.Background(), 67, 67)
		if err != nil {
			t.Errorf("got error %v, want %v", err, nil)
		}
	})
	t.Run("not found", func(t *testing.T) {
		repo := &mockUserRepo{
			deleteUserFunc: func(ctx context.Context, id int64) error {
				return sql.ErrNoRows
			},
		}
		svc := service.NewUserService(repo, hasher)
		err := svc.Delete(context.Background(), 67, 67)
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Errorf("got error %v, want %v", err, service.ErrUserNotFound)
		}
	})
	t.Run("validation failed", func(t *testing.T) {
		repo := &mockUserRepo{
			deleteUserFunc: func(ctx context.Context, id int64) error {
				return sql.ErrNoRows
			},
		}
		svc := service.NewUserService(repo, hasher)
		err := svc.Delete(context.Background(), 52, 67)
		if !errors.Is(err, service.ErrMethodNotAllowed) {
			t.Errorf("got error %v, want %v", err, service.ErrMethodNotAllowed)
		}
	})
}

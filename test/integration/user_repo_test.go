package integration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/repository/postgres"
)

func TestUserRepository_Create(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		user, err := repo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		if user.ID == 0 {
			t.Fatal("user ID is 0")
		}
		if user.CreatedAt.IsZero() {
			t.Fatal("created at is zero")
		}
		if user.Username != prepUser.Username || user.Email != prepUser.Email || user.PasswordHash != prepUser.PasswordHash {
			t.Fatalf("user data is not as expected, got %v, want %v", *user, *prepUser)
		}
		t.Cleanup(func() { _ = repo.Delete(ctx, user.ID) })
	})
	t.Run("duplicate username", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser1, _ := domain.NewUser(suffix+"johndoe", suffix+"user1@example.com", "passwordHash123")
		prepUser2, _ := domain.NewUser(suffix+"johndoe", suffix+"user2@example.com", "passwordHash123")
		u1, err := repo.Create(ctx, prepUser1)
		_, err = repo.Create(ctx, prepUser2)
		if err == nil {
			t.Fatalf("expected duplicate username error, got: %v", nil)
		}
		if !strings.Contains(err.Error(), "duplicate key value violates unique constraint \"users_username_key\"") {
			t.Fatalf("expected duplicate username error, got: %v", err)
		}

		t.Cleanup(func() { _ = repo.Delete(ctx, u1.ID) })
	})
	t.Run("duplicate email", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser1, _ := domain.NewUser(suffix+"johndoe1", suffix+"user@example.com", "passwordHash123")
		prepUser2, _ := domain.NewUser(suffix+"johndoe2", suffix+"user@example.com", "passwordHash123")
		u1, err := repo.Create(ctx, prepUser1)
		_, err = repo.Create(ctx, prepUser2)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatal("wanted no rows error, got:", err)
		}
		t.Cleanup(func() { _ = repo.Delete(ctx, u1.ID) })
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := repo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		u, err := repo.GetByID(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if u.ID != createdUser.ID || u.Username != createdUser.Username || u.Email != createdUser.Email || u.PasswordHash != createdUser.PasswordHash {
			t.Fatalf("user data is not as expected, got %v, want %v", *u, *createdUser)
		}
		t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })
	})
	t.Run("user not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		_, err := repo.GetByID(ctx, 67)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := repo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		u, err := repo.GetByEmail(ctx, createdUser.Email)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if u.ID != createdUser.ID || u.Username != createdUser.Username || u.Email != createdUser.Email || u.PasswordHash != createdUser.PasswordHash {
			t.Fatalf("user data is not as expected, got %v, want %v", *u, *createdUser)
		}
		t.Cleanup(func() { _ = repo.Delete(ctx, u.ID) })
	})
	t.Run("user not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		_, err := repo.GetByEmail(ctx, "67")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := repo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		userToUpdate, _ := domain.NewUser(suffix+"johndoe67", suffix+"user67@example.com", "passwordHash12367")
		userToUpdate.ID = createdUser.ID
		err = repo.Update(ctx, userToUpdate)
		if err != nil {
			t.Fatalf("failed to update user: %v", err)
		}
		updatedUser, err := repo.GetByID(ctx, userToUpdate.ID)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if updatedUser.ID != userToUpdate.ID || updatedUser.Username != userToUpdate.Username || updatedUser.Email != userToUpdate.Email || updatedUser.PasswordHash != userToUpdate.PasswordHash {
			t.Fatalf("user data is not as expected, got %v, want %v", *updatedUser, *createdUser)
		}
		t.Cleanup(func() { _ = repo.Delete(ctx, userToUpdate.ID) })
	})
	t.Run("user not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		userToUpdate, _ := domain.NewUser("johndoe67", "user67@example.com", "passwordHash12367")
		err := repo.Update(ctx, userToUpdate)
		if !errors.Is(err, postgres.ErrNoRows) {
			t.Fatalf("wanted no rows affected error, got: %v", err)
		}
	})
}

func TestUserRepository_Delete(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		u, err := repo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		err = repo.Delete(ctx, u.ID)
		if err != nil {
			t.Fatalf("failed to delete user: %v", err)
		}
		_, err = repo.GetByID(ctx, u.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
	t.Run("user not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewUserRepository(db)

		err := repo.Delete(ctx, 67)
		if err != nil {
			t.Fatalf("failed to delete not found user, wanted nil error, got: %v", err)
		}
	})
}

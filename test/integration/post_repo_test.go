package integration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/repository/postgres"
)

func TestPostRepository_Create(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		userRepo := postgres.NewUserRepository(db)
		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		post, err := domain.NewPost(createdUser.ID, "Test Post Title", "Test Post Content")
		if err != nil {
			t.Fatalf("failed to create post domain: %v", err)
		}
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}
		if createdPost.ID == 0 {
			t.Fatal("post ID is 0")
		}
		if createdPost.CreatedAt.IsZero() {
			t.Fatal("created at is zero")
		}
		if createdPost.UpdatedAt.IsZero() {
			t.Fatal("updated at is zero")
		}
		if createdPost.UserID != post.UserID || createdPost.Title != post.Title || createdPost.Content != post.Content {
			t.Fatalf("post data is not as expected, got %v, want %v", *createdPost, *post)
		}
		_, _ = repo.Delete(ctx, createdPost.ID)
	})
}

func TestPostRepository_GetByID(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		post, _ := domain.NewPost(createdUser.ID, "Test Post Title", "Test Post Content")
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		p, err := repo.GetByID(ctx, createdPost.ID)
		if err != nil {
			t.Fatalf("failed to get post: %v", err)
		}
		if p.ID != createdPost.ID || p.UserID != createdPost.UserID || p.Title != createdPost.Title || p.Content != createdPost.Content {
			t.Fatalf("post data is not as expected, got %v, want %v", *p, *createdPost)
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, p.ID) })
	})
	t.Run("post not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)

		_, err := repo.GetByID(ctx, 67)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
}

func TestPostRepository_GetByTitle(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		post, _ := domain.NewPost(createdUser.ID, "Unique Title", "Test Post Content")
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		p, err := repo.GetByTitle(ctx, "Unique Title")
		if err != nil {
			t.Fatalf("failed to get post by title: %v", err)
		}
		if p.ID != createdPost.ID || p.UserID != createdPost.UserID || p.Title != createdPost.Title || p.Content != createdPost.Content {
			t.Fatalf("post data is not as expected, got %v, want %v", *p, *createdPost)
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, p.ID) })
	})
	t.Run("post not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)

		_, err := repo.GetByTitle(ctx, "non-existent-title")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
}

func TestPostRepository_ReadAllUserPosts(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		post1, _ := domain.NewPost(createdUser.ID, "Post Title 1", "Content 1")
		createdPost1, err := repo.Create(ctx, post1)
		if err != nil {
			t.Fatalf("failed to create post 1: %v", err)
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, createdPost1.ID) })

		post2, _ := domain.NewPost(createdUser.ID, "Post Title 2", "Content 2")
		createdPost2, err := repo.Create(ctx, post2)
		if err != nil {
			t.Fatalf("failed to create post 2: %v", err)
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, createdPost2.ID) })

		posts, err := repo.ReadAllUserPosts(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("failed to read all user posts: %v", err)
		}
		if len(posts) < 2 {
			t.Fatalf("expected at least 2 posts, got %d", len(posts))
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, createdPost2.ID) })
	})
	t.Run("user with no posts", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		posts, err := repo.ReadAllUserPosts(ctx, createdUser.ID)
		if err != nil {
			t.Fatalf("failed to read all user posts: %v", err)
		}
		if len(posts) != 0 {
			t.Fatalf("expected 0 posts, got %d", len(posts))
		}
	})
}

func TestPostRepository_Update(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		post, _ := domain.NewPost(createdUser.ID, "Original Title", "Original Content")
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		postToUpdate := &domain.Post{
			ID:      createdPost.ID,
			UserID:  createdUser.ID,
			Title:   "Updated Title",
			Content: "Updated Content",
		}
		err = repo.Update(ctx, postToUpdate)
		if err != nil {
			t.Fatalf("failed to update post: %v", err)
		}
		updatedPost, err := repo.GetByID(ctx, postToUpdate.ID)
		if err != nil {
			t.Fatalf("failed to get post: %v", err)
		}
		if updatedPost.Title != postToUpdate.Title || updatedPost.Content != postToUpdate.Content {
			t.Fatalf("post data is not as expected, got %v, want %v", *updatedPost, *postToUpdate)
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, updatedPost.ID) })
	})
	t.Run("post not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)

		postToUpdate := &domain.Post{
			ID:      999,
			UserID:  1,
			Title:   "Updated Title",
			Content: "Updated Content",
		}
		err := repo.Update(ctx, postToUpdate)
		if !errors.Is(err, postgres.ErrNoRows) {
			t.Fatalf("wanted no rows affected error, got: %v", err)
		}
	})
}

func TestPostRepository_UpdateWithValidate(t *testing.T) {
	t.Run("happy path - owner updates", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		post, _ := domain.NewPost(createdUser.ID, "Original Title", "Original Content")
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		postToUpdate := &domain.Post{
			ID:      createdPost.ID,
			UserID:  createdUser.ID,
			Title:   "Updated Title",
			Content: "Updated Content",
		}
		err = repo.UpdateWithValidate(ctx, createdUser.ID, postToUpdate)
		if err != nil {
			t.Fatalf("failed to update post: %v", err)
		}
		updatedPost, err := repo.GetByID(ctx, postToUpdate.ID)
		if err != nil {
			t.Fatalf("failed to get post: %v", err)
		}
		if updatedPost.Title != postToUpdate.Title || updatedPost.Content != postToUpdate.Content {
			t.Fatalf("post data is not as expected, got %v, want %v", *updatedPost, *postToUpdate)
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, updatedPost.ID) })
	})
	t.Run("non-owner cannot update", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser1, _ := domain.NewUser(suffix+"user1", suffix+"user1@example.com", "passwordHash123")
		createdUser1, err := userRepo.Create(ctx, prepUser1)
		if err != nil {
			t.Fatalf("failed to create user 1: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser1.ID) })

		prepUser2, _ := domain.NewUser(suffix+"user2", suffix+"user2@example.com", "passwordHash123")
		createdUser2, err := userRepo.Create(ctx, prepUser2)
		if err != nil {
			t.Fatalf("failed to create user 2: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser2.ID) })

		post, _ := domain.NewPost(createdUser1.ID, "Original Title", "Original Content")
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		postToUpdate := &domain.Post{
			ID:      createdPost.ID,
			UserID:  createdUser1.ID,
			Title:   "Updated Title",
			Content: "Updated Content",
		}
		err = repo.UpdateWithValidate(ctx, createdUser2.ID, postToUpdate)
		if !errors.Is(err, postgres.ErrNoRows) {
			t.Fatalf("wanted no rows affected error (non-owner), got: %v", err)
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, createdPost.ID) })
	})
	t.Run("post not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)

		postToUpdate := &domain.Post{
			ID:      999,
			UserID:  1,
			Title:   "Updated Title",
			Content: "Updated Content",
		}
		err := repo.UpdateWithValidate(ctx, 1, postToUpdate)
		if !errors.Is(err, postgres.ErrNoRows) {
			t.Fatalf("wanted no rows affected error, got: %v", err)
		}
	})
}

func TestPostRepository_Delete(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		post, _ := domain.NewPost(createdUser.ID, "Post Title", "Post Content")
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		title, err := repo.Delete(ctx, createdPost.ID)
		if err != nil {
			t.Fatalf("failed to delete post: %v", err)
		}
		if title != createdPost.Title {
			t.Fatalf("expected title %q, got %q", createdPost.Title, title)
		}
		_, err = repo.GetByID(ctx, createdPost.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
	t.Run("post not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)

		_, err := repo.Delete(ctx, 67)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
}

func TestPostRepository_DeleteWithValidate(t *testing.T) {
	t.Run("happy path - owner deletes", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser, _ := domain.NewUser(suffix+"johndoe", suffix+"user@example.com", "passwordHash123")
		createdUser, err := userRepo.Create(ctx, prepUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser.ID) })

		post, _ := domain.NewPost(createdUser.ID, "Post Title", "Post Content")
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		title, err := repo.DeleteWithValidate(ctx, createdUser.ID, createdPost.ID)
		if err != nil {
			t.Fatalf("failed to delete post: %v", err)
		}
		if title != createdPost.Title {
			t.Fatalf("expected title %q, got %q", createdPost.Title, title)
		}
		_, err = repo.GetByID(ctx, createdPost.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
	t.Run("non-owner cannot delete", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)
		userRepo := postgres.NewUserRepository(db)

		suffix := fmt.Sprintf("%d", time.Now().UnixNano())

		prepUser1, _ := domain.NewUser(suffix+"user1", suffix+"user1@example.com", "passwordHash123")
		createdUser1, err := userRepo.Create(ctx, prepUser1)
		if err != nil {
			t.Fatalf("failed to create user 1: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser1.ID) })

		prepUser2, _ := domain.NewUser(suffix+"user2", suffix+"user2@example.com", "passwordHash123")
		createdUser2, err := userRepo.Create(ctx, prepUser2)
		if err != nil {
			t.Fatalf("failed to create user 2: %v", err)
		}
		t.Cleanup(func() { _ = userRepo.Delete(ctx, createdUser2.ID) })

		post, _ := domain.NewPost(createdUser1.ID, "Post Title", "Post Content")
		createdPost, err := repo.Create(ctx, post)
		if err != nil {
			t.Fatalf("failed to create post: %v", err)
		}

		_, err = repo.DeleteWithValidate(ctx, createdUser2.ID, createdPost.ID)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error (non-owner), got: %v", err)
		}
		t.Cleanup(func() { _, _ = repo.Delete(ctx, createdPost.ID) })
	})
	t.Run("post not found", func(t *testing.T) {
		db := testDB.DB
		ctx := context.Background()

		repo := postgres.NewPostRepository(db)

		_, err := repo.DeleteWithValidate(ctx, 1, 67)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("wanted no rows error, got: %v", err)
		}
	})
}

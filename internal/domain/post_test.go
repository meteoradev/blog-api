package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/meteoradev/BlogApi/internal/domain"
)

func TestNewPost(t *testing.T) {
	// negative
	tests := []struct {
		testName string
		expErr   error
		userID   int64
		title    string
		content  string
	}{
		{"empty title", domain.ErrInvalidTitle, 1, "", "content"},
		{"too long title", domain.ErrInvalidTitle, 1, strings.Repeat("title", 100), "content"},
		{"empty content", domain.ErrInvalidContent, 1, "title", ""},
		{"too long content", domain.ErrInvalidContent, 1, "title", strings.Repeat("content", 1000)},
		{"invalid user id", domain.ErrInvalidUserId, -10, "title", "content"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, err := domain.NewPost(tt.userID, tt.title, tt.content)
			if !errors.Is(err, tt.expErr) {
				t.Errorf("NewPost() negative case got error = %e, want = %e", err, tt.expErr)
			}
		})
	}
	t.Run("happy path", func(t *testing.T) {
		post, err := domain.NewPost(1, "title", "content")
		if err != nil {
			t.Errorf("NewPost() positive cases got error = %e, want = %v", err, nil)
		}
		if post.UserID != 1 || post.Title != "title" || post.Content != "content" {
			t.Errorf("NewPost() positive case got = %v, want = %v", post, domain.Post{0, 1, "title", "content", time.Now(), time.Now()})
		}
	})
}

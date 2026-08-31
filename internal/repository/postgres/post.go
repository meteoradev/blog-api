package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/jmoiron/sqlx"
)

type dbPost struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r dbPost) toDomain() *domain.Post {
	return &domain.Post{
		ID:        r.ID,
		UserID:    r.UserID,
		Title:     r.Title,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

type PostRepository struct {
	db *sqlx.DB
}

func NewPostRepository(db *sqlx.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) (*domain.Post, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
        INSERT INTO posts (user_id, title, content)
        VALUES ($1, $2, $3)
        RETURNING *
    `
	var dbPost dbPost
	err = tx.GetContext(txCtx, &dbPost, query, post.UserID, post.Title, post.Content)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return dbPost.toDomain(), nil
}

func (r *PostRepository) GetByID(ctx context.Context, postID int64) (*domain.Post, error) {
	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
        SELECT * FROM posts
        WHERE id = $1
    `
	var post dbPost
	err := r.db.GetContext(txCtx, &post, query, postID)
	if err != nil {
		return nil, err
	}
	return post.toDomain(), nil
}

func (r *PostRepository) GetByTitle(ctx context.Context, title string) (*domain.Post, error) {
	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
        SELECT * FROM posts
        WHERE title = $1
    `
	var post dbPost
	err := r.db.GetContext(txCtx, &post, query, title)
	if err != nil {
		return nil, err
	}
	return post.toDomain(), nil
}

func (r *PostRepository) ReadAllUserPosts(ctx context.Context, userID int64) ([]*domain.Post, error) {
	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
		SELECT * FROM posts
		WHERE user_id = $1
	`
	var dbPosts []dbPost
	err := r.db.SelectContext(txCtx, &dbPosts, query, userID)
	if err != nil {
		return nil, err
	}
	posts := make([]*domain.Post, len(dbPosts))
	for i, p := range dbPosts {
		posts[i] = p.toDomain()
	}
	return posts, nil
}

func (r *PostRepository) Update(ctx context.Context, post *domain.Post) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
		UPDATE posts SET title = $1, content = $2
		WHERE id = $3
	`
	result, err := tx.ExecContext(txCtx, query, post.Title, post.Content, post.ID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNoRows
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) UpdateWithValidate(ctx context.Context, currUserID int64, post *domain.Post) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `UPDATE posts SET title = $1, content = $2
			  WHERE id = $3 AND user_id = $4`
	result, err := tx.ExecContext(txCtx, query, post.Title, post.Content, post.ID, currUserID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNoRows
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) Delete(ctx context.Context, postID int64) (string, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `DELETE FROM posts WHERE id = $1 RETURNING title`
	var title string
	err = tx.QueryRowContext(txCtx, query, postID).Scan(&title)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", err
		}
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}
	return title, nil
}
func (r *PostRepository) DeleteWithValidate(ctx context.Context, currUserID, ID int64) (string, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `DELETE FROM posts WHERE id = $1 AND user_id = $2 RETURNING title`
	var title string
	err = tx.QueryRowContext(txCtx, query, ID, currUserID).Scan(&title)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", err
		}
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}
	return title, nil
}

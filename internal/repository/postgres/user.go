package postgres

import (
	"context"
	"time"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/jmoiron/sqlx"
)

type dbUser struct {
	ID           int64     `db:"id"`
	Email        string    `db:"email"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

func (r *dbUser) toDomain() *domain.User {
	return &domain.User{
		ID:           r.ID,
		Username:     r.Username,
		Email:        r.Email,
		PasswordHash: r.PasswordHash,
		CreatedAt:    r.CreatedAt,
	}
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
        INSERT INTO users (username, email, password_hash)
        VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
        RETURNING *
    `
	var userRow dbUser
	err = tx.GetContext(txCtx, &userRow, query, user.Username, user.Email, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return userRow.toDomain(), nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
        SELECT * FROM users
        WHERE id = $1
    `
	var user dbUser
	err := r.db.GetContext(txCtx, &user, query, userID)
	if err != nil {
		return nil, err
	}
	return user.toDomain(), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
        SELECT * FROM users
        WHERE email = $1
    `
	var user dbUser
	err := r.db.GetContext(txCtx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return user.toDomain(), nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
		UPDATE users SET username = $1, email = $2, password_hash = $3
		WHERE id = $4
	`
	result, err := tx.ExecContext(txCtx, query, user.Username, user.Email, user.PasswordHash, user.ID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNoRows
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, userID int64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `
		DELETE FROM users WHERE id = $1
	`
	_, err = tx.ExecContext(txCtx, query, userID)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

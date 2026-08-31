package service

import (
	"context"
	"database/sql"

	"github.com/meteoradev/BlogApi/internal/domain"
	"github.com/meteoradev/BlogApi/internal/port"
	"github.com/rs/zerolog/log"
)

type UserService struct {
	userRepo port.UserRepository
	hasher   port.Hasher
}

func NewUserService(userRepo port.UserRepository, hasher port.Hasher) *UserService {
	return &UserService{userRepo: userRepo, hasher: hasher}
}

func (u *UserService) Create(ctx context.Context, username, email, password string) (*domain.User, error) {
	passwordHash, err := processPassword(password, u.hasher)
	if err != nil {
		return nil, err
	}

	domainUser, err := domain.NewUser(username, email, passwordHash)
	if err != nil {
		return nil, err
	}

	user, err := u.userRepo.Create(ctx, domainUser)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrUserAlreadyExists
		default:
			return nil, err
		}

	}
	logger := log.Ctx(ctx)
	trace_id, _ := ctx.Value("trace_id").(string)
	logger.Info().
		Str("trace_id", trace_id).
		Int64("userId", user.ID).
		Str("username", username).
		Str("email", email).
		Msg("User created.")
	return user, nil
}

func (u *UserService) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrUserNotFound
		default:
			return nil, ErrUnexpected
		}
	}
	logger := log.Ctx(ctx)
	trace_id, _ := ctx.Value("trace_id").(string)
	logger.Debug().
		Str("trace_id", trace_id).
		Int64("userId", userID).
		Msg("Read user.")
	return user, nil
}

func (u *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrUserNotFound
		default:
			return nil, ErrUnexpected
		}
	}
	logger := log.Ctx(ctx)
	trace_id, _ := ctx.Value("trace_id").(string)
	logger.Debug().
		Str("trace_id", trace_id).
		Str("title", email).
		Msg("Read user.")
	return user, nil
}

func (u *UserService) Update(ctx context.Context, currUserID, userID int64, username, email, password string) error {
	if ok := u.validateCurrUser(currUserID, userID); !ok {
		return ErrMethodNotAllowed
	}
	passwordHash, err := processPassword(password, u.hasher)
	if err != nil {
		return err
	}
	domainUser, err := domain.NewUser(username, email, passwordHash)
	if err != nil {
		return err
	}
	err = u.userRepo.Update(ctx, domainUser)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrUserNotFound
		case sql.ErrTxDone:
			return ErrUserTimeOut
		default:
			return ErrUnexpected
		}
	}
	logger := log.Ctx(ctx)
	trace_id, _ := ctx.Value("trace_id").(string)
	logger.Debug().
		Str("trace_id", trace_id).
		Int64("user_id", userID).
		Str("username", username).
		Str("email", email).
		Msg("Update user.")
	return nil
}

func (u *UserService) Delete(ctx context.Context, currUserID int64, userID int64) error {
	if ok := u.validateCurrUser(currUserID, userID); !ok {
		return ErrMethodNotAllowed
	}
	err := u.userRepo.Delete(ctx, userID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrUserNotFound
		case sql.ErrTxDone:
			return ErrUserTimeOut
		default:
			return ErrUnexpected
		}
	}
	logger := log.Ctx(ctx)
	trace_id, _ := ctx.Value("trace_id").(string)
	logger.Debug().
		Str("trace_id", trace_id).
		Int64("user_id", userID).
		Msg("Delete user.")
	return nil
}
func (u *UserService) validateCurrUser(currUserID, userID int64) bool {
	if userID != currUserID {
		return false
	}
	return true
}

func processPassword(pass string, hasher port.Hasher) (string, error) {
	if pass == "" {
		return "", ErrEmptyPassword
	}
	if len(pass) < 8 || len(pass) > 60 {
		return "", ErrInvalidPasswordLength
	}
	pass, err := hasher.Hash(pass)
	if err != nil {
		return "", ErrCanNotCalculatePassHash
	}
	return pass, nil
}

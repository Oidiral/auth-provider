package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserRepo struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewUserRepo(db *sqlx.DB, log logger.Logger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: log,
	}
}

func (r *UserRepo) Create(ctx context.Context, input domain.User) (string, error) {
	log := r.logger.WithContext(ctx)
	log.Info("creating user", logger.Field{Key: "email", Value: input.Email})

	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO users (email, password_hash, first_name, last_name, username) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		input.Email, input.PasswordHash, input.FirstName, input.LastName, input.Username,
	).Scan(&input.ID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				log.Warn("user already exists", logger.Field{Key: "email", Value: input.Email})
				return "", domain.ErrUserAlreadyExists
			case "23502":
				log.Warn("invalid input for user creation", logger.Field{Key: "email", Value: input.Email})
				return "", domain.ErrInvalidInput
			default:
				log.Error("database error creating user", err, logger.Field{Key: "email", Value: input.Email}, logger.Field{Key: "code", Value: pqErr.Code})
				return "", fmt.Errorf("%w: %v", domain.ErrInternalServer, pqErr.Code)
			}
		}
		log.Error("failed to create user", err, logger.Field{Key: "email", Value: input.Email})
		return "", fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}

	log.Info("user created successfully", logger.Field{Key: "email", Value: input.Email}, logger.Field{Key: "user_id", Value: input.ID})
	return input.ID, nil
}

func (r *UserRepo) Get(ctx context.Context, id string) (domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23503":
				return user, domain.ErrUserNotFound
			case "23502":
				return user, domain.ErrInvalidInput
			default:
				return user, fmt.Errorf("%w: %v", domain.ErrInternalServer, pqErr.Code)
			}
		}
		return user, fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23503":
				return user, domain.ErrUserNotFound
			case "23502":
				return user, domain.ErrInvalidInput
			default:
				return user, fmt.Errorf("%w: %v", domain.ErrInternalServer, pqErr.Code)
			}
		}
		return user, fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return user, nil
}

func (r *UserRepo) Update(ctx context.Context, id string, input domain.UpdateUserInput) error {
	log := r.logger.WithContext(ctx)
	log.Info("updating user", logger.Field{Key: "user_id", Value: id})

	query := "UPDATE users SET updated_at = NOW()"
	var args []interface{}
	argCount := 1

	if input.Email != nil {
		query += fmt.Sprintf(", email = $%d", argCount)
		args = append(args, *input.Email)
		argCount++
	}
	if input.Username != nil {
		query += fmt.Sprintf(", username = $%d", argCount)
		args = append(args, *input.Username)
		argCount++
	}
	if input.FirstName != nil {
		query += fmt.Sprintf(", first_name = $%d", argCount)
		args = append(args, *input.FirstName)
		argCount++
	}
	if input.LastName != nil {
		query += fmt.Sprintf(", last_name = $%d", argCount)
		args = append(args, *input.LastName)
		argCount++
	}
	if input.Phone != nil {
		query += fmt.Sprintf(", phone = $%d", argCount)
		args = append(args, *input.Phone)
		argCount++
	}
	if input.PasswordHash != nil {
		query += fmt.Sprintf(", password_hash = $%d", argCount)
		args = append(args, *input.PasswordHash)
		argCount++
	}
	if input.IsVerified != nil {
		query += fmt.Sprintf(", is_verified = $%d", argCount)
		args = append(args, *input.IsVerified)
		argCount++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCount)
	args = append(args, id)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				log.Warn("duplicate user data on update", logger.Field{Key: "user_id", Value: id})
				return domain.ErrUserAlreadyExists
			case "23502":
				log.Warn("invalid input for user update", logger.Field{Key: "user_id", Value: id})
				return domain.ErrInvalidInput
			default:
				log.Error("database error updating user", err, logger.Field{Key: "user_id", Value: id}, logger.Field{Key: "code", Value: pqErr.Code})
				return fmt.Errorf("%w: %v", domain.ErrInternalServer, pqErr.Code)
			}
		}
		log.Error("failed to update user", err, logger.Field{Key: "user_id", Value: id})
		return fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error("failed to get rows affected", err, logger.Field{Key: "user_id", Value: id})
		return fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	if rowsAffected == 0 {
		log.Warn("user not found for update", logger.Field{Key: "user_id", Value: id})
		return domain.ErrUserNotFound
	}

	log.Info("user updated successfully", logger.Field{Key: "user_id", Value: id})
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	log := r.logger.WithContext(ctx)
	log.Info("deleting user", logger.Field{Key: "user_id", Value: id})

	if id == "" {
		log.Warn("empty user id provided for deletion")
		return domain.ErrInvalidUserID
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23503":
				log.Warn("user has dependencies, cannot delete", logger.Field{Key: "user_id", Value: id})
				return domain.ErrUserHasDependencies
			case "23502":
				log.Warn("invalid input for user deletion", logger.Field{Key: "user_id", Value: id})
				return domain.ErrInvalidInput
			case "22P02":
				log.Warn("invalid user id format", logger.Field{Key: "user_id", Value: id})
				return domain.ErrInvalidUserID
			default:
				log.Error("database error deleting user", err, logger.Field{Key: "user_id", Value: id}, logger.Field{Key: "code", Value: pqErr.Code})
				return fmt.Errorf("%w: %v", domain.ErrInternalServer, pqErr.Code)
			}
		}
		log.Error("failed to delete user", err, logger.Field{Key: "user_id", Value: id})
		return fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error("failed to get rows affected", err, logger.Field{Key: "user_id", Value: id})
		return fmt.Errorf("%w: failed to get rows affected", domain.ErrInternalServer)
	}
	if rowsAffected == 0 {
		log.Warn("user not found for deletion", logger.Field{Key: "user_id", Value: id})
		return domain.ErrUserNotFound
	}

	log.Info("user deleted successfully", logger.Field{Key: "user_id", Value: id})
	return nil
}

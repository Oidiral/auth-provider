package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/lib/pq"
)

type RoleRepo struct {
	db  DBTX
	log logger.Logger
}

func NewRole(db DBTX, log logger.Logger) *RoleRepo {
	return &RoleRepo{db: db, log: log}
}

func (r *RoleRepo) GetById(ctx context.Context, id int) (domain.Role, error) {
	var role domain.Role
	err := r.db.GetContext(ctx, &role, "SELECT * FROM roles WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return role, domain.ErrRoleNotFound
		}
		return role, fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return role, nil
}

func (r *RoleRepo) GetByName(ctx context.Context, name string) (domain.Role, error) {
	var role domain.Role
	err := r.db.GetContext(ctx, &role, "SELECT * FROM roles WHERE name = $1", name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return role, domain.ErrRoleNotFound
		}
		return role, fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return role, nil
}

func (r *RoleRepo) GetUserRoles(ctx context.Context, userId string) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.SelectContext(ctx, &roles,
		"SELECT roles.* FROM user_roles JOIN roles ON user_roles.role_id = roles.id WHERE user_roles.user_id = $1",
		userId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return roles, nil
}

func (r *RoleRepo) AssignToUser(ctx context.Context, userId string, roleId int) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)", userId, roleId)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return domain.ErrRoleAlreadyAssigned
			case "23503":
				return domain.ErrRoleNotFound
			}
		}
		return fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return nil
}

func (r *RoleRepo) RemoveFromUser(ctx context.Context, userId string, roleId int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2", userId, roleId)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return nil
}

func (r *RoleRepo) UserHasRole(ctx context.Context, userId string, roleId int) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)",
		userId, roleId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return exists, nil
}

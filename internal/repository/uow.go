package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Oidiral/auth-provider/internal/domain"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/jmoiron/sqlx"
)

type DBTX interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type unitOfWork struct {
	tx    *sqlx.Tx
	users Users
	roles Roles
	log   logger.Logger
	done  bool
}

func (u *unitOfWork) Users() Users { return u.users }
func (u *unitOfWork) Roles() Roles { return u.roles }

func (u *unitOfWork) Commit() error {
	if u.done {
		return nil
	}
	u.done = true
	if err := u.tx.Commit(); err != nil {
		u.log.Error("failed to commit transaction", err)
		return fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return nil
}

func (u *unitOfWork) Rollback() error {
	if u.done {
		return nil
	}
	u.done = true
	if err := u.tx.Rollback(); err != nil {
		u.log.Error("failed to rollback transaction", err)
		return fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}
	return nil
}

type uowFactory struct {
	db  *sqlx.DB
	log logger.Logger
}

func NewUoWFactory(db *sqlx.DB, log logger.Logger) UoWFactory {
	return &uowFactory{db: db, log: log}
}

func (f *uowFactory) Begin(ctx context.Context) (UnitOfWork, error) {
	tx, err := f.db.BeginTxx(ctx, nil)
	if err != nil {
		f.log.Error("failed to begin transaction", err)
		return nil, fmt.Errorf("%w: %v", domain.ErrInternalServer, err)
	}

	return &unitOfWork{
		tx:    tx,
		users: NewUserRepo(tx, f.log),
		roles: NewRole(tx, f.log),
		log:   f.log,
		done:  false,
	}, nil
}

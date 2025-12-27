package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var FS embed.FS

type MigrationsConfig struct {
	Timeout time.Duration
}

func RunMigrations(db *sql.DB, log logger.Logger, cfg MigrationsConfig) error {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	goose.SetBaseFS(FS)

	goose.SetLogger(NewGooseLogger(log))

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info("Migrations applied successfully")
	return nil
}

type GooseLogger struct {
	log logger.Logger
}

func NewGooseLogger(log logger.Logger) *GooseLogger {
	return &GooseLogger{log: log}
}

func (gl *GooseLogger) Fatalf(format string, v ...interface{}) {
	gl.log.Fatal(fmt.Sprintf(format, v...))
}

func (gl *GooseLogger) Printf(format string, v ...interface{}) {
	gl.log.Info(fmt.Sprintf(format, v...))
}

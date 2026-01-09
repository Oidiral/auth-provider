package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const timeout = 10 * time.Second

func NewClient(host string, port string, user string, password string, dbname string, log logger.Logger) (*sqlx.DB, error) {
	log.Info("connecting to postgres", logger.Field{Key: "host", Value: host}, logger.Field{Key: "port", Value: port}, logger.Field{Key: "database", Value: dbname})

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		log.Error("failed to ping postgres", err)
		db.Close()
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Minute * 3)

	log.Info("successfully connected to postgres")
	return db, nil
}

func GetSqlDB(sqlxDB *sqlx.DB) *sql.DB {
	return sqlxDB.DB
}

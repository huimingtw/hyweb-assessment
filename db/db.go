package db

import (
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func New(dsn string, logger *slog.Logger) (*sql.DB, error) {
	database, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, err
	}

	logger.Info("database connected")
	return database, nil
}

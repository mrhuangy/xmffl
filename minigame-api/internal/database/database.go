package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

func Wait(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastErr error
	for {
		if err := db.PingContext(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return lastErr
		case <-ticker.C:
		}
	}
}

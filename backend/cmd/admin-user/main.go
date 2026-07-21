package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	username := requiredEnv("ADMIN_USERNAME")
	password := requiredEnv("ADMIN_PASSWORD")
	displayName := env("ADMIN_DISPLAY_NAME", "系统管理员")
	role := env("ADMIN_ROLE", "owner")
	if len(password) < 10 {
		log.Fatal("ADMIN_PASSWORD must contain at least 10 characters")
	}
	if role != "owner" && role != "operator" && role != "viewer" {
		log.Fatal("ADMIN_ROLE must be owner, operator, or viewer")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}
	db, err := sql.Open("mysql", env("MYSQL_DSN", "fpxxl:fpxxl@tcp(127.0.0.1:3306)/fpxxl?parseTime=true&loc=Local"))
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `INSERT INTO admin_users (username, password_hash, display_name, role, status, password_changed_at)
		VALUES (?, ?, ?, ?, 'active', CURRENT_TIMESTAMP)
		ON DUPLICATE KEY UPDATE password_hash = VALUES(password_hash), display_name = VALUES(display_name),
		role = VALUES(role), status = 'active', failed_login_attempts = 0, locked_until = NULL,
		password_changed_at = CURRENT_TIMESTAMP`, username, string(hash), displayName, role)
	if err != nil {
		log.Fatalf("save admin: %v", err)
	}
	fmt.Printf("administrator %q is ready\n", username)
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

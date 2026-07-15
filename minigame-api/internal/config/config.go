package config

import "os"

type Config struct {
	HTTPAddr    string
	MySQLDSN    string
	AllowOrigin string
	GinMode     string
}

func Load() Config {
	return Config{
		HTTPAddr:    env("HTTP_ADDR", ":8090"),
		MySQLDSN:    env("MYSQL_DSN", "fpxxl:fpxxl@tcp(127.0.0.1:3306)/fpxxl?parseTime=true&loc=Local"),
		AllowOrigin: env("ALLOW_ORIGIN", "*"),
		GinMode:     env("GIN_MODE", "debug"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

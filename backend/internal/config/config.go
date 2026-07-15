package config

import "os"

type Config struct {
	HTTPAddr    string
	MySQLDSN    string
	AllowOrigin string
}

func Load() Config {
	return Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8080"),
		MySQLDSN:    getEnv("MYSQL_DSN", "fpxxl:fpxxl@tcp(127.0.0.1:3306)/fpxxl?parseTime=true&loc=Local"),
		AllowOrigin: getEnv("ALLOW_ORIGIN", "*"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

# FPXXL Backend

Go API for remote level configuration, ad frequency configuration, player progress, leaderboard, and event ingestion.

## Run

```bash
go run ./cmd/api
```

Environment variables:

- `HTTP_ADDR`: listen address, default `:8080`
- `MYSQL_DSN`: MySQL DSN, default `fpxxl:fpxxl@tcp(127.0.0.1:3306)/fpxxl?parseTime=true&loc=Local`
- `ALLOW_ORIGIN`: CORS origin, default `*`

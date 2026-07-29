# FPXXL Laravel APIs

This directory contains Laravel replacements for both Go APIs:

- `admin-api`: admin and legacy endpoints, served through `admin.fpxxl.local/api`
  and the compatible standalone host `admin-api.fpxxl.local`
- `minigame-api`: minigame endpoints, served as `api.fpxxl.local`
- `dockerconfig`: Nginx, PHP-FPM 8.2, MySQL 8 and Redis

The HTTP paths and JSON field names match `backend/` and `minigame-api/`. Both
applications use the same `fpxxl` database. MySQL initializes from
`docker_config/deploy/mysql`, including all existing schema upgrades and seeds.

## Start

Docker Desktop must be running.

```powershell
cd php/dockerconfig
docker compose up -d --build
docker compose exec admin_php php artisan admin:create admin "change-this-password" --name=Owner
```

Add local DNS entries when testing the production host names:

```text
127.0.0.1 admin.fpxxl.local
127.0.0.1 admin-api.fpxxl.local
127.0.0.1 api.fpxxl.local
```

Health checks:

```powershell
curl http://admin-api.fpxxl.local/healthz
curl http://admin.fpxxl.local/api/config/ads
curl http://api.fpxxl.local/healthz
```

Configuration uses `FPXXL_DB_DATABASE`, `FPXXL_DB_USERNAME`,
`FPXXL_DB_PASSWORD`, `FPXXL_JWT_SECRET`, `WECHAT_APP_ID`, and
`WECHAT_APP_SECRET`. When WeChat credentials are empty, login keeps the Go
service's deterministic development OpenID behavior.

For an existing MySQL volume, initialization scripts are not run again. Apply
the SQL files in `docker_config/deploy/mysql` manually or recreate only the
dedicated `fpxxl-php_mysql_data` volume after backing up its data.

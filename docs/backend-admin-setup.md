# 后端与后台环境

本项目新增独立后端与后台：

- `backend/`：Go HTTP API，使用 MySQL 持久化配置、进度、排行榜和事件。
- `admin/`：Vue 3 运营后台，当前包含关卡配置和广告频控。
- `deploy/mysql/`：MySQL 建表与初始化数据。
- `docker-compose.yml`：本地一键启动 MySQL、后端和后台。

## 本地启动

```bash
docker compose up --build
```

服务地址：

- 后台页面：http://localhost:5173
- 后端健康检查：http://localhost:8080/healthz
- MySQL：localhost:3306，账号 `fpxxl`，密码 `fpxxl`，库名 `fpxxl`

## 后端接口

| 接口 | 方法 | 说明 |
| --- | --- | --- |
| `/api/auth/login` | POST | 管理员登录，返回有效期 8 小时的 Bearer Token |
| `/api/auth/me` | GET | 获取当前管理员，需要 Bearer Token |
| `/api/config/levels` | GET | 获取已启用关卡配置 |
| `/api/config/levels?includeDisabled=true` | GET | 后台获取全部关卡配置 |
| `/api/admin/levels/{levelId}` | PUT | 新增或更新关卡配置 |
| `/api/config/ads` | GET | 获取广告频控配置 |
| `/api/admin/config/ads` | PUT | 更新广告频控配置 |
| `/api/admin/users` | GET/POST | 查询或创建后台管理员，仅 owner 可用 |
| `/api/admin/users/{id}` | GET/PUT/DELETE | 查看、修改或删除后台管理员，仅 owner 可用 |
| `/api/admin/system-controls` | GET/POST | 查询或创建特殊配置，仅 owner 可用 |
| `/api/admin/system-controls/{id}` | GET/PUT/DELETE | 查看、修改或删除特殊配置，仅 owner 可用 |
| `/api/player/progress` | GET/POST | 同步玩家进度，使用 `openId` 参数或 `X-Openid` 头 |
| `/api/player/level-results` | POST | 上报关卡结算结果 |
| `/api/leaderboard/submit` | POST | 提交排行榜成绩 |
| `/api/leaderboard` | GET | 查询排行榜，可传 `levelId` 和 `limit` |
| `/api/events/batch` | POST | 批量上报玩法、广告和奖励事件 |

`/api/admin/*` 以及带 `includeDisabled=true` 的完整关卡配置接口必须发送
`Authorization: Bearer <token>`。连续 5 次密码错误会锁定账号 15 分钟。

管理员管理接口禁止删除当前登录账号，并确保系统至少保留一个状态正常的 `owner`。
创建管理员时密码至少 10 位；更新时密码留空表示保持原密码不变。

特殊配置接口按照 `value_type` 校验数据：`json` 使用 `value_json`，其他类型使用
`value_text`。配置修改时 `version` 自动递增，并记录创建人和最后修改人。

## 创建管理员

管理员密码使用 bcrypt 存储。通过环境变量调用命令创建或重置管理员：

```bash
cd backend
ADMIN_USERNAME=admin \
ADMIN_PASSWORD='replace-with-a-strong-password' \
ADMIN_DISPLAY_NAME='系统管理员' \
ADMIN_ROLE=owner \
go run ./cmd/admin-user
```

生产环境必须配置随机生成的 `ADMIN_JWT_SECRET`，不得使用 Compose 中的本地开发默认值。

## 单独开发

后端：

```bash
cd backend
go mod tidy
go run ./cmd/api
```

后台：

```bash
cd admin
npm install
npm run dev
```

后台开发服务器会把 `/api` 代理到 `http://localhost:8080`。

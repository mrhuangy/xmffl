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
| `/api/config/levels` | GET | 获取已启用关卡配置 |
| `/api/config/levels?includeDisabled=true` | GET | 后台获取全部关卡配置 |
| `/api/admin/levels/{levelId}` | PUT | 新增或更新关卡配置 |
| `/api/config/ads` | GET | 获取广告频控配置 |
| `/api/admin/config/ads` | PUT | 更新广告频控配置 |
| `/api/player/progress` | GET/POST | 同步玩家进度，使用 `openId` 参数或 `X-Openid` 头 |
| `/api/player/level-results` | POST | 上报关卡结算结果 |
| `/api/leaderboard/submit` | POST | 提交排行榜成绩 |
| `/api/leaderboard` | GET | 查询排行榜，可传 `levelId` 和 `limit` |
| `/api/events/batch` | POST | 批量上报玩法、广告和奖励事件 |

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

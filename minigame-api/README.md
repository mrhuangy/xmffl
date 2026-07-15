# Minigame API

`minigame-api` 是给微信小游戏客户端使用的 Gin 服务，和 `admin/` 后台前端同级。它只承载小游戏运行时接口：登录、远程配置、玩家进度、体力消耗、关卡结算、商城兑换、排行榜和事件上报。

## 目录

```text
cmd/server              服务入口
internal/config         环境变量配置
internal/database       MySQL 连接
internal/domain         请求/响应和业务模型
internal/handler        Gin Handler
internal/middleware     鉴权中间件
internal/repository     MySQL 读写和事务
internal/router         路由注册
internal/service        业务服务
```

## 启动

```bash
cd minigame-api
go mod tidy
go run ./cmd/server
```

默认环境变量：

| 变量 | 默认值 |
| --- | --- |
| `HTTP_ADDR` | `:8090` |
| `MYSQL_DSN` | `fpxxl:fpxxl@tcp(127.0.0.1:3306)/fpxxl?parseTime=true&loc=Local` |
| `ALLOW_ORIGIN` | `*` |
| `GIN_MODE` | `debug` |

## 接口

公开接口：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/healthz` | 健康检查 |
| `POST` | `/api/v1/auth/login` | 小游戏登录，当前为开发模式 openId 占位 |
| `GET` | `/api/v1/config/levels` | 获取关卡配置 |
| `GET` | `/api/v1/config/ads` | 获取广告频控配置 |
| `GET` | `/api/v1/leaderboard` | 查询排行榜 |

鉴权接口需要 `Authorization: Bearer <token>` 或 `X-Openid: <openId>`：

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `GET` | `/api/v1/player/progress` | 获取玩家进度 |
| `POST` | `/api/v1/levels/start` | 开始关卡并消耗 1 点体力 |
| `POST` | `/api/v1/levels/results` | 提交关卡结算，成功时发放金币并写排行榜 |
| `GET` | `/api/v1/shop/products` | 获取商城商品 |
| `POST` | `/api/v1/shop/purchase` | 购买商品，当前支持金币兑换体力 |
| `POST` | `/api/v1/events/batch` | 批量上报玩法事件 |

## 登录说明

`POST /api/v1/auth/login`

```json
{
  "code": "wx-login-code",
  "nickname": "玩家昵称",
  "avatarUrl": "https://example.com/avatar.png"
}
```

当前项目没有接入微信 `jscode2session`，服务会用 `code` 生成开发用 `openId`，并把 `openId` 作为临时 token 返回。后续接微信真实登录时，只需要替换 `PlayerService.Login` 的 openId 获取逻辑。

## 金币换体力流程

`POST /api/v1/shop/purchase`

```json
{
  "productKey": "stamina_5_by_coins"
}
```

服务端会在同一个 MySQL 事务里：

1. 创建 `purchase_orders`。
2. 扣减 `player_progress.coins`。
3. 写入负数 `coin_transactions`。
4. 增加 `player_progress.stamina`。
5. 写入正数 `stamina_transactions`。
6. 写入 `reward_grants`。

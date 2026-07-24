# Docker 本地环境

本目录用于启动本地 MySQL、运营后台 API、小游戏 API、后台页面和统一 nginx 网关。

MySQL 数据保存在 `docker_config/mysql-data/`，与 `deploy/` 同级。该目录已加入
Git 忽略，停止或删除容器不会删除其中的数据。不要手工清空此目录。

如果此前使用 Docker 命名卷 `fpxxl_mysql-data`，切换配置前必须先迁移数据；
直接重新创建容器会得到一个空数据库。推荐先使用 `mysqldump` 导出，切换挂载后再导入。

## 启动

```bash
cd docker_config
docker compose up -d --build
```

首次使用可复制环境变量示例：

```bash
cp .env.example .env
```

微信小游戏配置填在 `.env`：

```env
WECHAT_MINIGAME_APPID=
WECHAT_MINIGAME_SECRET=
```

## 本地域名

在本机 hosts 文件增加：

```text
127.0.0.1 fpxxl.local admin.fpxxl.local api.fpxxl.local minigame-api.fpxxl.local game-api.fpxxl.local assets.fpxxl.local
```

Windows hosts 文件路径通常是：

```text
C:\Windows\System32\drivers\etc\hosts
```

## 访问地址

| 地址 | 服务 |
| --- | --- |
| `http://admin.fpxxl.local` | Vue 运营后台 |
| `http://api.fpxxl.local` | 后台配置 API |
| `http://minigame-api.fpxxl.local` | 小游戏运行时 API |
| `http://game-api.fpxxl.local` | 小游戏运行时 API 别名 |
| `http://fpxxl.local` | 默认指向运营后台 |

保留直连端口，方便调试：

| 地址 | 服务 |
| --- | --- |
| `http://127.0.0.1:5173` | 运营后台容器 |
| `http://127.0.0.1:8080` | 后台配置 API |
| `http://127.0.0.1:8090` | 小游戏运行时 API |
| `127.0.0.1:3306` | MySQL |

## 常用命令

```bash
docker compose config
docker compose ps
docker compose logs -f nginx
docker compose down


docker compose -f docker-compose.prod.yml  up
docker compose -f docker-compose.prod.yml  down
```

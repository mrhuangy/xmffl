# 仓库指南

## 项目结构

- `wechat_minigame/`：微信小游戏原生实现。
- `cocos_project/`：Cocos Creator 工程。
- `backend/`：Go 后端 API。
- `admin/`：Vue 运营后台。
- `deploy/mysql/`：MySQL 建表与种子数据。
- `docs/`：项目设计、架构和环境说明。

## 构建、测试与开发命令

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
npm run build
```

完整本地环境：

```bash
docker compose up --build
```

## 编码与测试约定

Go 代码提交前运行 `gofmt`。Vue 后台使用 Vite 默认构建流程，构建产物 `dist/` 不提交。

新增业务逻辑、配置解析、奖励发放、事件上报或数据同步行为时，应补充对应测试。变更数据库结构时，在 `deploy/mysql/` 增加新的 SQL 脚本，并同步更新 `docs/backend-admin-setup.md`。

## Agent 专用说明

编辑前先检查仓库当前状态，避免覆盖无关的用户改动。变更范围应聚焦于当前任务。新增工具链后同步更新本文件，写明贡献者需要运行的准确命令。

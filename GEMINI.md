# GEMINI.md

本文件为 Gemini 在本仓库工作时的项目指南。更完整的协作约束见 `AGENTS.md`，业务规划见 `docs/`。

## 项目概述

这是一个基于现有 Gin + GORM + Viper + Zap 骨架二次开发的二手书扫码回收判断系统。

项目保留双入口：

- `main.go`：本地 HTTP 服务入口。
- `api/client.go`：Vercel Serverless 入口。

项目保留两类路由：

- client API：`routes/client/route.go`
- web 页面：`routes/web/web.go`

不要删除 web 相关代码。

## 当前业务约定

V1 client API 挂在 `/app/v1` 下：

- `GET /app/v1/demo/ping.json`
- `GET /app/v1/book/check.json`
- `GET /app/v1/book/detail.json`
- `GET /app/v1/scan/history.json`

路由、表名、模型名不额外增加模块名前缀。

V1 核心表：

- `book`
- `price_snapshot`
- `scan_log`

默认物理表名受 `DB_PREFIX=t_` 影响：

- `t_book`
- `t_price_snapshot`
- `t_scan_log`

数据库使用 MySQL 8.4。

## 开发方式

本地开发使用 `application.yaml` 时，需要环境变量：

```text
DEV=1
TLS=false
```

建议本地端口不要使用 80，可通过启动参数覆盖：

```bash
go run main.go -APP_PORT=8888
```

验证：

```bash
GOCACHE=.tmp/go-build-cache go test ./...
```

模型生成使用：

```bash
go run bin/gormt.go
```

`config.yml` 应只指向当前 V1 三张业务表。

## 代码约定

V1 新增业务代码按当前仓库骨架组织：

```text
routes/client
  -> app/controllers/client/v1
  -> app/requests
  -> app/response
  -> app/models
```

controller 负责参数解析、V1 用例编排和响应返回；V1 只扩展当前已有业务目录。

Go 文件名统一使用驼峰命名，不使用下划线，例如 `BookController.go`。

API 响应沿用 `pkg/echo.Success` 和 `pkg/echo.Error`，响应字段为 `code`、`msg`、`data`、`reqId`。

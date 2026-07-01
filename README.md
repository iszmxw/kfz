# 二手书扫码回收判断系统

这是一个基于现有 Gin + GORM + Viper + Zap 骨架二次开发的 Go 后端项目，用于支持二手书扫码后的回收判断、价格快照和扫码历史。

更完整的协作约束见 `AGENT.md`，业务规划见 `docs/`。

## 技术栈

- Web 框架：Gin
- ORM：GORM v2
- 数据库：MySQL 8.4
- 缓存：Redis
- 配置：Viper，支持 `application.yaml` 和 Vercel `CONFIG`
- 日志：Zap
- 部署入口：`main.go` 本地运行，`api/client.go` Vercel Serverless

## 本地运行

本地使用 `application.yaml` 时需要开启开发模式：

```bash
DEV=1 TLS=false go run main.go -APP_PORT=8888
```

启动后可以访问：

```text
http://127.0.0.1:8888/route
http://127.0.0.1:8888/app/v1/demo/ping.json
```

GoLand 建议配置：

```text
Program arguments:
-APP_PORT=8888

Environment variables:
DEV=1;TLS=false
```

## 当前 API

V1 API 挂在 `/app/v1` 下：

- `GET /app/v1/demo/ping.json`
- `GET /app/v1/book/check.json`
- `GET /app/v1/book/detail.json`
- `GET /app/v1/scan/history.json`

响应结构沿用 `pkg/echo`：

```json
{
  "code": 1,
  "data": {},
  "msg": "success.",
  "reqId": "request-id"
}
```

## 数据库

默认数据库名：

```text
kfz
```

V1 逻辑表：

- `book`
- `price_snapshot`
- `scan_log`

默认 `DB_PREFIX=t_`，因此物理表名为：

- `t_book`
- `t_price_snapshot`
- `t_scan_log`

DDL 设计见 `docs/database.md`。

## 模型生成

项目使用 `bin/gormt.go` 和 `config.yml` 生成 `app/models`。

当前 `config.yml` 只应指向 V1 三张业务表：

```yaml
table_prefix: "-t_"
table_names: "t_book,t_price_snapshot,t_scan_log"
```

生成命令：

```bash
go run bin/gormt.go
```

如果生成结果不理想，V1 可以手写 `Book`、`PriceSnapshot`、`ScanLog` 三个模型，但需要保持 `TableName()` 和 column 常量风格一致。

## 测试

```bash
GOCACHE=.tmp/go-build-cache go test ./...
```

## 目录重点

- `docs/`：PRD、API、数据库、数据模型、后端架构文档
- `routes/client/route.go`：client API 路由
- `routes/web/web.go`：web 页面路由，保留不要删除
- `app/controllers/client/v1`：client 控制器
- `app/controllers/web/v1`：web 控制器
- `app/models`：GORM 模型
- `pkg/echo`：统一响应封装

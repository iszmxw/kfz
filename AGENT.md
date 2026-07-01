# AGENT.md

本文件用于约束后续 AI Agent 或开发者在本仓库中的协作方式。修改代码前应先阅读本文档和 `docs/` 下的规划文档。

## 项目定位

这是一个基于现有 Gin/GORM 骨架二次开发的孔夫子二手书扫码回收判断系统。

当前项目不是从零搭建的新框架，后续开发应复用现有骨架：

- 本地入口：`main.go`
- Vercel Serverless 入口：`api/client.go`
- 路由注册：`bootstrap/route.go`
- client API 路由：`routes/client/route.go`
- web 页面路由：`routes/web/web.go`
- client 控制器：`app/controllers/client/v1`
- web 控制器：`app/controllers/web/v1`
- GORM 模型：`app/models`
- 配置：`application.yaml` 或 Vercel `CONFIG`
- 响应封装：`pkg/echo`

## 当前业务边界

本仓库后续只做二手书扫码回收判断系统，不再按多模块平台设计。

命名约定：

- 路由直接使用业务资源名，例如 `/book`、`/scan`。
- 表逻辑名直接使用业务实体名，例如 `book`、`price_snapshot`、`scan_log`。
- 数据库名以实际部署配置为准，当前本地示例使用 `kfz`。
- 文档文件名保持小写。

核心 V1 接口规划：

- `GET /app/v1/book/check.json`
- `GET /app/v1/book/detail.json`
- `GET /app/v1/scan/history.json`

当前已保留一个 client demo：

- `GET /app/v1/demo/ping.json`

web 相关代码需要保留，不要删除：

- `app/controllers/web/v1`
- `routes/web/web.go`
- `templates`

## 架构约定

V1 新增业务代码优先贴合当前仓库骨架：

```text
routes/client
  -> app/controllers/client/v1
  -> app/requests
  -> app/response
  -> app/models
```

各层职责：

- route：只注册 URL 和控制器方法。
- controller：解析请求、编排 V1 用例、返回 `pkg/echo` 响应。
- request：请求参数结构和校验。
- response：API 响应 DTO。
- model：GORM 数据模型，可以由 gormt 生成，也可以 V1 手写。

V1 只扩展以上现有目录；当 controller 内业务编排明显变复杂时，再基于真实重复代码抽取公共能力。

Go 文件名统一使用驼峰命名，不使用下划线，例如 `BookController.go`、`PriceSnapshot.go`、`ScanLog.go`。文档文件名继续保持小写。

## 数据库约定

数据库使用 MySQL 8.4。

V1 核心逻辑表：

- `book`
- `price_snapshot`
- `scan_log`

当前项目通过 `DB_PREFIX` 控制物理表前缀。默认配置为：

```yaml
DB_PREFIX: t_
```

因此默认物理表名为：

- `t_book`
- `t_price_snapshot`
- `t_scan_log`

金额字段在数据库中统一使用整数分，例如：

- `avg_price_cent`
- `min_price_cent`
- `max_price_cent`

API 层可以转换成人民币元展示。

## 模型生成约定

项目使用 `bin/gormt.go` 和根目录 `config.yml` 生成 `app/models`。

如果使用 gormt，建议只生成当前业务表：

```yaml
table_names: "t_book,t_price_snapshot,t_scan_log"
table_prefix: "-t_"
```

如果 gormt 生成结果不理想，V1 可以手写模型，但需要保持：

- `TableName()` 明确返回真实物理表名或符合项目表前缀机制。
- 字段 tag 与数据库 column 一致。
- column 常量风格与现有 `app/models` 保持一致。

## API 响应约定

V1 沿用当前项目响应风格，不引入新的响应框架。

成功响应使用 `pkg/echo.Success`，结构为：

```json
{
  "code": 1,
  "data": {},
  "msg": "success.",
  "reqId": "request-id"
}
```

失败响应使用 `pkg/echo.Error`，结构为：

```json
{
  "code": 0,
  "msg": "error message",
  "reqId": "request-id"
}
```

当前 HTTP 状态码通常保持 200，业务状态通过 `code` 表达。

## 扫码业务规则

ISBN 表示书籍品种，不表示具体某一本实体书。

同一本书或同 ISBN 重复扫码时：

- `book` 不重复创建。
- `scan_log` 每次扫码新增一条行为记录。
- `price_snapshot` 保存价格快照，不因扫码行为直接覆盖历史记录。
- `client_request_id` 用于处理客户端重试幂等。
- `batch_id` 用于识别同一批次内重复扫码。

当没有有效价格数据时，判断结果应倾向于 `NEED_REVIEW`，而不是直接 `ACCEPT`。

## 本地运行

本地使用 `application.yaml` 时，需要设置：

```text
DEV=1
TLS=false
```

建议开发端口使用 `8888` 或 `8080`，不要使用 80。

GoLand Run Configuration 建议：

```text
Program arguments:
-APP_PORT=8888

Environment variables:
DEV=1;TLS=false
```

启动后可访问：

```text
http://127.0.0.1:8888/route
http://127.0.0.1:8888/app/v1/demo/ping.json
```

## 常用验证命令

测试：

```bash
GOCACHE=.tmp/go-build-cache go test ./...
```

文档一致性检查：

```bash
rg -n "/api/v1" docs
rg -n "SQLite|sqlite" docs
rg -n "二手书业务前缀|模块前缀" docs
```

命名检查：

```bash
rg -n "/book|/scan|book|price_snapshot|scan_log" docs
```

路由检查：

```bash
go run main.go -APP_PORT=8888
```

然后访问 `/route` 查看已注册路由。

## 修改代码注意事项

- 不要删除 web 相关文件。
- 不要引入另一套后端框架。
- 不要把 V1 接口挂回 `/api/v1`。
- 不要给二手书业务额外增加模块名前缀；本项目只有这一个业务域。
- 不要直接修改用户未要求变更的旧逻辑。
- 若工作区已有未提交修改，先确认差异来源，不要随意回滚。
- 文档和代码应与 `docs/backend_design.md`、`docs/api.md`、`docs/database.md`、`docs/data_model.md` 保持一致。

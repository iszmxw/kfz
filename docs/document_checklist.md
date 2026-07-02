# 文档检查清单

本文档用于在修改 PRD、API、数据库、数据模型、后端架构或代码实现前后，快速检查项目规划是否仍然一致。

适用范围：

- `AGENTS.md`
- `README.md`
- `CLAUDE.md`
- `GEMINI.md`
- `docs/prd.md`
- `docs/backend_design.md`
- `docs/api.md`
- `docs/database.md`
- `docs/data_model.md`

## 1. 使用方式

每次涉及接口、表结构、路由、模型、配置或业务规则变更时，按以下顺序检查：

1. 先读 `AGENTS.md`，确认项目边界和命名约定。
2. 再读 `docs/prd.md`，确认改动是否仍属于当前版本范围。
3. 修改接口时同步检查 `docs/api.md`、`docs/backend_design.md`、路由和中间件。
4. 修改表结构时同步检查 `docs/database.md`、`docs/data_model.md`、`config.yml` 和 `app/models`。
5. 修改启动、配置或部署逻辑时同步检查 `README.md`、`CLAUDE.md`、`GEMINI.md` 和对应代码入口。

## 2. 当前待修正项

本节记录当前文档与代码、文档与文档之间已经发现的不一致，后续处理完可从本节移除。

暂无明确待修正项。

## 3. 项目边界检查

- [ ] 项目定位仍是二手书扫码回收判断系统。
- [ ] 没有重新引入多模块平台设计。
- [ ] 新增路由直接使用业务资源名，例如 `/book`、`/scan`。
- [ ] 未将 V1 接口挂回 `/api/v1`。
- [ ] 未给二手书业务增加额外模块名前缀。
- [ ] web 相关代码和文档仍保留，包括 `routes/web/web.go`、`app/controllers/web/v1` 和 `templates`。

## 4. 版本范围检查

- [ ] V1 仍聚焦最小扫码判断闭环。
- [ ] V1 包含扫码 ISBN、后端判断、书籍信息、判断结果、建议回收价、判断原因和扫码日志。
- [ ] V1 不提前承诺完整库存、销售、财务、绩效、重运营后台。
- [ ] 连续扫码、批次、人工确认、门店账号、库存入库和标签打印仍作为 V2/V3 能力处理。
- [ ] 若 V2/V3 能力提前实现，相关文档已明确版本范围变化。

## 5. API 检查

- [ ] V1 API 统一挂在 `/app/v1` 下。
- [ ] 核心接口保持一致：
  - `GET /app/v1/book/check.json`
  - `GET /app/v1/book/detail.json`
  - `GET /app/v1/scan/history.json`
- [ ] 当前 demo 接口 `GET /app/v1/demo/ping.json` 保留。
- [ ] `docs/api.md` 的请求参数与 `app/requests` 中结构一致。
- [ ] `docs/api.md` 的响应 DTO 与 `app/response` 中结构一致。
- [ ] 响应结构仍沿用 `pkg/echo.Success` 和 `pkg/echo.Error`。
- [ ] 成功响应仍为 `code=1`、`msg=success.`、`data`、`reqId`。
- [ ] 失败响应仍为 `code=0`、`msg`、`reqId`。
- [ ] HTTP 状态码策略如有变化，`docs/api.md` 和响应封装说明已同步。
- [ ] API 金额返回人民币元，数据库使用 `decimal(20,2)` 保存人民币元。
- [ ] ISBN 入参允许 ISBN-10 或 ISBN-13，并返回 `normalized_isbn`。

## 6. 路由与中间件检查

- [ ] `routes/client/route.go` 中已注册文档声明的 client API。
- [ ] `app/controllers/client/v1/BaseController.go` 的 `Group` 包含已注册控制器。
- [ ] `BookController`、`ScanController` 方法名与路由文档一致。
- [ ] `app/middlewares/v1/client.go` 白名单与当前 V1 鉴权策略一致。
- [ ] 如果某接口需要登录，文档中已说明 token 或后续门店登录策略。
- [ ] `bootstrap/route.go` 仍同时注册 client 路由和 web 路由。
- [ ] `/route` 仍可用于本地查看已注册路由。

## 7. 数据库检查

- [ ] 数据库版本仍按 MySQL 8.4 规划。
- [ ] V1 第一轮建表仍只包含三张核心逻辑表：
  - `book`
  - `price_snapshot`
  - `scan_log`
- [ ] 默认 `DB_PREFIX=t_` 时，物理表名为：
  - `t_book`
  - `t_price_snapshot`
  - `t_scan_log`
- [ ] `docs/database.md` DDL 与 `docs/data_model.md` 字段描述一致。
- [ ] 金额字段在数据库中统一使用十进制定点金额口径，MySQL 类型为 `decimal(20,2)`，Go 业务层使用 `shopspring/decimal.Decimal`，字段名不使用 `_cent` 后缀。
- [ ] 时间字段仍使用 `datetime(3)`。
- [ ] `book.isbn` 表示书籍品种，不表示具体某一本实体书。
- [ ] `price_snapshot` 保存价格快照，不因扫码行为覆盖历史。
- [ ] `scan_log` 每次扫码新增一条行为记录。
- [ ] `client_request_id` 唯一约束仍符合幂等设计。
- [ ] 外键、索引和字段可空性与模型实现一致。

## 8. 模型生成检查

- [ ] `config.yml` 的 `table_names` 只包含当前业务表。
- [ ] `config.yml` 的 `table_prefix` 与实际 `DB_PREFIX` 策略一致。
- [ ] `app/models` 中模型字段 tag 与数据库 column 一致。
- [ ] 如果模型由 gormt 生成，生成后未覆盖用户手写业务逻辑。
- [ ] 如果模型手写，`TableName()` 明确返回真实物理表名或符合项目表前缀机制。
- [ ] column 常量风格与现有 `app/models` 保持一致。

## 9. 业务规则检查

- [ ] 同 ISBN 重复扫码时，`book` 不重复创建。
- [ ] 同 ISBN 重复扫码时，`scan_log` 每次新增。
- [ ] 扫码行为不直接覆盖历史价格快照。
- [ ] 没有有效价格数据时，判断结果倾向于 `NEED_REVIEW`。
- [ ] `batch_id` 只用于识别批量扫码场景和批次内重复。
- [ ] `duplicate_recently` 与 `duplicate_in_batch` 语义清晰且 API 文档一致。
- [ ] 历史记录保存当时判断结果，不因后续规则或价格变化被改写。

## 10. 配置与部署检查

- [ ] 本地模式仍通过 `DEV=1` 使用 `application.yaml`。
- [ ] 本地开发示例端口保持 `8888` 或 `8080`。
- [ ] Vercel 模式仍通过环境变量 `CONFIG` 传入 JSON 配置。
- [ ] 配置读取优先级仍为 `APPENV_` 环境变量高于配置文件。
- [ ] `main.go` 和 `api/client.go` 初始化流程保持一致。
- [ ] Redis、MySQL、模板、路由和 pprof 初始化在两个入口中均有覆盖。
- [ ] `vercel.json` 路由仍指向 `api/client.go`。

## 11. 文档同步检查

- [ ] 改业务目标时同步 `docs/prd.md`。
- [ ] 改接口契约时同步 `docs/api.md`。
- [ ] 改表结构或索引时同步 `docs/database.md`。
- [ ] 改领域对象或版本规划时同步 `docs/data_model.md`。
- [ ] 改分层、目录、启动链路或中间件策略时同步 `docs/backend_design.md`。
- [ ] 改运行方式、端口、配置示例时同步 `README.md`。
- [ ] 改 AI Agent 协作约束时同步 `AGENTS.md`。
- [ ] 如 `CLAUDE.md`、`GEMINI.md` 保留为协作入口，关键约束已同步。

## 12. 常用一致性命令

```bash
rg -n "/api/v1" docs AGENTS.md README.md CLAUDE.md GEMINI.md
rg -n "SQLite|sqlite" docs AGENTS.md README.md CLAUDE.md GEMINI.md
rg -n "二手书业务前缀|模块前缀" docs AGENTS.md README.md CLAUDE.md GEMINI.md
rg -n "_cent|整数分|单位分|decimal\\(10,2\\)" docs AGENTS.md README.md CLAUDE.md GEMINI.md app --glob '!docs/document_checklist.md'
rg -n "localhost:80|APP_PORT=80|:80" docs AGENTS.md README.md CLAUDE.md GEMINI.md --glob '!docs/document_checklist.md'
rg -n "/app/v1/book/check.json|/app/v1/book/detail.json|/app/v1/scan/history.json" docs AGENTS.md README.md CLAUDE.md GEMINI.md routes app
rg -n "t_book|t_price_snapshot|t_scan_log|DB_PREFIX" docs AGENTS.md README.md CLAUDE.md GEMINI.md config.yml application.yaml
```

本地运行路由检查：

```bash
DEV=1 TLS=false go run main.go -APP_PORT=8888
```

启动后访问：

```text
http://127.0.0.1:8888/route
```

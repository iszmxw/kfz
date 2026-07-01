# 二手书扫码回收判断系统 API 设计

## 1. 设计目标

本文档定义微信小程序与当前 Gin 后端之间的接口契约。接口挂载在现有 client 路由 `/app/v1` 下，二手书业务统一使用 `/app/v1` 前缀。

设计原则：

- 复用现有 Gin 路由、client 中间件和 `pkg/echo` 响应封装。
- V1 优先跑通扫码判断闭环，不引入新的响应框架。
- 历史记录保留当时判断快照，后续规则变化不改写历史。
- 同 ISBN 重复扫码时，书籍数据复用，扫码日志新增。
- 批量扫码重复 ISBN 时，由后端返回重复标记，小程序负责让用户确认。

## 2. 基础约定

### 2.1 Base URL

开发环境示例：

```text
http://localhost:80
```

接口前缀：

```text
/app/v1
```

### 2.2 数据格式

- 请求体：`application/json`
- 响应体：`application/json`
- 时间格式：ISO 8601，例如 `2026-06-30T10:00:00+08:00`
- 金额单位：API 返回人民币元，数据库保存人民币分
- ISBN：接口入参允许 ISBN-10 或 ISBN-13，后端返回 `normalized_isbn`

### 2.3 统一响应结构

V1 先沿用当前项目的 `pkg/echo.Success` / `pkg/echo.Error`。

成功响应：

```json
{
  "code": 1,
  "data": {},
  "msg": "success.",
  "reqId": "0b8c0a8e-8d24-49d4-bdf7-52f80dd0d91a"
}
```

失败响应：

```json
{
  "code": 0,
  "msg": "未识别到有效 ISBN",
  "reqId": "0b8c0a8e-8d24-49d4-bdf7-52f80dd0d91a"
}
```

说明：

- HTTP 状态码当前统一为 200，业务状态由 `code` 表达，符合现有 `pkg/echo` 行为。
- `reqId` 来自 `Tracking-Id` 中间件。
- V1 错误先复用 `echo.Error(c, "Failed", msg)`；后续可补充二手书专用错误码。

## 3. 枚举

### 3.1 decision

| 值 | 说明 |
| --- | --- |
| `ACCEPT` | 建议回收 |
| `REJECT` | 不建议回收 |
| `NEED_REVIEW` | 需要人工确认 |

### 3.2 confidence

| 值 | 说明 |
| --- | --- |
| `HIGH` | 样本数充足，价格稳定 |
| `MEDIUM` | 有价格数据，但样本一般 |
| `LOW` | 样本少、价格异常或数据偏旧 |
| `NONE` | 无有效价格数据 |

### 3.3 batch_status

| 值 | 说明 |
| --- | --- |
| `DRAFT` | 草稿批次 |
| `SUBMITTED` | 已提交 |
| `CANCELLED` | 已取消 |

## 4. 通用对象

### 4.1 book

```json
{
  "isbn": "9787111128069",
  "normalized_isbn": "9787111128069",
  "title": "示例书名",
  "author": "示例作者",
  "publisher": "示例出版社",
  "publish_year": "2018",
  "cover_url": ""
}
```

### 4.2 market_price

```json
{
  "source": "local_mock",
  "min": 18.0,
  "avg": 24.5,
  "max": 39.0,
  "sample_count": 8,
  "confidence": "HIGH",
  "collected_at": "2026-06-30T10:00:00+08:00"
}
```

### 4.3 duplicate_info

```json
{
  "scanned_recently": true,
  "duplicate_in_current_batch": false,
  "last_scanned_at": "2026-06-30T10:20:00+08:00"
}
```

字段说明：

| 字段 | 说明 |
| --- | --- |
| `scanned_recently` | 近期是否扫过同 ISBN |
| `duplicate_in_current_batch` | 当前批次内是否已存在同 ISBN |
| `last_scanned_at` | 最近一次扫码时间，无记录时为空 |

## 5. 扫码判断

### 5.1 GET /app/v1/book/check.json

用途：小程序扫码后立即判断是否回收。

请求参数：

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| `isbn` | query | 是 | 扫描得到的 ISBN |
| `batch_id` | query | 否 | 当前批次 ID，批量扫码时传 |
| `client_request_id` | query | 否 | 小程序生成的请求幂等 ID |

请求示例：

```text
GET /app/v1/book/check.json?isbn=9787111128069&batch_id=batch_001
```

成功响应：

```json
{
  "code": 1,
  "msg": "success.",
  "reqId": "0b8c0a8e-8d24-49d4-bdf7-52f80dd0d91a",
  "data": {
    "scan_log_id": "scan_001",
    "book": {
      "isbn": "9787111128069",
      "normalized_isbn": "9787111128069",
      "title": "示例书名",
      "author": "示例作者",
      "publisher": "示例出版社",
      "publish_year": "2018",
      "cover_url": ""
    },
    "decision": "ACCEPT",
    "reason": "二手市场均价满足回收规则",
    "suggested_recycle_price": 7.4,
    "market_price": {
      "source": "local_mock",
      "min": 18.0,
      "avg": 24.5,
      "max": 39.0,
      "sample_count": 8,
      "confidence": "HIGH",
      "collected_at": "2026-06-30T10:00:00+08:00"
    },
    "duplicate": {
      "scanned_recently": false,
      "duplicate_in_current_batch": false,
      "last_scanned_at": null
    },
    "scanned_at": "2026-06-30T10:30:00+08:00"
  }
}
```

重复扫码响应示例：

```json
{
  "code": 1,
  "msg": "success.",
  "reqId": "0b8c0a8e-8d24-49d4-bdf7-52f80dd0d91a",
  "data": {
    "scan_log_id": "scan_002",
    "book": {
      "isbn": "9787111128069",
      "normalized_isbn": "9787111128069",
      "title": "示例书名",
      "author": "示例作者",
      "publisher": "示例出版社",
      "publish_year": "2018",
      "cover_url": ""
    },
    "decision": "ACCEPT",
    "reason": "二手市场均价满足回收规则",
    "suggested_recycle_price": 7.4,
    "market_price": {
      "source": "local_mock",
      "min": 18.0,
      "avg": 24.5,
      "max": 39.0,
      "sample_count": 8,
      "confidence": "HIGH",
      "collected_at": "2026-06-30T10:00:00+08:00"
    },
    "duplicate": {
      "scanned_recently": true,
      "duplicate_in_current_batch": true,
      "last_scanned_at": "2026-06-30T10:29:30+08:00"
    },
    "batch_action_required": "CONFIRM_DUPLICATE",
    "scanned_at": "2026-06-30T10:30:00+08:00"
  }
}
```

前端处理：

- `decision = ACCEPT`：显示“收”，播放 `shou.mp3`。
- `decision = REJECT`：显示“不收”，播放 `bushou.mp3`。
- `decision = NEED_REVIEW`：显示“需人工确认”。
- `duplicate.duplicate_in_current_batch = true`：提示“本批次已扫过这本书”，让用户确认是否还有一本相同书。

## 6. 书籍详情

### 6.1 GET /app/v1/book/detail.json

用途：查看书籍完整信息和价格来源。

请求参数：

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| `isbn` | query | 是 | ISBN |

请求示例：

```text
GET /app/v1/book/detail.json?isbn=9787111128069
```

成功响应：

```json
{
  "code": 1,
  "msg": "success.",
  "reqId": "0b8c0a8e-8d24-49d4-bdf7-52f80dd0d91a",
  "data": {
    "book": {
      "isbn": "9787111128069",
      "normalized_isbn": "9787111128069",
      "title": "示例书名",
      "author": "示例作者",
      "publisher": "示例出版社",
      "publish_year": "2018",
      "cover_url": ""
    },
    "latest_market_price": {
      "source": "local_mock",
      "min": 18.0,
      "avg": 24.5,
      "max": 39.0,
      "sample_count": 8,
      "confidence": "HIGH",
      "collected_at": "2026-06-30T10:00:00+08:00"
    },
    "latest_decision": {
      "decision": "ACCEPT",
      "reason": "二手市场均价满足回收规则",
      "suggested_recycle_price": 7.4,
      "updated_at": "2026-06-30T10:30:00+08:00"
    }
  }
}
```

## 7. 扫码历史

### 7.1 GET /app/v1/scan/history.json

用途：查询扫码历史记录。

请求参数：

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| `page` | query | 否 | 页码，默认 1 |
| `page_size` | query | 否 | 每页数量，默认 20 |
| `decision` | query | 否 | `ACCEPT`、`REJECT`、`NEED_REVIEW` |
| `isbn` | query | 否 | 按 ISBN 查询 |
| `batch_id` | query | 否 | 按批次查询 |

请求示例：

```text
GET /app/v1/scan/history.json?page=1&page_size=20&decision=ACCEPT
```

成功响应：

```json
{
  "code": 1,
  "msg": "success.",
  "reqId": "0b8c0a8e-8d24-49d4-bdf7-52f80dd0d91a",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": [
      {
        "scan_log_id": "scan_001",
        "isbn": "9787111128069",
        "title": "示例书名",
        "decision": "ACCEPT",
        "reason": "二手市场均价满足回收规则",
        "suggested_recycle_price": 7.4,
        "confidence": "HIGH",
        "scanned_at": "2026-06-30T10:30:00+08:00"
      }
    ]
  }
}
```

## 8. 批量扫码 V2 预留

### 8.1 POST /app/v1/scan/batch/create.json

用途：创建批量扫码批次。

### 8.2 POST /app/v1/scan/batch/confirmDuplicate.json

用途：用户确认当前批次中确实还有一本相同 ISBN 后，增加批次明细数量。

请求体：

```json
{
  "batch_id": "batch_001",
  "isbn": "9787111128069",
  "scan_log_id": "scan_002"
}
```

### 8.3 POST /app/v1/scan/batch/submit.json

用途：提交连续扫码批次。

### 8.4 POST /app/v1/book/manualDecision.json

用途：当系统返回 `NEED_REVIEW` 时，由操作员人工确认收或不收。

### 8.5 GET /app/v1/stat/today.json

用途：查询今日扫码统计。

## 9. V1 必做接口

V1 只需要实现：

- `GET /app/v1/book/check.json`
- `GET /app/v1/book/detail.json`
- `GET /app/v1/scan/history.json`

V1 可以暂缓：

- `POST /app/v1/scan/batch/create.json`
- `POST /app/v1/scan/batch/confirmDuplicate.json`
- `POST /app/v1/scan/batch/submit.json`
- `POST /app/v1/book/manualDecision.json`
- `GET /app/v1/stat/today.json`

## 10. 待确认问题

- V1 是否就启用 `client_request_id` 幂等控制。
- V1 二手书接口是否全部加入 token 白名单。
- 无书籍数据时，接口返回 `REJECT` 还是 `NEED_REVIEW`。
- 无价格数据时，接口返回 `REJECT` 还是 `NEED_REVIEW`。
- 小程序是否需要离线缓存最近 N 条扫码结果。

# 小程序扫码人识别方案

## 1. 背景

当前小程序没有登录功能。用户扫码后，后端会写入 `scan_log`，后台也有扫码日志模块，但目前日志无法可靠展示“是谁扫的”。

现状结论：

- `scan_log` 已预留 `operator_id` 字段，但扫码接口没有写入该字段。
- 小程序扫码请求只传 `isbn` 和 `client_request_id`。
- 后台扫码日志列表只展示 ISBN、判断结果、原因、可信度和时间。

本方案目标是使用微信小程序自带能力识别扫码用户，并在后台扫码日志中展示扫码人。

## 2. 目标

- 使用微信 `wx.login()` 建立用户身份。
- 后端通过微信 `code2Session` 获取并保存 `openid`。
- 保存用户头像和昵称，用于后台展示扫码人。
- 每次扫码写入 `scan_log.operator_id`。
- 后台扫码日志展示扫码人的头像、昵称和用户 ID。

## 3. 不做范围

- 不获取手机号。
- 不获取真实姓名、地址、性别等隐私数据。
- 不做头像上传。
- 不做对象存储转存。
- 不做门店账号体系。
- 不合并后台管理员账号和小程序扫码用户。

## 4. 核心原则

1. `openid` 是微信小程序下的用户身份识别 ID。
2. 系统内部使用 `operator_id` 关联扫码日志。
3. 头像和昵称只用于后台展示，不作为唯一身份依据。
4. 小程序使用微信自带头像昵称填写能力获取头像和昵称。
5. 扫码接口只需要携带登录态，不需要每次传头像和昵称。

## 5. 数据模型

### 5.1 新增 operator 表

新增物理表：

```text
t_operator
```

字段建议：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| id | bigint unsigned | 是 | 自增主键 |
| openid | varchar(128) | 是 | 微信小程序 openid |
| nickname | varchar(100) | 否 | 用户昵称 |
| avatar_url | varchar(1000) | 否 | 用户头像 URL |
| status | varchar(20) | 是 | `ACTIVE`、`DISABLED` |
| last_login_at | datetime(3) | 否 | 最近登录时间 |
| created_at | datetime(3) | 是 | 创建时间 |
| updated_at | datetime(3) | 是 | 更新时间 |

索引建议：

```sql
UNIQUE KEY uk_operator_openid (openid),
KEY idx_operator_status_updated_at (status, updated_at)
```

### 5.2 scan_log 关联扫码人

当前 `scan_log` 已有：

```text
operator_id varchar(36) NULL
```

本轮实现时写入当前扫码用户的 `operator.id`。

如果数据库尚未上线，建议将 `operator_id` 调整为：

```sql
operator_id bigint unsigned NULL COMMENT '扫码用户 ID'
```

并增加索引：

```sql
KEY idx_scan_log_operator_scanned_at (operator_id, scanned_at)
```

如果数据库已经上线，为降低迁移风险，也可以先保留 `varchar(36)`，写入 `operator.id` 的字符串形式。

## 6. 后端接口

### 6.1 微信登录

新增接口：

```text
POST /app/v1/operator/login.json
```

请求：

```json
{
  "code": "wx.login 返回的 code"
}
```

流程：

1. 后端使用微信 `code2Session` 换取 `openid`。
2. 根据 `openid` 查找或创建 `operator`。
3. 更新 `last_login_at`。
4. 生成系统 token。
5. 返回 token 和当前扫码用户信息。

响应：

```json
{
  "token": "client-token",
  "operator": {
    "id": 1001,
    "nickname": "",
    "avatar_url": ""
  }
}
```

### 6.2 当前用户

新增接口：

```text
GET /app/v1/operator/me.json
```

响应：

```json
{
  "id": 1001,
  "nickname": "小王",
  "avatar_url": "https://..."
}
```

### 6.3 保存头像昵称

新增接口：

```text
POST /app/v1/operator/profile.json
```

请求：

```json
{
  "nickname": "小王",
  "avatar_url": "https://..."
}
```

说明：

- 小程序通过微信自带头像昵称填写能力获取头像和昵称。
- 后端只更新当前 token 对应的 `operator`。
- 后端限制昵称长度和头像 URL 长度。
- 不新增头像上传接口。

## 7. 小程序改造

### 7.1 启动登录

小程序启动或进入首页时：

1. 读取本地 token。
2. 有 token 时请求 `/operator/me.json` 校验。
3. token 无效或不存在时调用 `wx.login()`。
4. 将 `code` 发给 `/operator/login.json`。
5. 保存后端返回的 token 和 operator 信息。

### 7.2 头像昵称

在设置页或首次使用提示中使用微信自带能力：

- 头像：使用 `button open-type="chooseAvatar"`。
- 昵称：使用 `input type="nickname"`。
- 保存：调用 `/operator/profile.json`。

没有填写头像昵称时仍允许扫码，后台展示用户 ID。

### 7.3 扫码请求

在 `miniprogram/utils/request.ts` 统一携带系统 token。

建议请求头：

```text
Authorization: Bearer <token>
```

后端从 token 解析当前 `operator_id`。

## 8. 扫码日志改造

现有扫码接口：

```text
GET /app/v1/book/check.json
```

改造点：

1. 从 token 解析当前扫码用户。
2. `BookController.Check` 写入 `scan_log.operator_id`。
3. `client_request_id` 幂等逻辑保持不变。
4. 历史未关联用户的扫码记录在后台显示为“未识别”。

## 9. 后台展示

后台 `扫码日志` 页面增加扫码人信息：

- 头像。
- 昵称。
- 用户 ID。

`GET /admin/api/v1/scan-log/list` 建议返回关联后的展示 DTO，不再直接返回原始 `models.ScanLog`。

响应项示例：

```json
{
  "id": 1,
  "normalized_isbn": "9787111128069",
  "decision": "ACCEPT",
  "operator": {
    "id": 1001,
    "nickname": "小王",
    "avatar_url": "https://..."
  },
  "scanned_at": "2026-07-24T10:00:00+08:00"
}
```

## 10. 配置

新增微信小程序配置：

```yaml
wechat:
  miniprogram_app_id: ""
  miniprogram_app_secret: ""
  token_ttl_seconds: 2592000
```

Vercel 部署时通过 `CONFIG` 或 `APPENV_` 环境变量注入。

## 11. 测试计划

后端测试：

- 首次微信登录会创建 `operator`。
- 同一个 `openid` 重复登录不会重复创建用户。
- 保存头像昵称后，`operator.nickname` 和 `operator.avatar_url` 正确更新。
- 已登录用户扫码后，`scan_log.operator_id` 正确写入。
- 后台扫码日志能展示对应扫码人的头像和昵称。

小程序测试：

- 首次打开可以通过 `wx.login()` 获取系统 token。
- 用户可使用微信头像昵称能力完善资料。
- 扫码请求会携带 token。
- 未填写头像昵称时仍能扫码。

回归测试：

- 原有扫码判断规则不变。
- 原有价格规则不变。
- `client_request_id` 幂等逻辑不变。
- 后台管理员登录不受影响。

## 12. 最终方案结论

本轮只做微信自带能力下的轻量扫码人识别：

- 微信 `wx.login()` 获取 `openid`。
- 后端保存 `openid`、头像、昵称。
- 小程序扫码时携带系统 token。
- 后端写入 `scan_log.operator_id`。
- 后台扫码日志展示扫码人。

不接入手机号等隐私授权能力，不做头像上传，不扩展门店账号体系。

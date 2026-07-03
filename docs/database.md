# 二手书扫码回收判断系统数据库设计

## 1. 设计目标

本文档定义 V1 后端使用的 MySQL 数据库结构、索引、初始化脚本和 mock 数据策略。

V1 只落三张核心表：

- 逻辑表 `book`
- 逻辑表 `price_snapshot`
- 逻辑表 `scan_log`

当前项目通过 `DB_PREFIX` 配置 GORM 表前缀。默认按 `DB_PREFIX=t_` 规划时，实际物理表名为：

- `t_book`
- `t_price_snapshot`
- `t_scan_log`

批量扫码、人工确认、门店账号和库存相关表在 `docs/data_model.md` 中已经规划，但不进入 V1 第一轮建表。

## 2. MySQL 约定

### 2.1 版本

使用 MySQL 8.4。

原因：

- 字符集和排序规则更稳定。
- JSON、索引、时间处理能力更完整。
- 后续统计和窗口函数能力更好。

### 2.2 字符集

数据库和表统一使用：

```sql
CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci
```

如部署环境存在兼容性限制，不支持 `utf8mb4_0900_ai_ci`，可降级为：

```sql
CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
```

### 2.3 时间

- 数据库字段使用 `datetime(3)`。
- 后端统一按本地配置时区写入。
- API 返回 ISO 8601 字符串。
- V1 不强制使用数据库时区转换。

### 2.4 金额

数据库内金额统一使用十进制定点金额口径，MySQL 字段类型使用 `decimal(20,2)`；Go 业务层使用 `github.com/shopspring/decimal.Decimal`。

示例：

| 展示金额 | 数据库存储 |
| --- | --- |
| `7.40` 元 | `7.40` |
| `24.50` 元 | `24.50` |

原因：

- 避免浮点误差，并保留金额小数精度。
- 便于后续统计汇总。
- API 层直接按人民币元展示。

## 3. 数据库初始化

建议数据库名：

```sql
CREATE DATABASE IF NOT EXISTS kfz
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_0900_ai_ci;
```

当前项目不直接配置 DSN，而是通过 `application.yaml` 或 Vercel `CONFIG` 中的 `DB_*` 字段组装 GORM MySQL 连接。

本地开发配置示例：

```yaml
DB_CONNECTION: mysql
DB_HOST: 127.0.0.1
DB_PORT: 3306
DB_DATABASE: kfz
DB_USERNAME: root
DB_PASSWORD: your_password
DB_PREFIX: t_
```

## 4. V1 DDL

建议文件路径：

```text
migrations/001_init.sql
```

### 4.1 book

```sql
CREATE TABLE IF NOT EXISTS t_book (
  id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  isbn varchar(20) NOT NULL COMMENT '归一化后的 ISBN',
  title varchar(255) NOT NULL COMMENT '书名',
  author varchar(255) NULL COMMENT '作者',
  publisher varchar(255) NULL COMMENT '出版社',
  publish_year varchar(20) NULL COMMENT '出版年份',
  cover_url varchar(1000) NULL COMMENT '封面地址',
  source varchar(50) NOT NULL DEFAULT 'manual' COMMENT '数据来源：manual/mock/external',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_book_isbn (isbn),
  KEY idx_book_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='书籍基础信息';
```

设计说明：

- `id` 是数据库自增主键，便于后台管理和表关联。
- `isbn` 表示书籍品种，不表示具体一本实体书，并通过唯一索引保证不重复。
- 同 ISBN 重复扫码时复用此表记录。
- `title` 建索引用于后续后台搜索，V1 接口可以暂不使用。
- `t_` 是当前默认 `DB_PREFIX`；如果环境使用其他前缀，物理表名随之调整。

### 4.2 price_snapshot

```sql
CREATE TABLE IF NOT EXISTS t_price_snapshot (
  id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  isbn varchar(20) NOT NULL COMMENT '关联 book.isbn',
  source varchar(50) NOT NULL COMMENT '价格来源：local_mock/manual/kongfz/external',
  min_price decimal(20,2) NULL COMMENT '最低价',
  avg_price decimal(20,2) NULL COMMENT '平均价',
  max_price decimal(20,2) NULL COMMENT '最高价',
  sample_count int NOT NULL DEFAULT 0 COMMENT '有效样本数',
  confidence varchar(20) NOT NULL DEFAULT 'NONE' COMMENT 'HIGH/MEDIUM/LOW/NONE',
  raw_url varchar(1000) NULL COMMENT '来源链接',
  raw_payload_ref varchar(255) NULL COMMENT '原始数据存储引用',
  collected_at datetime(3) NOT NULL COMMENT '采集或录入时间',
  expires_at datetime(3) NULL COMMENT '过期时间',
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_price_snapshot_isbn_collected_at (isbn, collected_at),
  KEY idx_price_snapshot_isbn_expires_at (isbn, expires_at),
  CONSTRAINT fk_price_snapshot_book
    FOREIGN KEY (isbn) REFERENCES t_book (isbn)
    ON UPDATE CASCADE
    ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='二手价格快照';
```

设计说明：

- 价格快照不覆盖历史，新增记录为主。
- 判断接口读取同 ISBN 最新且未过期的快照。
- `min_price`、`avg_price`、`max_price` 允许为空，用于表达无有效价格。

### 4.3 scan_log

```sql
CREATE TABLE IF NOT EXISTS t_scan_log (
  id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  isbn varchar(20) NOT NULL COMMENT '原始或归一化 ISBN',
  normalized_isbn varchar(20) NOT NULL COMMENT '归一化后的 ISBN',
  batch_id varchar(36) NULL COMMENT '批次 ID，V2 使用',
  store_id varchar(36) NULL COMMENT '门店 ID，V2 使用',
  operator_id varchar(36) NULL COMMENT '操作员 ID，V2 使用',
  decision varchar(20) NOT NULL COMMENT 'ACCEPT/REJECT/NEED_REVIEW',
  reason varchar(500) NOT NULL COMMENT '判断原因',
  market_min_price decimal(20,2) NULL COMMENT '判断时最低价',
  market_avg_price decimal(20,2) NULL COMMENT '判断时平均价',
  market_max_price decimal(20,2) NULL COMMENT '判断时最高价',
  market_sample_count int NULL COMMENT '判断时样本数',
  suggested_recycle_price decimal(20,2) NULL COMMENT '建议回收价',
  confidence varchar(20) NOT NULL DEFAULT 'NONE' COMMENT '判断时数据可信度',
  price_snapshot_id bigint unsigned NULL COMMENT '使用的价格快照 ID',
  duplicate_recently tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否近期扫过',
  duplicate_in_batch tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否当前批次重复',
  client_request_id varchar(64) NULL COMMENT '小程序请求幂等 ID',
  scanned_at datetime(3) NOT NULL COMMENT '扫码时间',
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_scan_log_isbn_scanned_at (normalized_isbn, scanned_at),
  KEY idx_scan_log_batch_created_at (batch_id, created_at),
  KEY idx_scan_log_decision_scanned_at (decision, scanned_at),
  UNIQUE KEY uk_scan_log_client_request_id (client_request_id),
  CONSTRAINT fk_scan_log_book
    FOREIGN KEY (normalized_isbn) REFERENCES t_book (isbn)
    ON UPDATE CASCADE
    ON DELETE RESTRICT,
  CONSTRAINT fk_scan_log_price_snapshot
    FOREIGN KEY (price_snapshot_id) REFERENCES t_price_snapshot (id)
    ON UPDATE CASCADE
    ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='扫码日志';
```

设计说明：

- 每次扫码新增一条 `scan_log`。
- 历史判断结果冗余保存价格和原因，避免后续价格变化影响历史。
- `client_request_id` 用于防止网络重试重复写入。
- MySQL unique key 允许多个 `NULL`，所以不传 `client_request_id` 时不会互相冲突。

## 5. V1 Seed 数据

建议文件路径：

```text
seeds/v1_mock_data.sql
```

### 5.1 示例数据

```sql
INSERT INTO t_book (
  isbn, title, author, publisher, publish_year, cover_url, source, created_at, updated_at
) VALUES
  ('9787111128069', '示例可回收图书', '示例作者 A', '机械工业出版社', '2018', '', 'mock', NOW(3), NOW(3)),
  ('9787115428028', '示例低价图书', '示例作者 B', '人民邮电出版社', '2016', '', 'mock', NOW(3), NOW(3)),
  ('9787300000001', '示例样本不足图书', '示例作者 C', '示例出版社', '2020', '', 'mock', NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  author = VALUES(author),
  publisher = VALUES(publisher),
  publish_year = VALUES(publish_year),
  cover_url = VALUES(cover_url),
  source = VALUES(source),
  updated_at = NOW(3);

INSERT INTO t_price_snapshot (
  isbn, source, min_price, avg_price, max_price,
  sample_count, confidence, raw_url, raw_payload_ref, collected_at, expires_at, created_at
) VALUES
  ('9787111128069', 'local_mock', 18.00, 24.50, 39.00, 8, 'HIGH', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3)),
  ('9787115428028', 'local_mock', 3.00, 8.00, 12.00, 6, 'MEDIUM', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3)),
  ('9787300000001', 'local_mock', 18.00, 30.00, 45.00, 1, 'LOW', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3));
```

### 5.2 Seed 覆盖场景

| ISBN | 场景 | 预期结果 |
| --- | --- | --- |
| `9787111128069` | 均价 24.50 元，样本 8 | `ACCEPT` |
| `9787115428028` | 均价 8.00 元，样本 6 | `REJECT` |
| `9787300000001` | 均价 30.00 元，样本 1 | `NEED_REVIEW` |
| 未录入 ISBN | 无书籍数据 | `REJECT` |

## 6. 查询 SQL 参考

### 6.1 查询书籍

```sql
SELECT
  isbn, title, author, publisher, publish_year, cover_url, source, created_at, updated_at
FROM t_book
WHERE isbn = ?;
```

### 6.2 查询最新有效价格快照

```sql
SELECT
  id, isbn, source, min_price, avg_price, max_price,
  sample_count, confidence, raw_url, raw_payload_ref, collected_at, expires_at, created_at
FROM t_price_snapshot
WHERE isbn = ?
  AND (expires_at IS NULL OR expires_at > NOW(3))
ORDER BY collected_at DESC
LIMIT 1;
```

### 6.3 查询近期重复扫码

```sql
SELECT
  id, scanned_at
FROM t_scan_log
WHERE normalized_isbn = ?
  AND scanned_at >= DATE_SUB(NOW(3), INTERVAL ? MINUTE)
ORDER BY scanned_at DESC
LIMIT 1;
```

注意：插入当前扫码日志前查询，避免当前请求把自己判断成重复。

### 6.4 插入扫码日志

```sql
INSERT INTO t_scan_log (
  isbn, normalized_isbn, batch_id, store_id, operator_id,
  decision, reason,
  market_min_price, market_avg_price, market_max_price,
  market_sample_count, suggested_recycle_price, confidence,
  price_snapshot_id, duplicate_recently, duplicate_in_batch,
  client_request_id, scanned_at, created_at
) VALUES (
  ?, ?, ?, ?, ?,
  ?, ?,
  ?, ?, ?,
  ?, ?, ?,
  ?, ?, ?,
  ?, ?, ?
);
```

### 6.5 扫码历史分页

```sql
SELECT
  sl.id,
  sl.normalized_isbn,
  b.title,
  sl.decision,
  sl.reason,
  sl.suggested_recycle_price,
  sl.confidence,
  sl.scanned_at
FROM t_scan_log sl
JOIN t_book b ON b.isbn = sl.normalized_isbn
WHERE (? IS NULL OR sl.decision = ?)
  AND (? IS NULL OR sl.normalized_isbn = ?)
  AND (? IS NULL OR sl.batch_id = ?)
ORDER BY sl.scanned_at DESC
LIMIT ? OFFSET ?;
```

## 7. 迁移策略

V1 可采用简单迁移策略：

1. `migrations/001_init.sql` 只创建表和索引。
2. `seeds/v1_mock_data.sql` 插入 V1 mock 数据和开发后台管理员账号。
3. 服务启动时可以检查表是否存在。
4. 开发环境允许自动执行 seed。
5. 生产环境不自动执行 seed。

开发后台 seed 账号为 `admin`，默认密码为 `admin123`。该账号仅用于本地开发和初始化验证，生产环境应改用环境配置或后台用户管理创建正式账号。

后续进入多环境部署时，再引入正式迁移工具，例如：

- `golang-migrate`
- `goose`

## 8. gormt 模型生成约定

当前项目使用 `bin/gormt.go` 和 `config.yml` 生成 `app/models`。

V1 建议在建表后将 `config.yml` 调整为：

```yaml
table_prefix: "-t_"
table_names: "t_book,t_price_snapshot,t_scan_log"
```

说明：

- `table_names` 限定只生成二手书 V1 三张表，避免刷新旧业务模型。
- `table_prefix: "-t_"` 表示生成结构体时去掉 `t_` 前缀。
- 预期生成或手写的模型为 `Book`、`PriceSnapshot`、`ScanLog`。
- 如果 gormt 生成结果不理想，V1 可以手写三张模型，但必须保持当前项目风格：包含 `TableName()` 和 column 常量。

## 9. 备份与清理

V1 开发阶段：

- 每日手动备份 MySQL 数据库即可。
- `scan_log` 暂不清理。
- `price_snapshot` 暂不清理。

后续生产阶段需要补充：

- 自动备份策略。
- 价格快照归档策略。
- 扫码日志保留周期。
- 敏感字段脱敏策略。

## 10. 待确认问题

- 线上 MySQL 排序规则是否支持 `utf8mb4_0900_ai_ci`。
- 开发环境是否使用 Docker 启动 MySQL。
- `client_request_id` 是否 V1 强制要求小程序传入。
- 金额 API 和数据库都使用人民币元口径，数据库字段类型为 `decimal(20,2)`。
- `book.title` 是否允许为空；当前规划为必填，缺失书名时应使用“未知书名”占位。

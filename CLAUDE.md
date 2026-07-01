# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

更完整的项目级约束见 `AGENTS.md`，业务规划见 `docs/`。

## 项目概述

这是一个基于现有 Gin + GORM + Viper + Zap 骨架二次开发的二手书扫码回收判断系统。项目支持本地独立运行，也支持通过 `api/client.go` 部署到 Vercel Serverless。

## 配置管理

项目使用 Viper 库管理配置，有两种配置模式：

1. **本地开发模式**：使用 `application.yaml` 配置文件
   - 需要在 `.env` 文件中设置 `DEV=1` 来启用此模式
   - 配置文件位于项目根目录
   - Viper 会自动向上查找最多 5 层目录来定位 `application.yaml`
   - 通过 `pkg/config` 包的 `Get`、`GetString`、`GetInt` 等函数读取配置

2. **Vercel 部署模式**：通过环境变量 `CONFIG` 传入 JSON 格式的配置
   - 配置通过 `config.Initialize()` 在启动时加载
   - JSON 配置会被 Viper 解析为配置树
   - 参考 `.env.example` 和 README.md 中的配置示例

配置读取优先级：环境变量（带 `APPENV_` 前缀）> 配置文件

## 常用命令

### 本地开发

```bash
# 本地独立运行（使用 application.yaml 配置）
DEV=1 TLS=false go run main.go -APP_PORT=8888

# 使用 Vercel CLI 本地开发（使用环境变量配置）
vercel dev

# 安装依赖
go mod tidy

# 生成/更新 GORM 模型（当前只生成 V1 三张业务表）
go run bin/gormt.go
```

### 模型管理

项目使用 `gormt` 工具自动生成 GORM 模型：
- 配置文件：`config.yml`（位于项目根目录）
- 生成的模型位置：`app/models/`
- 运行 `go run bin/gormt.go` 会根据 `config.yml` 中指定的表更新模型文件
- 当前 V1 表为 `t_book`、`t_price_snapshot`、`t_scan_log`
- 配置说明：
  - `out_dir`：输出目录（默认 `./app/models`）
  - `simple: true`：简单输出模式，只输出主键和字段标签
  - `is_out_func: true`：生成快捷函数（如 `Find`、`Update` 等）
  - `is_out_page: true`：生成分页函数
  - `table_prefix`：表前缀处理（如果以 `-` 开头表示去掉该前缀）
  - `db_info`：数据库连接信息，默认指向本地 `kfz` 数据库

## 架构说明

### 双入口设计

项目有两个入口点：

1. **main.go**：本地独立运行入口
   - 创建 HTTP 服务器，监听指定端口
   - 支持优雅关闭（graceful shutdown）
   - 启用 pprof 性能分析

2. **api/client.go**：Vercel Serverless 函数入口
   - 导出 `Handler` 函数供 Vercel 调用
   - 在 `init()` 中完成所有初始化工作
   - 所有请求通过此函数路由

### 初始化流程

两个入口都遵循相同的初始化顺序：
1. 设置时区为 CST（东八区）
2. 加载配置（`config.Initialize()`）
3. 初始化日志系统（`logger.Init()`）
4. 初始化数据库连接（`bootstrap.SetupDB()`）
5. 初始化 Redis 连接（`bootstrap.SetupRedis()`）
6. 加载模板文件（`bootstrap.SetupTemplate()`）
7. 设置路由（`bootstrap.SetupRoute()`）
8. 注册 pprof 性能分析

### 目录结构

- `app/`：应用核心代码
  - `controllers/`：控制器，分为 `client/` 和 `web/` 两个版本
  - `middlewares/`：中间件（跨域、日志追踪等）
  - `models/`：GORM 数据模型（由 gormt 自动生成）
  - `requests/`：请求验证
  - `response/`：响应结构

- `bootstrap/`：应用启动引导
  - `db.go`：数据库初始化
  - `redis.go`：Redis 初始化
  - `route.go`：路由注册
  - `template.go`：模板加载

- `config/`：配置初始化逻辑
  - 各个配置模块的初始化代码

- `pkg/`：可复用的工具包
  - `config/`：配置读取
  - `logger/`：日志系统
  - `mysql/`：MySQL 连接
  - `redis/`：Redis 连接
  - `helpers/`：辅助函数
  - `notice/`：通知服务（保留骨架能力，业务未优先使用）
  - `portscanner/`：端口扫描工具
  - 其他工具模块

- `routes/`：路由定义
  - `client/`：客户端 API 路由
  - `web/`：Web 页面路由

- `templates/`：模板文件和静态资源
  - 使用 embed 嵌入到二进制文件中

### 路由系统

路由通过 `bootstrap.SetupRoute()` 注册：
- 使用 Gin 框架
- 全局中间件：日志追踪（`TraceLogger`）、跨域（`Cors`）
- 路由分组：`/` 下注册 client 和 web 路由
- 特殊路由：`/route` 显示所有已注册的路由信息

### 模板系统

使用 CloudyKit Jet 模板引擎：
- 模板文件位于 `templates/views/`
- 静态文件位于 `templates/statics/`
- 通过 Go embed 嵌入，支持 Vercel 部署

### 数据库连接

- 使用 GORM v2
- 支持连接池配置（最大连接数、空闲连接数、连接生命周期）
- 表前缀通过配置文件设置（默认 `t_`）

### Redis 连接

- 支持密码认证
- 可配置数据库编号（默认 0）
- 在应用关闭时自动清理连接

## 开发注意事项

### 添加新路由

1. 在 `app/controllers/` 下创建控制器
2. 在 `routes/client/` 或 `routes/web/` 中注册路由
3. 路由会自动通过 `bootstrap.SetupRoute()` 加载

### 添加新模型

1. 在 MySQL 8.4 中创建表（使用配置的表前缀，默认 `t_`）
2. 确认 `config.yml` 中的 `table_names` 只包含当前业务表
3. 运行 `go run bin/gormt.go` 自动生成模型文件
4. 生成的文件位于 `app/models/`
5. 注意：gormt 会根据 `config.yml` 的配置生成模型，包括字段标签、快捷函数等

### 配置新环境

本地开发：
1. 复制 `application.yaml` 并修改数据库、Redis 配置
2. 创建 `.env` 文件，添加 `DEV=1`
3. 可选：设置 `TLS=false` 用于本地开发

Vercel 部署：
1. 在 Vercel 后台设置环境变量 `CONFIG`
2. 值为 JSON 格式，参考 README.md 中的示例
3. `vercel.json` 配置了构建和路由规则（所有请求路由到 `api/client.go`）

使用 Vercel CLI 本地开发：
1. 复制 `.env.example` 为 `.env`
2. 修改 `CONFIG` 环境变量为 JSON 格式配置
3. 运行 `vercel dev` 启动本地开发服务器

### 时区处理

项目默认使用东八区时区（CST），在 `init()` 函数中设置 `time.Local`。

## 重要技术细节

### 配置系统实现

- 使用 Viper 库进行配置管理（`pkg/config/config.go`）
- 支持点式路径访问配置：`config.GetString("redis.host")`
- 环境变量前缀：`APPENV_`（通过 `Viper.SetEnvPrefix` 设置）
- 配置加载在 `pkg/config` 包的 `init()` 函数中自动执行

### 日志系统

- 使用 Zap 日志库（`pkg/logger`）
- 日志在 `logger.Init()` 中初始化
- 支持 MySQL 日志开关（通过配置控制 debug、warn、error 级别）

### 通知系统

项目保留了通知相关基础包（`pkg/notice`），当前二手书 V1 不优先依赖这些能力。

### 性能分析

- 启用了 pprof 性能分析工具
- 通过 `pprof.Register(router)` 注册
- 可通过 `/debug/pprof/` 路径访问性能数据

### 静态资源和模板

- 使用 Go 1.16+ 的 `embed` 特性嵌入静态文件
- 模板引擎：CloudyKit Jet v6
- 静态文件路由：`/statics`
- 模板文件位于 `templates/views/`，通过 `templates.LoadTemplates()` 加载

# 二手书扫码回收判断系统

这是一个基于现有 Gin + GORM + Viper + Zap 骨架二次开发的 Go 后端项目，用于支持二手书扫码后的回收判断、价格快照和扫码历史。

更完整的协作约束见 `AGENTS.md`，业务规划见 `docs/`。

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

## Docker Compose 启动应用

Compose 只启动 Go 后端，MySQL 和 Redis 使用你已安装的服务。

```bash
docker compose up --build
```

默认会把项目根目录映射到容器内 `/config`，并从 `/config/application.yaml` 读取配置：

```text
.:/config
```

如需使用服务器上的配置目录，可通过 `CONFIG_DIR` 指定。目录里需要包含 `application.yaml`：

```bash
CONFIG_DIR=/home/ubuntu/kfz docker compose up --build
```

启动后访问：

```text
http://127.0.0.1:8888/route
http://127.0.0.1:8888/app/v1/demo/ping.json
```

## Docker 镜像上传

确认服务器 CPU 架构：

```bash
uname -m
```

如果输出是 `x86_64`，Docker 平台对应 `linux/amd64`。

如果在服务器本机上构建并上传镜像：

```bash
# kfz 服务
docker build -t kfz-app -f Dockerfile .

# kfz 镜像
docker tag kfz-app iszmxw/kfz-app:latest
docker push iszmxw/kfz-app:latest
```

如果在 Mac M 系列或其他 ARM 机器上构建，并上传给 x86_64 服务器使用：

```bash
docker buildx build \
  --platform linux/amd64 \
  -t iszmxw/kfz-app:latest \
  -f Dockerfile \
  --push \
  .
```

## Docker 镜像运行

服务器上已安装 MySQL 和 Redis 时，推荐把服务器上的配置目录映射到容器内 `/config`。项目在 `DEV=1` 时会从当前工作目录读取 `application.yaml`，所以运行时需要加上 `-w=/config`。

如果 MySQL 和 Redis 都在服务器本机，推荐使用 `--network=host`。这样配置文件里的 `DB_HOST` 和 `REDIS_HOST` 可以直接写 `127.0.0.1`。

```bash
docker run \
  --name=kfz-app \
  --network=host \
  -v=/home/ubuntu/kfz:/config:ro \
  -w=/config \
  -e DEV=1 \
  -e TLS=false \
  -d iszmxw/kfz-app:latest
```

如果不是 host 网络，可以改用端口映射，但此时配置文件里的 `DB_HOST` 和 `REDIS_HOST` 不能写 `127.0.0.1`，需要改成服务器内网 IP 或 Docker 可访问的地址。

```bash
docker run \
  --name=kfz-app \
  -p 8888:8888 \
  -v=/home/ubuntu/kfz:/config:ro \
  -w=/config \
  -e DEV=1 \
  -e TLS=false \
  -d iszmxw/kfz-app:latest
```

常用管理命令：

```bash
docker logs -f kfz-app
docker stop kfz-app
docker rm kfz-app
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

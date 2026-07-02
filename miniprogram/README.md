# 小程序 V1

微信小程序原生 TypeScript + Vant Weapp 实现。

## 本地开发

```bash
cd miniprogram
npm install
```

在微信开发者工具中导入 `miniprogram/`，然后执行：

1. 工具 -> 构建 npm
2. 详情 -> 本地设置 -> 勾选不校验合法域名
3. 启动后端：`DEV=1 TLS=false go run main.go -APP_PORT=8888`

默认接口地址在 `config/index.ts`：

```text
http://127.0.0.1:8888
```

# 安装与接入

基于 Vant Weapp 官方快速上手整理。

## 前置认知

- 目标框架是微信小程序原生开发。
- 默认读者已经理解小程序页面结构、自定义组件、微信开发者工具和 npm 构建流程。

## 安装步骤

### 1. 安装依赖

```bash
npm i @vant/weapp -S --production
```

也可以用：

```bash
yarn add @vant/weapp --production
```

### 2. 检查 `app.json`

- 移除 `"style": "v2"`，否则微信新版基础组件样式可能和 Vant Weapp 冲突。

### 3. 微信开发者工具构建 npm

- 打开微信开发者工具
- 执行“工具 -> 构建 npm”
- 勾选“使用 npm 模块”

## 组件引入

以按钮为例，在页面 `json` 或 `app.json` 中注册：

```json
{
  "usingComponents": {
    "van-button": "@vant/weapp/button/index"
  }
}
```

如果是下载源码或项目内 vendoring 的方式，常见路径改为：

```json
{
  "usingComponents": {
    "van-button": "path/to/@vant/weapp/dist/button/index"
  }
}
```

## 基础使用

```xml
<van-button type="primary">按钮</van-button>
```

## TypeScript 约定

如果项目使用 TypeScript，通常还需要：

- 安装 `miniprogram-api-typings`
- 在 `tsconfig.json` 里补 `types`
- 给 `@vant/weapp/*` 配置路径映射到 `dist/*`

## 典型坑位

- 忘记构建 npm，导致组件路径存在但开发者工具识别失败。
- 同时保留 `"style": "v2"`，导致样式混乱。
- 把 Vant Weapp 当成 Vue 组件库写，误用 `<script setup>` 或 import 组件注册方式。

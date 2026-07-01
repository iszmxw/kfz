---
name: vant-weapp
description: Vant Weapp 微信小程序组件库参考与开发技能。用于在微信小程序、微信小程序原生页面、Mini Program 自定义组件、`app.json`/页面 `json` `usingComponents` 配置、`wxml`/`wxss`/`js`/`ts` 开发中使用 `@vant/weapp`。当用户要求使用 vant-weapp、Vant Weapp、`@vant/weapp`、`van-button`、`van-field`、`van-popup`、`van-toast`、`van-dialog`、`van-cell`、`van-form`、`van-tabs` 等组件编写或修改小程序界面时使用。覆盖安装接入、npm 构建、微信开发者工具构建、样式覆盖、CSS 变量主题、弹窗与提示、表单录入、列表与导航等常见场景。
---

# Vant Weapp

`@vant/weapp` 是 Vant 的微信小程序组件库版本，适合在微信小程序原生技术栈中快速搭建移动端界面。

## 关键约定

编写 Vant Weapp 代码时始终遵守这些规则：

- 这是微信小程序组件库，不是 Vue 组件库。使用小程序原生文件结构：`json`、`wxml`、`wxss`、`js` 或 `ts`。
- 组件通过 `usingComponents` 注册，不要写成 Vue/React 的 import + JSX 形式。
- npm 安装场景的常用引用路径是 `@vant/weapp/<component>/index`。
- 下载源码或 vendoring 场景常用路径是项目内 `dist/<component>/index`。
- 页面示例优先使用 `Page({ data, methods... })` 或小程序组件写法，不要混入 Vue 语法。
- 表单输入事件常用 `bind:change`、`bind:input`，组件显隐常用 `show="{{ visible }}"` + `setData` 驱动。
- 使用 Vant Weapp 时，`app.json` 里的 `"style": "v2"` 通常需要去掉，否则新版基础组件默认样式可能影响展示。
- 修改样式优先按以下顺序考虑：
  1. 外部样式类
  2. 页面或组件级样式覆盖
  3. CSS 变量主题
- 在自定义组件中使用 Vant Weapp 并希望外层样式生效时，记得设置 `styleIsolation: 'shared'`。
- 弹窗、遮罩、滚动锁定等场景要考虑小程序环境限制，不要照搬 Web 端弹层方案。

## 快速路由

- 安装与接入：读 [getting-started.md](references/overview/getting-started.md)
- 样式覆盖：读 [styling.md](references/overview/styling.md)
- 主题定制：读 [theming.md](references/overview/theming.md)
- 常见页面骨架和开发约定：读 [architecture.md](references/overview/architecture.md)

高频组件：

- 按钮：读 [button.md](references/components/button.md)
- 输入框：读 [field.md](references/components/field.md)
- 弹出层：读 [popup.md](references/components/popup.md)
- 轻提示：读 [toast.md](references/components/toast.md)

## 常见任务

### 1. 接入新页面

1. 确认项目已安装 `@vant/weapp`，并且微信开发者工具已经执行过“构建 npm”。
2. 在页面或全局 `json` 中补 `usingComponents`。
3. 在 `wxml` 中使用 `van-` 前缀组件。
4. 在 `js/ts` 中通过 `setData` 管理状态和事件回调。
5. 若样式异常，先检查 `style: "v2"`、样式隔离和自定义样式类使用方式。

### 2. 写表单交互

- 输入组件优先看 `Field`、`Cell`、`Button`。
- 提交反馈优先看 `Toast`。
- 弹出式选择器或操作面板优先看 `Popup`。
- 表单错误提示可直接走组件自身 `error` / `error-message` 一类属性，不要先手写一套重复状态层。

### 3. 写弹层和反馈

- 内容容器优先用 `Popup`。
- 瞬时反馈优先用 `Toast`。
- 如果遇到滚动穿透，先用 `lock-scroll`，再结合 `page-meta` 的 `page-style` 方案处理。

## 组件选择建议

- 触发操作：`Button`
- 录入文本：`Field`
- 承载弹窗内容：`Popup`
- 成功/失败/加载提示：`Toast`
- 列表项容器：`Cell`、`CellGroup`
- 标签切换：`Tabs`
- 底部确认或商品操作：`SubmitBar`、`GoodsAction`

## 参考资料

只在需要时读取对应参考文件，避免一次性加载过多内容：

- `references/overview/getting-started.md`
- `references/overview/styling.md`
- `references/overview/theming.md`
- `references/overview/architecture.md`
- `references/components/button.md`
- `references/components/field.md`
- `references/components/popup.md`
- `references/components/toast.md`

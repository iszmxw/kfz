# 开发约定

这份说明用于把 Vant Weapp 的使用方式约束到小程序原生语境里。

## 文件组织

一个典型页面通常会包含：

- 页面 `json`：注册 `usingComponents`
- 页面 `wxml`：写 `van-*` 组件结构
- 页面 `wxss`：覆写样式、定义容器 CSS 变量
- 页面 `js/ts`：`Page({ data, methods })` 维护状态和事件

## 数据流

- 组件显示状态：通过 `data` + `setData` 控制
- 输入值：通过 `bind:change` / `bind:input` 或 `model:value`
- 弹层关闭：通过 `bind:close` 回写状态
- 提示调用：通常是引入函数式 API，比如 `Toast`

## 常见模式

### 按钮触发弹层

1. `data` 中维护 `showPopup`
2. 点击按钮 `setData({ showPopup: true })`
3. `Popup` 的 `bind:close` 回写为 `false`

### 输入 + 提交

1. 使用 `Field` 采集输入
2. 用按钮触发表单提交
3. 用 `Toast.loading` / `Toast.success` / `Toast.fail` 反馈结果

### 样式策略

- 优先用组件官方 props
- 其次用外部样式类
- 再用 CSS 变量
- 最后才直接覆盖内部类名

## 反模式

- 不要把它写成 Vue SFC
- 不要用 React JSX 思维去组织 `van-*` 组件
- 不要假设 Web 端弹层滚动锁定方案在小程序里完全可复用
- 不要在未构建 npm 的情况下判断组件“不可用”

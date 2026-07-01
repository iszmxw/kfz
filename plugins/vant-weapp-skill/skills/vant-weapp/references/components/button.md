# Button

## 引入

```json
{
  "usingComponents": {
    "van-button": "@vant/weapp/button/index"
  }
}
```

## 适用场景

- 表单提交
- 弹窗确认
- 列表页主要操作
- 图标型按钮

## 常用写法

### 基础类型

```xml
<van-button type="primary">主要按钮</van-button>
<van-button type="warning">警告按钮</van-button>
<van-button type="danger">危险按钮</van-button>
```

### 加载和禁用

```xml
<van-button loading type="primary" />
<van-button disabled type="info">禁用状态</van-button>
```

### 图标按钮

```xml
<van-button icon="star-o" type="primary">按钮</van-button>
```

### 块级按钮

```xml
<van-button type="primary" block>提交</van-button>
```

## 关键属性

- `type`: `default` / `primary` / `info` / `warning` / `danger`
- `size`: `large` / `normal` / `small` / `mini`
- `plain`
- `block`
- `round`
- `square`
- `disabled`
- `loading`
- `icon`
- `color`
- `form-type`

## 关键事件

- `bind:click`

注意：按钮提供的是 `click` 事件，不是原生 `tap` 事件；禁用状态下 `click` 不会触发。

## 实战建议

- 页面主 CTA 优先用 `type="primary"` + `block`
- 异步提交时优先切 `loading`
- 列表行内小操作优先用 `size="small"` 或 `mini`

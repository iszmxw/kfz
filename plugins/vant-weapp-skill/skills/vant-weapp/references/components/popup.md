# Popup

## 引入

```json
{
  "usingComponents": {
    "van-popup": "@vant/weapp/popup/index"
  }
}
```

## 适用场景

- 底部弹出操作面板
- 居中弹窗内容容器
- 顶部/侧边滑入层

## 常用写法

### 基础显隐

```xml
<van-popup show="{{ show }}" bind:close="onClose">内容</van-popup>
```

```js
Page({
  data: { show: false },
  onClose() {
    this.setData({ show: false });
  },
});
```

### 底部弹出

```xml
<van-popup
  show="{{ show }}"
  position="bottom"
  round
  custom-style="height: 40%;"
  bind:close="onClose"
/>
```

### 可关闭图标

```xml
<van-popup
  show="{{ show }}"
  closeable
  close-icon-position="top-right"
  bind:close="onClose"
/>
```

## 关键属性

- `show`
- `position`
- `overlay`
- `round`
- `closeable`
- `close-on-click-overlay`
- `lock-scroll`
- `root-portal`
- `custom-style`

## 关键事件

- `bind:close`
- `bind:click-overlay`
- 进入/离开阶段事件

## 滚动穿透建议

- 先使用 `lock-scroll`
- 若还不够，结合 `page-meta` 的 `page-style="{{ show ? 'overflow: hidden;' : '' }}"``
- 小程序环境下，弹层内容区滚动穿透不能完全按 Web 思路处理

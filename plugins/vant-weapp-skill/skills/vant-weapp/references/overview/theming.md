# 主题定制

基于 Vant Weapp 官方“定制主题”文档整理。

## 基本原则

- 小程序自定义组件机制决定了 Vant Weapp 用 CSS 变量做主题定制。
- 对不支持 CSS 变量的设备，默认样式仍会生效，但自定义主题不一定生效。

## 单组件定制

```xml
<van-button class="my-button">默认按钮</van-button>
```

```css
.my-button {
  --button-border-radius: 10px;
  --button-default-color: #f2f3f5;
}
```

也可以动态传 style：

```xml
<van-button style="{{ buttonStyle }}">默认按钮</van-button>
```

## 多组件定制

给一个公共容器挂 CSS 变量：

```xml
<view class="container">
  <van-button>默认按钮</van-button>
  <van-toast id="van-toast" />
</view>
```

```css
.container {
  --button-border-radius: 10px;
  --toast-background-color: pink;
}
```

## 全局定制

直接在 `app.wxss` 里写到 `page`：

```css
page {
  --button-border-radius: 10px;
  --toast-background-color: pink;
}
```

## 使用建议

- 单一页面局部换肤：容器级变量
- 全应用统一主题：`app.wxss` 全局变量
- 单个组件特殊风格：组件类名 + CSS 变量组合

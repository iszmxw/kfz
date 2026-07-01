# 样式覆盖

基于 Vant Weapp 官方“样式覆盖”文档整理。

## 三种主要方式

### 1. 直接覆盖组件类名

Vant Weapp 组件默认开启 `addGlobalClass: true`，可以接受外部样式影响。

页面中直接使用时，可在页面 `wxss` 覆盖：

```css
.van-button--primary {
  font-size: 20px;
  background-color: pink;
}
```

### 2. 在自定义组件中共享样式

如果在自定义组件内部使用 Vant Weapp，并希望外层样式继续生效：

```js
Component({
  options: {
    styleIsolation: 'shared',
  },
});
```

## 外部样式类

很多组件提供外部样式类，例如：

```xml
<van-cell
  title="单元格"
  value="内容"
  title-class="cell-title"
  value-class="cell-value"
/>
```

```css
.cell-title {
  color: pink !important;
}

.cell-value {
  color: green !important;
}
```

注意：

- 外部样式类与普通类优先级未定义
- 官方建议显式加 `!important`

## 实战建议

- 单个组件外观小调整：优先试外部样式类
- 页面局部统一换肤：优先试容器级 CSS 变量
- 大面积主题修改：优先读 `theming.md`

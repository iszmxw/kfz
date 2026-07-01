# Toast

## 引入

组件和函数式 API 通常一起使用：

```json
{
  "usingComponents": {
    "van-toast": "@vant/weapp/toast/index"
  }
}
```

```js
import Toast from '@vant/weapp/toast/toast';
```

## 适用场景

- 操作成功/失败提示
- 提交中加载提示
- 短消息通知

## 常用写法

### 文字提示

```js
Toast('保存成功');
```

```xml
<van-toast id="van-toast" />
```

### 加载提示

```js
Toast.loading({
  message: '加载中...',
  forbidClick: true,
});
```

### 成功失败提示

```js
Toast.success('提交成功');
Toast.fail('提交失败');
```

### 动态更新

```js
const toast = Toast.loading({
  duration: 0,
  forbidClick: true,
  message: '处理中...',
});

toast.setData({
  message: '即将完成',
});
```

## 关键点

- 常用方法：`Toast`、`Toast.loading`、`Toast.success`、`Toast.fail`、`Toast.clear`
- 长耗时流程常用 `duration: 0`
- 自定义组件内部使用时，通常要配 `selector` 和 `context`

## 实战建议

- 表单提交：先 `Toast.loading`，结束后 `Toast.success` 或 `Toast.fail`
- 纯提示文字尽量简短，官方也建议不超过十五字左右

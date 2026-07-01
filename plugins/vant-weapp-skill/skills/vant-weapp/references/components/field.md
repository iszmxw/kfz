# Field

## 引入

```json
{
  "usingComponents": {
    "van-field": "@vant/weapp/field/index",
    "van-cell-group": "@vant/weapp/cell-group/index"
  }
}
```

## 适用场景

- 用户名、手机号、验证码输入
- 多行文本输入
- 表单错误提示
- 输入尾部附加按钮

## 常用写法

### 基础输入

```xml
<van-cell-group>
  <van-field
    value="{{ value }}"
    placeholder="请输入用户名"
    bind:change="onChange"
  />
</van-cell-group>
```

### 双向绑定

```xml
<van-field model:value="{{ value }}" placeholder="请输入用户名" />
```

### 错误提示

```xml
<van-field
  value="{{ phone }}"
  label="手机号"
  error-message="手机号格式错误"
/>
```

### textarea 自适应

```xml
<van-field
  value="{{ message }}"
  type="textarea"
  autosize
  placeholder="请输入留言"
/>
```

### 尾部按钮

```xml
<van-field
  value="{{ sms }}"
  label="短信验证码"
  use-button-slot
>
  <van-button slot="button" size="small" type="primary">
    发送验证码
  </van-button>
</van-field>
```

## 关键点

- 值变更常用 `bind:change`
- 高基础库版本可用 `model:value`
- 表单错误优先用组件自带 `error` / `error-message`
- 如果要替换输入值或控制光标，可使用 `callback`

## 实战建议

- 小程序原生输入框聚焦时可能出现 placeholder 加粗、闪烁，这是环境特性，不一定是代码错误
- 表单录入页面常和 `Button`、`Toast`、`Popup` 联动使用

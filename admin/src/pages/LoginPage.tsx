import { Button, Form, Input, Message, Typography } from '@arco-design/web-react';
import { IconLock, IconUser } from '@arco-design/web-react/icon';
import { post } from '../api';
import type { LoginResponse } from '../types';

interface Props {
  onLogin: (data: LoginResponse) => void;
}

export default function LoginPage({ onLogin }: Props) {
  const [form] = Form.useForm();

  const submit = async () => {
    const values = await form.validate();
    const data = await post<LoginResponse>('/auth/login', values);
    Message.success('登录成功');
    onLogin(data);
  };

  return (
    <div className="login-page">
      <div className="login-panel">
        <Typography.Title heading={3}>旧书回收管理后台</Typography.Title>
        <Form form={form} layout="vertical" autoComplete="off">
          <Form.Item field="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input prefix={<IconUser />} placeholder="admin" />
          </Form.Item>
          <Form.Item field="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password prefix={<IconLock />} placeholder="请输入密码" />
          </Form.Item>
          <Button type="primary" long onClick={submit}>
            登录
          </Button>
        </Form>
      </div>
    </div>
  );
}

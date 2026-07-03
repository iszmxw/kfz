import { useEffect, useMemo, useState } from 'react';
import { Button, Card, Form, Input, Message, Modal, Select, Space, Table, Tabs, Tag, Typography } from '@arco-design/web-react';
import { IconEdit, IconPlus } from '@arco-design/web-react/icon';
import { get, post } from '../api';
import type { AdminRole, AdminUserListItem, PageResponse } from '../types';

function SimpleTable({ path }: { path: string }) {
  const [data, setData] = useState<Record<string, unknown>[]>([]);
  useEffect(() => {
    get<PageResponse<Record<string, unknown>>>(`${path}?page=1&page_size=50`).then((res) => setData(res.items));
  }, [path]);
  const keys = Object.keys(data[0] || {}).slice(0, 6);
  return <Table rowKey={(record) => String(record.id)} data={data} columns={keys.map((key) => ({ title: key, dataIndex: key }))} />;
}

function UserManager() {
  const [users, setUsers] = useState<AdminUserListItem[]>([]);
  const [roles, setRoles] = useState<AdminRole[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [visible, setVisible] = useState(false);
  const [editing, setEditing] = useState<AdminUserListItem | null>(null);
  const [form] = Form.useForm();

  const roleOptions = useMemo(() => roles.map((role) => ({ label: `${role.name} (${role.code})`, value: role.id })), [roles]);
  const operatorRoleID = useMemo(() => roles.find((role) => role.code === 'OPERATOR')?.id, [roles]);

  const loadUsers = () => {
    const query = keyword ? `&keyword=${encodeURIComponent(keyword)}` : '';
    return get<PageResponse<AdminUserListItem>>(`/system/user/list?page=${page}&page_size=20${query}`).then((res) => {
      setUsers(res.items);
      setTotal(res.total);
    });
  };

  const loadRoles = () =>
    get<PageResponse<AdminRole>>('/system/role/list?page=1&page_size=100').then((res) => {
      setRoles(res.items);
    });

  useEffect(() => {
    loadRoles();
  }, []);

  useEffect(() => {
    loadUsers();
  }, [page, keyword]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({ status: 'ACTIVE', role_ids: operatorRoleID ? [operatorRoleID] : [] });
    setVisible(true);
  };

  const openEdit = (record: AdminUserListItem) => {
    setEditing(record);
    form.resetFields();
    form.setFieldsValue({
      id: record.id,
      username: record.username,
      name: record.name,
      status: record.status,
      role_ids: record.roleIds || []
    });
    setVisible(true);
  };

  const submit = async () => {
    const values = await form.validate();
    await post('/system/user/save', { ...values, id: editing?.id || 0 });
    Message.success('用户已保存');
    setVisible(false);
    form.resetFields();
    loadUsers();
  };

  return (
    <>
      <Space className="filter-bar">
        <Input.Search
          allowClear
          placeholder="用户名 / 姓名"
          style={{ width: 260 }}
          onSearch={(value) => {
            setPage(1);
            setKeyword(value);
          }}
        />
        <Button type="primary" icon={<IconPlus />} onClick={openCreate}>新建用户</Button>
      </Space>
      <Table
        rowKey="id"
        data={users}
        pagination={{ current: page, pageSize: 20, total, onChange: setPage }}
        columns={[
          { title: 'ID', dataIndex: 'id', width: 90 },
          { title: '用户名', dataIndex: 'username' },
          { title: '姓名', dataIndex: 'name' },
          { title: '状态', dataIndex: 'status', render: (value) => <Tag color={value === 'ACTIVE' ? 'green' : 'gray'}>{value}</Tag> },
          { title: '角色', render: (_, record) => (record.roles || []).join(', ') },
          { title: '最近登录', dataIndex: 'lastLoginAt' },
          { title: '操作', width: 100, render: (_, record) => <Button type="text" icon={<IconEdit />} onClick={() => openEdit(record)} /> }
        ]}
      />
      <Modal title={editing ? '编辑用户' : '新建用户'} visible={visible} onOk={submit} onCancel={() => setVisible(false)} afterClose={() => form.resetFields()}>
        <Form form={form} layout="vertical">
          <Form.Item field="username" label="用户名" rules={[{ required: true }]}>
            <Input disabled={Boolean(editing)} />
          </Form.Item>
          <Form.Item field="name" label="姓名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item field="password" label={editing ? '重置密码' : '密码'} rules={editing ? [] : [{ required: true, message: '请输入初始密码' }]}>
            <Input.Password placeholder={editing ? '留空则不修改密码' : '请输入初始密码'} />
          </Form.Item>
          <Form.Item field="status" label="状态" rules={[{ required: true }]}>
            <Select options={[{ label: '启用', value: 'ACTIVE' }, { label: '禁用', value: 'DISABLED' }]} />
          </Form.Item>
          <Form.Item field="role_ids" label="角色" rules={[{ required: true, message: '请选择角色' }]}>
            <Select mode="multiple" options={roleOptions} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

export default function SystemPage() {
  return (
    <div>
      <Typography.Title heading={4}>系统权限</Typography.Title>
      <Card>
        <Tabs defaultActiveTab="users">
          <Tabs.TabPane key="users" title="用户"><UserManager /></Tabs.TabPane>
          <Tabs.TabPane key="roles" title="角色"><SimpleTable path="/system/role/list" /></Tabs.TabPane>
          <Tabs.TabPane key="menus" title="菜单"><SimpleTable path="/system/menu/list" /></Tabs.TabPane>
          <Tabs.TabPane key="apis" title="API权限"><SimpleTable path="/system/api-permission/list" /></Tabs.TabPane>
          <Tabs.TabPane key="logs" title="操作日志"><SimpleTable path="/system/operation-log/list" /></Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  );
}

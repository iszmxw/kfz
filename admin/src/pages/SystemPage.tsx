import { useEffect, useState } from 'react';
import { Card, Tabs, Table, Typography } from '@arco-design/web-react';
import { get } from '../api';
import type { PageResponse } from '../types';

function SimpleTable({ path }: { path: string }) {
  const [data, setData] = useState<Record<string, unknown>[]>([]);
  useEffect(() => {
    get<PageResponse<Record<string, unknown>>>(`${path}?page=1&page_size=50`).then((res) => setData(res.items));
  }, [path]);
  const keys = Object.keys(data[0] || {}).slice(0, 6);
  return <Table rowKey={(record) => String(record.id)} data={data} columns={keys.map((key) => ({ title: key, dataIndex: key }))} />;
}

export default function SystemPage() {
  return (
    <div>
      <Typography.Title heading={4}>系统权限</Typography.Title>
      <Card>
        <Tabs defaultActiveTab="users">
          <Tabs.TabPane key="users" title="用户"><SimpleTable path="/system/user/list" /></Tabs.TabPane>
          <Tabs.TabPane key="roles" title="角色"><SimpleTable path="/system/role/list" /></Tabs.TabPane>
          <Tabs.TabPane key="menus" title="菜单"><SimpleTable path="/system/menu/list" /></Tabs.TabPane>
          <Tabs.TabPane key="apis" title="API权限"><SimpleTable path="/system/api-permission/list" /></Tabs.TabPane>
          <Tabs.TabPane key="logs" title="操作日志"><SimpleTable path="/system/operation-log/list" /></Tabs.TabPane>
        </Tabs>
      </Card>
    </div>
  );
}

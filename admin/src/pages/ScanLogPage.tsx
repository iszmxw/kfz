import { useEffect, useState } from 'react';
import { Form, Input, Button, Select, Table, Typography, Space } from '@arco-design/web-react';
import { get } from '../api';
import type { PageResponse, ScanLog } from '../types';

export default function ScanLogPage() {
  const [form] = Form.useForm();
  const [data, setData] = useState<ScanLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [params, setParams] = useState('');

  const load = () => get<PageResponse<ScanLog>>(`/scan-log/list?page=${page}&page_size=20${params}`).then((res) => {
    setData(res.items);
    setTotal(res.total);
  });
  useEffect(() => { load(); }, [page, params]);

  return (
    <div>
      <Typography.Title heading={4}>扫码日志</Typography.Title>
      <Form form={form} layout="inline" className="filter-bar">
        <Form.Item field="isbn" label="ISBN"><Input placeholder="ISBN" /></Form.Item>
        <Form.Item field="decision" label="判断"><Select allowClear options={['ACCEPT', 'REJECT', 'NEED_REVIEW']} /></Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" onClick={() => {
              const values = form.getFieldsValue();
              setPage(1);
              setParams(`${values.isbn ? `&isbn=${values.isbn}` : ''}${values.decision ? `&decision=${values.decision}` : ''}`);
            }}>查询</Button>
            <Button onClick={() => { form.resetFields(); setParams(''); }}>重置</Button>
          </Space>
        </Form.Item>
      </Form>
      <Table rowKey="id" data={data} pagination={{ current: page, pageSize: 20, total, onChange: setPage }} columns={[
        { title: '扫码ID', dataIndex: 'id', width: 220 },
        { title: 'ISBN', render: (_, record) => record.normalized_isbn || record.normalizedIsbn },
        { title: '判断', dataIndex: 'decision' },
        { title: '原因', dataIndex: 'reason' },
        { title: '可信度', dataIndex: 'confidence' },
        { title: '时间', render: (_, record) => record.scanned_at || record.scannedAt }
      ]} />
    </div>
  );
}

import { useEffect, useState } from 'react';
import { Button, Form, Input, InputNumber, Modal, Select, Space, Table, Typography, Message } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { get, post } from '../api';
import type { PageResponse, PriceSnapshot } from '../types';

export default function PriceSnapshotPage() {
  const [data, setData] = useState<PriceSnapshot[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [isbn, setIsbn] = useState('');
  const [visible, setVisible] = useState(false);
  const [form] = Form.useForm();
  const load = () => get<PageResponse<PriceSnapshot>>(`/price-snapshot/list?page=${page}&page_size=20${isbn ? `&isbn=${isbn}` : ''}`).then((res) => {
    setData(res.items);
    setTotal(res.total);
  });
  useEffect(() => { load(); }, [page, isbn]);
  const submit = async () => {
    const values = await form.validate();
    await post('/price-snapshot/create', {
      ...values,
      min_price: String(values.min_price || ''),
      avg_price: String(values.avg_price || ''),
      max_price: String(values.max_price || '')
    });
    Message.success('已新增价格快照');
    setVisible(false);
    form.resetFields();
    load();
  };
  return (
    <div>
      <Typography.Title heading={4}>价格快照</Typography.Title>
      <Space className="filter-bar">
        <Input.Search placeholder="ISBN" style={{ width: 240 }} onSearch={(v) => { setPage(1); setIsbn(v); }} />
        <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>新增快照</Button>
      </Space>
      <Table rowKey="id" data={data} pagination={{ current: page, pageSize: 20, total, onChange: setPage }} columns={[
        { title: 'ISBN', dataIndex: 'isbn' },
        { title: '来源', dataIndex: 'source' },
        { title: '最低价', render: (_, r) => r.minPrice },
        { title: '均价', render: (_, r) => r.avgPrice },
        { title: '最高价', render: (_, r) => r.maxPrice },
        { title: '样本', render: (_, r) => r.sampleCount },
        { title: '可信度', dataIndex: 'confidence' }
      ]} />
      <Modal title="新增价格快照" visible={visible} onOk={submit} onCancel={() => setVisible(false)} afterClose={() => form.resetFields()}>
        <Form form={form} layout="vertical" initialValues={{ source: 'manual', confidence: 'MEDIUM', sample_count: 1 }}>
          <Form.Item field="isbn" label="ISBN" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item field="source" label="来源"><Input /></Form.Item>
          <Form.Item field="min_price" label="最低价"><InputNumber min={0} precision={2} style={{ width: '100%' }} /></Form.Item>
          <Form.Item field="avg_price" label="均价" rules={[{ required: true }]}><InputNumber min={0} precision={2} style={{ width: '100%' }} /></Form.Item>
          <Form.Item field="max_price" label="最高价"><InputNumber min={0} precision={2} style={{ width: '100%' }} /></Form.Item>
          <Form.Item field="sample_count" label="样本数"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          <Form.Item field="confidence" label="可信度"><Select options={['HIGH', 'MEDIUM', 'LOW', 'NONE']} /></Form.Item>
          <Form.Item field="raw_url" label="来源链接"><Input /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

import { useEffect, useState } from 'react';
import { Button, Form, Input, InputNumber, Modal, Select, Table, Typography, Message } from '@arco-design/web-react';
import { get, post } from '../api';
import type { PageResponse, ScanLog } from '../types';

export default function ManualReviewPage() {
  const [data, setData] = useState<ScanLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [visible, setVisible] = useState(false);
  const [current, setCurrent] = useState<ScanLog | null>(null);
  const [form] = Form.useForm();

  const load = () => get<PageResponse<ScanLog>>(`/manual-review/list?page=${page}&page_size=20`).then((res) => {
    setData(res.items);
    setTotal(res.total);
  });

  useEffect(() => { load(); }, [page]);

  const submit = async () => {
    const values = await form.validate();
    await post('/manual-review/decide', { ...values, scan_log_id: current?.id, actual_recycle_price: String(values.actual_recycle_price || '') });
    Message.success('已提交人工确认');
    setVisible(false);
    form.resetFields();
    load();
  };

  return (
    <div>
      <Typography.Title heading={4}>人工确认</Typography.Title>
      <Table
        rowKey="id"
        data={data}
        pagination={{ current: page, pageSize: 20, total, onChange: setPage }}
        columns={[
          { title: '扫码ID', dataIndex: 'id', width: 220 },
          { title: 'ISBN', render: (_, record) => record.normalized_isbn || record.normalizedIsbn },
          { title: '系统判断', dataIndex: 'decision' },
          { title: '原因', dataIndex: 'reason' },
          { title: '可信度', dataIndex: 'confidence' },
          { title: '操作', width: 120, render: (_, record) => <Button type="text" onClick={() => { setCurrent(record); setVisible(true); }}>确认</Button> }
        ]}
      />
      <Modal title="人工确认" visible={visible} onOk={submit} onCancel={() => setVisible(false)} afterClose={() => form.resetFields()}>
        <Form form={form} layout="vertical" initialValues={{ manual_decision: 'ACCEPT' }}>
          <Form.Item field="manual_decision" label="人工结果" rules={[{ required: true }]}>
            <Select options={[{ label: '收', value: 'ACCEPT' }, { label: '不收', value: 'REJECT' }]} />
          </Form.Item>
          <Form.Item field="actual_recycle_price" label="实际回收价">
            <InputNumber min={0} precision={2} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item field="note" label="备注">
            <Input.TextArea />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

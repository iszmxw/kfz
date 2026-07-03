import { useEffect, useState } from 'react';
import { Button, Card, Form, Input, InputNumber, Message, Select, Typography } from '@arco-design/web-react';
import { get, post } from '../api';

export default function RecycleRulePage() {
  const [form] = Form.useForm();
  const [current, setCurrent] = useState<Record<string, unknown> | null>(null);
  useEffect(() => {
    get<Record<string, unknown> | null>('/recycle-rule/current').then((rule) => {
      setCurrent(rule);
      if (rule) {
        form.setFieldsValue({
          name: rule.name,
          min_accept_avg_price: rule.minAcceptAvgPrice,
          recycle_rate: rule.recycleRate,
          min_sample_count: rule.minSampleCount,
          price_valid_days: rule.priceValidDays,
          low_confidence_mode: rule.lowConfidenceMode
        });
      }
    });
  }, [form]);
  const submit = async () => {
    const values = await form.validate();
    await post('/recycle-rule/save-and-enable', {
      ...values,
      min_accept_avg_price: String(values.min_accept_avg_price),
      recycle_rate: String(values.recycle_rate)
    });
    Message.success('规则已启用');
  };
  return (
    <div>
      <Typography.Title heading={4}>规则配置</Typography.Title>
      <Card title="启用中的规则" style={{ maxWidth: 720 }}>
        <Form form={form} layout="vertical" initialValues={{ min_accept_avg_price: 10, recycle_rate: 0.3, min_sample_count: 3, price_valid_days: 30, low_confidence_mode: 'NEED_REVIEW' }}>
          <Form.Item field="name" label="规则名称" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item field="min_accept_avg_price" label="最低可收均价" rules={[{ required: true }]}><InputNumber min={0} precision={2} style={{ width: '100%' }} /></Form.Item>
          <Form.Item field="recycle_rate" label="回收比例" rules={[{ required: true }]}><InputNumber min={0} max={1} step={0.01} precision={4} style={{ width: '100%' }} /></Form.Item>
          <Form.Item field="min_sample_count" label="最低样本数"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item field="price_valid_days" label="价格有效天数"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item field="low_confidence_mode" label="低可信处理"><Select options={['NEED_REVIEW']} /></Form.Item>
          <Button type="primary" onClick={submit}>保存并启用</Button>
        </Form>
        {current ? <Typography.Text type="secondary">当前版本：{String(current.version || '')}</Typography.Text> : null}
      </Card>
    </div>
  );
}

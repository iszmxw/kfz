import { useEffect, useState } from 'react';
import { Card, Grid, Statistic, Typography } from '@arco-design/web-react';
import { get } from '../api';

const { Row, Col } = Grid;

interface Summary {
  today_scan_count: number;
  pending_review_count: number;
  accepted_today_count: number;
  suggested_total_price: number;
  low_confidence_count: number;
}

export default function DashboardPage() {
  const [summary, setSummary] = useState<Summary | null>(null);

  useEffect(() => {
    get<Summary>('/dashboard/summary').then(setSummary);
  }, []);

  return (
    <div>
      <Typography.Title heading={4}>工作台</Typography.Title>
      <Row gutter={[16, 16]}>
        <Col span={6}><Card><Statistic title="今日扫码" value={summary?.today_scan_count || 0} /></Card></Col>
        <Col span={6}><Card><Statistic title="待人工确认" value={summary?.pending_review_count || 0} /></Card></Col>
        <Col span={6}><Card><Statistic title="今日建议回收" value={summary?.accepted_today_count || 0} /></Card></Col>
        <Col span={6}><Card><Statistic title="建议金额" value={summary?.suggested_total_price || 0} precision={2} prefix="¥" /></Card></Col>
      </Row>
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={12}><Card title="待处理事项">低可信价格：{summary?.low_confidence_count || 0}</Card></Col>
        <Col span={12}><Card title="运营提示">优先处理人工确认和导入失败行。</Card></Col>
      </Row>
    </div>
  );
}

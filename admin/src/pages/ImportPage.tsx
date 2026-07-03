import { useEffect, useState } from 'react';
import { Table, Typography, Upload, Message, Card } from '@arco-design/web-react';
import { get, request } from '../api';
import type { PageResponse } from '../types';

interface ImportTask {
  id: number;
  fileName?: string;
  file_name?: string;
  status: string;
  totalRows?: number;
  total_rows?: number;
  successRows?: number;
  success_rows?: number;
  failedRows?: number;
  failed_rows?: number;
}

export default function ImportPage() {
  const [data, setData] = useState<ImportTask[]>([]);
  const load = () => get<PageResponse<ImportTask>>('/import/list?page=1&page_size=50').then((res) => setData(res.items));
  useEffect(() => { load(); }, []);
  return (
    <div>
      <Typography.Title heading={4}>导入任务</Typography.Title>
      <Card className="filter-bar">
        <Upload
          drag
          accept={{ type: '.csv,.xlsx', strict: true }}
          customRequest={({ file, onSuccess, onError }) => {
            const formData = new FormData();
            formData.append('file', file as File);
            request('/import/upload', { method: 'POST', body: formData })
              .then((res) => { Message.success('导入完成'); onSuccess?.({ data: res }); load(); })
              .catch(onError);
            return { abort: () => undefined };
          }}
        />
      </Card>
      <Table rowKey="id" data={data} columns={[
        { title: '任务ID', dataIndex: 'id', width: 220 },
        { title: '文件名', render: (_, r) => r.file_name || r.fileName },
        { title: '状态', dataIndex: 'status' },
        { title: '总行数', render: (_, r) => r.total_rows ?? r.totalRows },
        { title: '成功', render: (_, r) => r.success_rows ?? r.successRows },
        { title: '失败', render: (_, r) => r.failed_rows ?? r.failedRows }
      ]} />
    </div>
  );
}

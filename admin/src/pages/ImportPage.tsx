import { useEffect, useState } from 'react';
import { Button, Card, Form, Input, InputNumber, Message, Modal, Select, Space, Table, Tag, Typography, Upload } from '@arco-design/web-react';
import { IconEye, IconRefresh, IconSync, IconUpload } from '@arco-design/web-react/icon';
import { get, post, request } from '../api';
import { BookCover } from '../components/BookCover';
import type { KongfzCollectRawRow, KongfzCollectTask, PageResponse } from '../types';

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

interface KongfzTaskDetail {
  task: KongfzCollectTask;
  rows: KongfzCollectRawRow[];
}

const statusColor: Record<string, string> = {
  PENDING: 'gray',
  RUNNING: 'blue',
  STOP_REQUESTED: 'orange',
  STOPPED: 'orange',
  COMPLETED: 'green',
  FAILED: 'red'
};

const kongfzCategories = [
  { name: '小说', catId: 43 },
  { name: '文学', catId: 1 },
  { name: '语言文字', catId: 13 },
  { name: '历史', catId: 3 },
  { name: '地理', catId: 23 },
  { name: '艺术', catId: 4 },
  { name: '政治', catId: 18 },
  { name: '法律', catId: 5 },
  { name: '军事', catId: 28 },
  { name: '哲学心理学', catId: 44 },
  { name: '宗教', catId: 21 },
  { name: '经济', catId: 6 },
  { name: '社会文化', catId: 7 },
  { name: '教育', catId: 20 },
  { name: '管理', catId: 42 },
  { name: '童书', catId: 16 },
  { name: '生活', catId: 26 },
  { name: '体育', catId: 17 },
  { name: '工程技术', catId: 9 },
  { name: '计算机与互联网', catId: 24 },
  { name: '自然科学', catId: 8 },
  { name: '医药卫生', catId: 25 },
  { name: '综合性图书', catId: 29 },
  { name: '教材教辅考试', catId: 33 },
  { name: '国学古籍', catId: 12 },
  { name: '收藏与鉴赏', catId: 59 },
  { name: '红色文献', catId: 34 },
  { name: '连环画', catId: 35 }
];

const kongfzCategoryOptions = kongfzCategories.map((item) => ({
  label: `${item.name}（${item.catId}）`,
  value: item.catId
}));

function formatKongfzCategory(catId?: number | null) {
  if (!catId) {
    return '-';
  }
  const category = kongfzCategories.find((item) => item.catId === catId);
  return category ? `${category.name}（${category.catId}）` : catId;
}

function rawRowCoverUrl(row: KongfzCollectRawRow) {
  return row.coverUrl || row.cover_url || '';
}

export default function ImportPage() {
  const [imports, setImports] = useState<ImportTask[]>([]);
  const [collectTasks, setCollectTasks] = useState<KongfzCollectTask[]>([]);
  const [collectTotal, setCollectTotal] = useState(0);
  const [collectPage, setCollectPage] = useState(1);
  const [createVisible, setCreateVisible] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detail, setDetail] = useState<KongfzTaskDetail | null>(null);
  const [form] = Form.useForm();

  const loadImports = () => get<PageResponse<ImportTask>>('/import/list?page=1&page_size=50').then((res) => setImports(res.items));
  const loadCollectTasks = () => get<PageResponse<KongfzCollectTask>>(`/kongfz-collect/task/list?page=${collectPage}&page_size=20`).then((res) => {
    setCollectTasks(res.items);
    setCollectTotal(res.total);
  });
  const loadAll = () => {
    void loadImports();
    void loadCollectTasks();
  };

  useEffect(() => { loadImports(); }, []);
  useEffect(() => { loadCollectTasks(); }, [collectPage]);
  useEffect(() => {
    const timer = window.setInterval(loadCollectTasks, 3000);
    return () => window.clearInterval(timer);
  }, [collectPage]);

  const createTask = async () => {
    const values = await form.validate();
    await post('/kongfz-collect/task/create', values);
    Message.success('采集任务已创建');
    setCreateVisible(false);
    form.resetFields();
    setCollectPage(1);
    await loadCollectTasks();
  };

  const stopTask = async (id: number) => {
    await post('/kongfz-collect/task/stop', { id });
    Message.success('已请求停止');
    await loadCollectTasks();
  };

  const retryTask = async (id: number) => {
    await post('/kongfz-collect/task/retry', { id });
    Message.success('已创建重试任务');
    setCollectPage(1);
    await loadCollectTasks();
  };

  const openDetail = async (id: number) => {
    const data = await get<KongfzTaskDetail>(`/kongfz-collect/task/detail?id=${id}&row_limit=100`);
    setDetail(data);
    setDetailVisible(true);
  };

  return (
    <div>
      <Typography.Title heading={4}>导入任务</Typography.Title>

      <Card className="filter-bar" title="孔夫子采集">
        <Space style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<IconSync />} onClick={() => setCreateVisible(true)}>创建采集任务</Button>
          <Button icon={<IconRefresh />} onClick={loadAll}>刷新</Button>
        </Space>
        <Table rowKey="id" data={collectTasks} pagination={{ current: collectPage, pageSize: 20, total: collectTotal, onChange: setCollectPage }} columns={[
          { title: '任务ID', dataIndex: 'id', width: 100 },
          { title: '分类', render: (_, r) => formatKongfzCategory(r.cat_id ?? r.catId) },
          { title: '页码', render: (_, r) => `${r.start_page ?? r.startPage}-${r.end_page || r.endPage || '自动'} / 当前 ${r.current_page ?? r.currentPage ?? 0}` },
          { title: '状态', render: (_, r) => <Tag color={statusColor[r.status] || 'gray'}>{r.status}</Tag> },
          { title: '采集', render: (_, r) => `${r.collected_count ?? r.collectedCount ?? 0} 条` },
          { title: '有效/无效', render: (_, r) => `${r.valid_count ?? r.validCount ?? 0}/${r.invalid_count ?? r.invalidCount ?? 0}` },
          { title: '已导入', render: (_, r) => r.imported_count ?? r.importedCount ?? 0 },
          { title: '错误', render: (_, r) => r.error_message || r.errorMessage || '-' },
          {
            title: '操作',
            width: 210,
            render: (_, r) => (
              <Space>
                <Button type="text" icon={<IconEye />} onClick={() => openDetail(r.id)}>详情</Button>
                <Button type="text" disabled={!['PENDING', 'RUNNING'].includes(r.status)} onClick={() => stopTask(r.id)}>停止</Button>
                <Button type="text" disabled={!['FAILED', 'STOPPED'].includes(r.status)} onClick={() => retryTask(r.id)}>重试</Button>
              </Space>
            )
          }
        ]} />
      </Card>

      <Card className="filter-bar" title="文件导入">
        <Upload
          drag
          accept={{ type: '.csv,.xlsx', strict: true }}
          customRequest={({ file, onSuccess, onError }) => {
            const formData = new FormData();
            formData.append('file', file as File);
            request('/import/upload', { method: 'POST', body: formData })
              .then((res) => { Message.success('导入完成'); onSuccess?.({ data: res }); loadImports(); })
              .catch(onError);
            return { abort: () => undefined };
          }}
        >
          <div>
            <IconUpload style={{ fontSize: 28, marginBottom: 8 }} />
            <div>拖拽或点击上传 CSV/XLSX</div>
          </div>
        </Upload>
      </Card>

      <Table rowKey="id" data={imports} columns={[
        { title: '任务ID', dataIndex: 'id', width: 120 },
        { title: '文件名', render: (_, r) => r.file_name || r.fileName },
        { title: '状态', dataIndex: 'status' },
        { title: '总行数', render: (_, r) => r.total_rows ?? r.totalRows },
        { title: '成功', render: (_, r) => r.success_rows ?? r.successRows },
        { title: '失败', render: (_, r) => r.failed_rows ?? r.failedRows }
      ]} />

      <Modal title="创建孔夫子采集任务" visible={createVisible} onOk={createTask} onCancel={() => setCreateVisible(false)} afterClose={() => form.resetFields()}>
        <Form form={form} layout="vertical" initialValues={{ cat_id: 43, start_page: 1, delay_ms: 1200, retry_times: 2 }}>
          <Form.Item field="cat_id" label="图书分类" rules={[{ required: true }]}>
            <Select showSearch placeholder="请选择孔夫子图书分类" options={kongfzCategoryOptions} />
          </Form.Item>
          <Form.Item field="start_page" label="起始页" rules={[{ required: true }]}><InputNumber min={1} /></Form.Item>
          <Form.Item field="end_page" label="结束页"><InputNumber min={1} placeholder="留空时自动，最多100页" /></Form.Item>
          <Form.Item field="delay_ms" label="限速间隔 ms"><InputNumber min={0} /></Form.Item>
          <Form.Item field="retry_times" label="重试次数"><InputNumber min={0} /></Form.Item>
          <Form.Item field="user_agent" label="User-Agent"><Input /></Form.Item>
          <Form.Item field="cookie" label="Cookie"><Input.Password /></Form.Item>
        </Form>
      </Modal>

      <Modal title={`采集详情${detail?.task?.id ? ` #${detail.task.id}` : ''}`} visible={detailVisible} footer={null} onCancel={() => setDetailVisible(false)} style={{ width: 980 }}>
        <Table rowKey="id" data={detail?.rows || []} pagination={false} columns={[
          { title: '页码', dataIndex: 'page', width: 80 },
          { title: '序号', render: (_, r) => r.row_index ?? r.rowIndex },
          { title: '封面', width: 88, render: (_, r) => <BookCover url={rawRowCoverUrl(r)} title={r.title} /> },
          { title: 'ISBN', render: (_, r) => r.normalized_isbn || r.normalizedIsbn || '-' },
          { title: '书名', dataIndex: 'title' },
          { title: '有效', render: (_, r) => r.valid ? <Tag color="green">有效</Tag> : <Tag color="red">无效</Tag> },
          { title: '同步', render: (_, r) => r.sync_status || r.syncStatus },
          { title: '导入行', render: (_, r) => r.import_task_row_id || r.importTaskRowId || '-' },
          { title: '错误', render: (_, r) => r.error_message || r.errorMessage || '-' }
        ]} />
      </Modal>
    </div>
  );
}

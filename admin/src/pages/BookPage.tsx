import { useEffect, useState } from 'react';
import { Button, Form, Input, Modal, Space, Table, Typography, Message } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { get, post } from '../api';
import { BookCover } from '../components/BookCover';
import type { Book, PageResponse } from '../types';

function bookCoverUrl(book: Book) {
  return book.coverUrl || book.cover_url || '';
}

function formValuesFromBook(book: Book) {
  return {
    ...book,
    publish_year: book.publish_year || book.publishYear || '',
    cover_url: bookCoverUrl(book),
  };
}

export default function BookPage() {
  const [data, setData] = useState<Book[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState('');
  const [visible, setVisible] = useState(false);
  const [coverPreview, setCoverPreview] = useState('');
  const [editingTitle, setEditingTitle] = useState('');
  const [form] = Form.useForm();

  const load = () => get<PageResponse<Book>>(`/book/list?page=${page}&page_size=20${keyword ? `&keyword=${keyword}` : ''}`).then((res) => {
    setData(res.items);
    setTotal(res.total);
  });
  useEffect(() => { load(); }, [page, keyword]);

  const submit = async () => {
    const values = await form.validate();
    await post('/book/save', values);
    Message.success('已保存');
    setVisible(false);
    form.resetFields();
    setCoverPreview('');
    setEditingTitle('');
    load();
  };

  const openCreate = () => {
    form.resetFields();
    setCoverPreview('');
    setEditingTitle('');
    setVisible(true);
  };

  const openEdit = (record: Book) => {
    const values = formValuesFromBook(record);
    form.setFieldsValue(values);
    setCoverPreview(values.cover_url);
    setEditingTitle(record.title);
    setVisible(true);
  };

  return (
    <div>
      <Typography.Title heading={4}>书籍库</Typography.Title>
      <Space className="filter-bar">
        <Input.Search placeholder="ISBN / 书名 / 作者" style={{ width: 280 }} onSearch={(v) => { setPage(1); setKeyword(v); }} />
        <Button type="primary" icon={<IconPlus />} onClick={openCreate}>新增书籍</Button>
      </Space>
      <Table rowKey="isbn" data={data} pagination={{ current: page, pageSize: 20, total, onChange: setPage }} columns={[
        { title: '封面', width: 88, render: (_, record) => <BookCover url={bookCoverUrl(record)} title={record.title} /> },
        { title: 'ISBN', dataIndex: 'isbn', width: 160 },
        { title: '书名', dataIndex: 'title' },
        { title: '作者', dataIndex: 'author' },
        { title: '出版社', dataIndex: 'publisher' },
        { title: '来源', dataIndex: 'source' },
        { title: '操作', width: 100, render: (_, record) => <Button type="text" onClick={() => openEdit(record)}>编辑</Button> }
      ]} />
      <Modal title="书籍信息" visible={visible} onOk={submit} onCancel={() => setVisible(false)} afterClose={() => { form.resetFields(); setCoverPreview(''); setEditingTitle(''); }}>
        <Form
          form={form}
          layout="vertical"
          onValuesChange={(_, values) => {
            setCoverPreview(values.cover_url || '');
            setEditingTitle(values.title || '');
          }}
        >
          <div className="book-form-cover">
            <BookCover url={coverPreview} title={editingTitle} size="preview" />
          </div>
          <Form.Item field="isbn" label="ISBN" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item field="title" label="书名" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item field="author" label="作者"><Input /></Form.Item>
          <Form.Item field="publisher" label="出版社"><Input /></Form.Item>
          <Form.Item field="publish_year" label="出版年份"><Input /></Form.Item>
          <Form.Item field="cover_url" label="封面URL"><Input placeholder="https://..." /></Form.Item>
          <Form.Item field="source" label="来源" initialValue="manual"><Input /></Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

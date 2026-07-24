import { useState } from 'react';

export function BookCover({ url, title, size = 'table' }: { url?: string; title?: string; size?: 'table' | 'preview' }) {
  const [failed, setFailed] = useState(false);
  const src = (url || '').trim();
  if (!src || failed) {
    return <div className={`book-cover book-cover-${size} book-cover-empty`}>无封面</div>;
  }
  return (
    <img
      className={`book-cover book-cover-${size}`}
      src={src}
      alt={title ? `${title}封面` : '书籍封面'}
      loading="lazy"
      onError={() => setFailed(true)}
    />
  );
}

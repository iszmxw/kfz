INSERT INTO t_book (
  isbn, title, author, publisher, publish_year, cover_url, source, created_at, updated_at
) VALUES
  ('9787111128069', '示例可回收图书', '示例作者 A', '机械工业出版社', '2018', '', 'mock', NOW(3), NOW(3)),
  ('9787115428028', '示例低价图书', '示例作者 B', '人民邮电出版社', '2016', '', 'mock', NOW(3), NOW(3)),
  ('9787300000001', '示例样本不足图书', '示例作者 C', '示例出版社', '2020', '', 'mock', NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  author = VALUES(author),
  publisher = VALUES(publisher),
  publish_year = VALUES(publish_year),
  cover_url = VALUES(cover_url),
  source = VALUES(source),
  updated_at = NOW(3);

INSERT INTO t_price_snapshot (
  id, isbn, source, min_price, avg_price, max_price,
  sample_count, confidence, raw_url, raw_payload_ref, collected_at, expires_at, created_at
) VALUES
  ('price_9787111128069_mock_001', '9787111128069', 'local_mock', 18.00, 24.50, 39.00, 8, 'HIGH', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3)),
  ('price_9787115428028_mock_001', '9787115428028', 'local_mock', 3.00, 8.00, 12.00, 6, 'MEDIUM', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3)),
  ('price_9787300000001_mock_001', '9787300000001', 'local_mock', 18.00, 30.00, 45.00, 1, 'LOW', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3))
ON DUPLICATE KEY UPDATE
  min_price = VALUES(min_price),
  avg_price = VALUES(avg_price),
  max_price = VALUES(max_price),
  sample_count = VALUES(sample_count),
  confidence = VALUES(confidence),
  collected_at = VALUES(collected_at),
  expires_at = VALUES(expires_at);

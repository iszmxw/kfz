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

INSERT INTO t_recycle_rule (
  id, version, name, min_accept_avg_price, recycle_rate, min_sample_count,
  price_valid_days, low_confidence_mode, enabled, created_by, created_at, updated_at
) VALUES (
  'rule_default_v1', 'default-v1', '默认回收规则', 10.00, 0.3000, 3,
  30, 'NEED_REVIEW', 1, 'seed', NOW(3), NOW(3)
) ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  min_accept_avg_price = VALUES(min_accept_avg_price),
  recycle_rate = VALUES(recycle_rate),
  min_sample_count = VALUES(min_sample_count),
  price_valid_days = VALUES(price_valid_days),
  low_confidence_mode = VALUES(low_confidence_mode),
  enabled = VALUES(enabled),
  updated_at = NOW(3);

INSERT INTO t_admin_role (
  id, code, name, description, status, created_at, updated_at
) VALUES
  ('role_super_admin', 'SUPER_ADMIN', '超级管理员', '拥有全部后台权限', 'ACTIVE', NOW(3), NOW(3)),
  ('role_operator', 'OPERATOR', '运营人员', '处理人工确认、书籍、价格和导入维护', 'ACTIVE', NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description),
  status = VALUES(status),
  updated_at = NOW(3);

INSERT INTO t_admin_menu (
  id, parent_id, title, path, icon, permission_code, sort, status, created_at, updated_at
) VALUES
  ('menu_dashboard', '', '工作台', '/dashboard', 'IconDashboard', 'dashboard:read', 10, 'ACTIVE', NOW(3), NOW(3)),
  ('menu_manual_review', '', '人工确认', '/manual-review', 'IconCheckCircle', 'manual_review:read', 20, 'ACTIVE', NOW(3), NOW(3)),
  ('menu_scan_log', '', '扫码日志', '/scan-log', 'IconHistory', 'scan_log:read', 30, 'ACTIVE', NOW(3), NOW(3)),
  ('menu_book', '', '书籍库', '/book', 'IconBook', 'book:read', 40, 'ACTIVE', NOW(3), NOW(3)),
  ('menu_price_snapshot', '', '价格快照', '/price-snapshot', 'IconStorage', 'price_snapshot:read', 50, 'ACTIVE', NOW(3), NOW(3)),
  ('menu_import', '', '导入任务', '/import', 'IconUpload', 'import:read', 60, 'ACTIVE', NOW(3), NOW(3)),
  ('menu_recycle_rule', '', '规则配置', '/recycle-rule', 'IconSettings', 'recycle_rule:read', 70, 'ACTIVE', NOW(3), NOW(3)),
  ('menu_system', '', '系统权限', '/system', 'IconSafe', 'system:read', 80, 'ACTIVE', NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  path = VALUES(path),
  icon = VALUES(icon),
  sort = VALUES(sort),
  status = VALUES(status),
  updated_at = NOW(3);

INSERT INTO t_admin_role_permission (
  id, role_id, permission_type, permission_code, created_at
) VALUES
  ('rp_super_all', 'role_super_admin', 'API', '*', NOW(3)),
  ('rp_operator_dashboard', 'role_operator', 'API', 'dashboard:read', NOW(3)),
  ('rp_operator_manual_review', 'role_operator', 'API', 'manual_review:*', NOW(3)),
  ('rp_operator_scan_log', 'role_operator', 'API', 'scan_log:read', NOW(3)),
  ('rp_operator_book', 'role_operator', 'API', 'book:*', NOW(3)),
  ('rp_operator_price', 'role_operator', 'API', 'price_snapshot:*', NOW(3)),
  ('rp_operator_import', 'role_operator', 'API', 'import:*', NOW(3))
ON DUPLICATE KEY UPDATE
  permission_code = VALUES(permission_code);

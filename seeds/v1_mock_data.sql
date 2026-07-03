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
  isbn, source, min_price, avg_price, max_price,
  sample_count, confidence, raw_url, raw_payload_ref, collected_at, expires_at, created_at
) VALUES
  ('9787111128069', 'local_mock', 18.00, 24.50, 39.00, 8, 'HIGH', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3)),
  ('9787115428028', 'local_mock', 3.00, 8.00, 12.00, 6, 'MEDIUM', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3)),
  ('9787300000001', 'local_mock', 18.00, 30.00, 45.00, 1, 'LOW', NULL, NULL, NOW(3), DATE_ADD(NOW(3), INTERVAL 30 DAY), NOW(3));

INSERT INTO t_admin_user (
  username, password_hash, name, status, created_at, updated_at
) VALUES (
  'admin', '$2y$10$B8tkEwIu6pkXHPWuc1LfmOJD7wmMky8GAVJMf43oa0s6J/uf9S2mS', '超级管理员', 'ACTIVE', NOW(3), NOW(3)
) ON DUPLICATE KEY UPDATE
  password_hash = VALUES(password_hash),
  name = VALUES(name),
  status = VALUES(status),
  updated_at = NOW(3);

INSERT INTO t_admin_role (
  code, name, description, status, created_at, updated_at
) VALUES
  ('SUPER_ADMIN', '超级管理员', '拥有全部后台权限', 'ACTIVE', NOW(3), NOW(3)),
  ('OPERATOR', '运营人员', '处理人工确认、书籍、价格和导入维护', 'ACTIVE', NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description),
  status = VALUES(status),
  updated_at = NOW(3);

INSERT INTO t_admin_user_role (
  user_id, role_id, created_at
)
SELECT u.id, r.id, NOW(3)
FROM t_admin_user u
JOIN t_admin_role r ON r.code = 'SUPER_ADMIN'
WHERE u.username = 'admin'
ON DUPLICATE KEY UPDATE
  user_id = VALUES(user_id),
  role_id = VALUES(role_id);

INSERT INTO t_recycle_rule (
  version, name, min_accept_avg_price, recycle_rate, min_sample_count,
  price_valid_days, low_confidence_mode, enabled, created_by, created_at, updated_at
)
SELECT
  'default-v1', '默认回收规则', 10.00, 0.3000, 3,
  30, 'NEED_REVIEW', 1, u.id, NOW(3), NOW(3)
FROM t_admin_user u
WHERE u.username = 'admin'
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  min_accept_avg_price = VALUES(min_accept_avg_price),
  recycle_rate = VALUES(recycle_rate),
  min_sample_count = VALUES(min_sample_count),
  price_valid_days = VALUES(price_valid_days),
  low_confidence_mode = VALUES(low_confidence_mode),
  enabled = VALUES(enabled),
  updated_at = NOW(3);

INSERT INTO t_admin_menu (
  parent_id, title, path, icon, permission_code, sort, status, created_at, updated_at
) VALUES
  (0, '工作台', '/dashboard', 'IconDashboard', 'dashboard:read', 10, 'ACTIVE', NOW(3), NOW(3)),
  (0, '人工确认', '/manual-review', 'IconCheckCircle', 'manual_review:read', 20, 'ACTIVE', NOW(3), NOW(3)),
  (0, '扫码日志', '/scan-log', 'IconHistory', 'scan_log:read', 30, 'ACTIVE', NOW(3), NOW(3)),
  (0, '书籍库', '/book', 'IconBook', 'book:read', 40, 'ACTIVE', NOW(3), NOW(3)),
  (0, '价格快照', '/price-snapshot', 'IconStorage', 'price_snapshot:read', 50, 'ACTIVE', NOW(3), NOW(3)),
  (0, '导入任务', '/import', 'IconUpload', 'import:read', 60, 'ACTIVE', NOW(3), NOW(3)),
  (0, '规则配置', '/recycle-rule', 'IconSettings', 'recycle_rule:read', 70, 'ACTIVE', NOW(3), NOW(3)),
  (0, '系统权限', '/system', 'IconSafe', 'system:read', 80, 'ACTIVE', NOW(3), NOW(3))
ON DUPLICATE KEY UPDATE
  title = VALUES(title),
  path = VALUES(path),
  icon = VALUES(icon),
  sort = VALUES(sort),
  status = VALUES(status),
  updated_at = NOW(3);

INSERT INTO t_admin_role_permission (
  role_id, permission_type, permission_code, created_at
)
SELECT r.id, 'API', p.permission_code, NOW(3)
FROM t_admin_role r
JOIN (
  SELECT 'SUPER_ADMIN' AS role_code, '*' AS permission_code
  UNION ALL SELECT 'OPERATOR', 'dashboard:read'
  UNION ALL SELECT 'OPERATOR', 'manual_review:*'
  UNION ALL SELECT 'OPERATOR', 'scan_log:read'
  UNION ALL SELECT 'OPERATOR', 'book:*'
  UNION ALL SELECT 'OPERATOR', 'price_snapshot:*'
  UNION ALL SELECT 'OPERATOR', 'import:*'
) p ON p.role_code = r.code
ON DUPLICATE KEY UPDATE
  permission_code = VALUES(permission_code);

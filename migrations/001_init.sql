CREATE TABLE IF NOT EXISTS t_book (
  id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  isbn varchar(20) NOT NULL COMMENT 'normalized ISBN',
  title text NOT NULL COMMENT 'book title',
  author varchar(255) NULL COMMENT 'author',
  publisher varchar(255) NULL COMMENT 'publisher',
  publish_year varchar(20) NULL COMMENT 'publish year',
  cover_url varchar(1000) NULL COMMENT 'cover URL',
  source varchar(50) NOT NULL DEFAULT 'manual' COMMENT 'data source: manual/mock/external/scan',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_book_isbn (isbn),
  KEY idx_book_title (title(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='book base information';

CREATE TABLE IF NOT EXISTS t_price_snapshot (
  id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  isbn varchar(20) NOT NULL COMMENT 'book.isbn',
  source varchar(50) NOT NULL COMMENT 'price source: local_mock/manual/kongfz/external',
  min_price decimal(20,2) NULL COMMENT 'minimum price',
  avg_price decimal(20,2) NULL COMMENT 'average price',
  max_price decimal(20,2) NULL COMMENT 'maximum price',
  sample_count int NOT NULL DEFAULT 0 COMMENT 'valid sample count',
  confidence varchar(20) NOT NULL DEFAULT 'NONE' COMMENT 'HIGH/MEDIUM/LOW/NONE',
  raw_url varchar(1000) NULL COMMENT 'source URL',
  raw_payload_ref varchar(255) NULL COMMENT 'raw payload reference',
  collected_at datetime(3) NOT NULL COMMENT 'collection time',
  expires_at datetime(3) NULL COMMENT 'expiration time',
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_price_snapshot_isbn_collected_at (isbn, collected_at),
  KEY idx_price_snapshot_isbn_expires_at (isbn, expires_at),
  CONSTRAINT fk_price_snapshot_book
    FOREIGN KEY (isbn) REFERENCES t_book (isbn)
    ON UPDATE CASCADE
    ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='second-hand price snapshot';

CREATE TABLE IF NOT EXISTS t_scan_log (
  id bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'primary key',
  isbn varchar(20) NOT NULL COMMENT 'raw or normalized ISBN',
  normalized_isbn varchar(20) NOT NULL COMMENT 'normalized ISBN',
  batch_id varchar(36) NULL COMMENT 'batch ID for V2',
  store_id varchar(36) NULL COMMENT 'store ID for V2',
  operator_id varchar(36) NULL COMMENT 'operator ID for V2',
  decision varchar(20) NOT NULL COMMENT 'ACCEPT/REJECT/NEED_REVIEW',
  reason varchar(500) NOT NULL COMMENT 'decision reason',
  market_min_price decimal(20,2) NULL COMMENT 'minimum market price at scan time',
  market_avg_price decimal(20,2) NULL COMMENT 'average market price at scan time',
  market_max_price decimal(20,2) NULL COMMENT 'maximum market price at scan time',
  market_sample_count int NULL COMMENT 'market sample count at scan time',
  suggested_recycle_price decimal(20,2) NULL COMMENT 'suggested recycle price',
  confidence varchar(20) NOT NULL DEFAULT 'NONE' COMMENT 'price confidence',
  price_snapshot_id bigint unsigned NULL COMMENT 'price snapshot ID',
  duplicate_recently tinyint(1) NOT NULL DEFAULT 0 COMMENT 'recent duplicate flag',
  duplicate_in_batch tinyint(1) NOT NULL DEFAULT 0 COMMENT 'batch duplicate flag',
  client_request_id varchar(64) NULL COMMENT 'client idempotency request ID',
  scanned_at datetime(3) NOT NULL COMMENT 'scan time',
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_scan_log_isbn_scanned_at (normalized_isbn, scanned_at),
  KEY idx_scan_log_batch_created_at (batch_id, created_at),
  KEY idx_scan_log_decision_scanned_at (decision, scanned_at),
  UNIQUE KEY uk_scan_log_client_request_id (client_request_id),
  CONSTRAINT fk_scan_log_book
    FOREIGN KEY (normalized_isbn) REFERENCES t_book (isbn)
    ON UPDATE CASCADE
    ON DELETE RESTRICT,
  CONSTRAINT fk_scan_log_price_snapshot
    FOREIGN KEY (price_snapshot_id) REFERENCES t_price_snapshot (id)
    ON UPDATE CASCADE
    ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='scan log';

CREATE TABLE IF NOT EXISTS t_admin_user (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  username varchar(100) NOT NULL,
  password_hash varchar(255) NOT NULL,
  name varchar(100) NOT NULL,
  status varchar(20) NOT NULL DEFAULT 'ACTIVE',
  last_login_at datetime(3) NULL,
  last_login_ip varchar(100) NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_user_username (username),
  KEY idx_admin_user_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='admin user';

CREATE TABLE IF NOT EXISTS t_admin_role (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  code varchar(50) NOT NULL,
  name varchar(100) NOT NULL,
  description varchar(500) NULL,
  status varchar(20) NOT NULL DEFAULT 'ACTIVE',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_role_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='admin role';

CREATE TABLE IF NOT EXISTS t_admin_user_role (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint unsigned NOT NULL,
  role_id bigint unsigned NOT NULL,
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_user_role (user_id, role_id),
  KEY idx_admin_user_role_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='admin user role';

CREATE TABLE IF NOT EXISTS t_admin_menu (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  parent_id bigint unsigned NOT NULL DEFAULT 0,
  title varchar(100) NOT NULL,
  path varchar(255) NOT NULL,
  icon varchar(100) NULL,
  permission_code varchar(100) NOT NULL,
  sort int NOT NULL DEFAULT 0,
  status varchar(20) NOT NULL DEFAULT 'ACTIVE',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_menu_permission_code (permission_code),
  KEY idx_admin_menu_parent_sort (parent_id, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='admin menu';

CREATE TABLE IF NOT EXISTS t_admin_api_permission (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  method varchar(20) NOT NULL,
  path varchar(255) NOT NULL,
  permission_code varchar(100) NOT NULL,
  description varchar(255) NULL,
  status varchar(20) NOT NULL DEFAULT 'ACTIVE',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_api_method_path (method, path),
  KEY idx_admin_api_permission_code (permission_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='admin api permission';

CREATE TABLE IF NOT EXISTS t_admin_role_permission (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  role_id bigint unsigned NOT NULL,
  permission_type varchar(20) NOT NULL,
  permission_code varchar(100) NOT NULL,
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_admin_role_permission (role_id, permission_type, permission_code),
  KEY idx_admin_role_permission_code (permission_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='admin role permission';

CREATE TABLE IF NOT EXISTS t_admin_operation_log (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint unsigned NOT NULL,
  username varchar(100) NOT NULL,
  action varchar(100) NOT NULL,
  resource varchar(100) NOT NULL,
  request_brief varchar(1000) NULL,
  result varchar(20) NOT NULL,
  ip varchar(100) NULL,
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_admin_operation_log_user_created_at (user_id, created_at),
  KEY idx_admin_operation_log_resource_created_at (resource, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='admin operation log';

CREATE TABLE IF NOT EXISTS t_manual_decision (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  scan_log_id bigint unsigned NOT NULL,
  isbn varchar(20) NOT NULL,
  admin_user_id bigint unsigned NOT NULL,
  manual_decision varchar(20) NOT NULL,
  actual_recycle_price decimal(20,2) NULL,
  note varchar(1000) NULL,
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_manual_decision_scan_log (scan_log_id),
  KEY idx_manual_decision_isbn_created_at (isbn, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='manual decision';

CREATE TABLE IF NOT EXISTS t_recycle_rule (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  version varchar(50) NOT NULL,
  name varchar(100) NOT NULL,
  min_accept_avg_price decimal(20,2) NOT NULL,
  recycle_rate decimal(10,4) NOT NULL,
  min_sample_count int NOT NULL,
  price_valid_days int NOT NULL DEFAULT 30,
  low_confidence_mode varchar(20) NOT NULL DEFAULT 'NEED_REVIEW',
  enabled tinyint(1) NOT NULL DEFAULT 0,
  created_by bigint unsigned NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_recycle_rule_version (version),
  KEY idx_recycle_rule_enabled_updated_at (enabled, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='recycle rule';

CREATE TABLE IF NOT EXISTS t_import_task (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  file_name varchar(255) NOT NULL,
  file_type varchar(20) NOT NULL,
  status varchar(20) NOT NULL,
  total_rows int NOT NULL DEFAULT 0,
  success_rows int NOT NULL DEFAULT 0,
  failed_rows int NOT NULL DEFAULT 0,
  created_by bigint unsigned NOT NULL,
  error_message varchar(1000) NULL,
  created_at datetime(3) NOT NULL,
  completed_at datetime(3) NULL,
  PRIMARY KEY (id),
  KEY idx_import_task_created_at (created_at),
  KEY idx_import_task_status_created_at (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='import task';

CREATE TABLE IF NOT EXISTS t_import_task_row (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  task_id bigint unsigned NOT NULL,
  row_no int NOT NULL,
  isbn varchar(20) NULL,
  title text NULL,
  valid tinyint(1) NOT NULL DEFAULT 0,
  error_message varchar(1000) NULL,
  raw_payload json NULL,
  created_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_import_task_row_task_row_no (task_id, row_no),
  KEY idx_import_task_row_valid (task_id, valid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='import task row';

CREATE TABLE IF NOT EXISTS t_kongfz_collect_task (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  cat_id int NOT NULL,
  start_page int NOT NULL DEFAULT 1,
  end_page int NOT NULL DEFAULT 0,
  delay_ms int NOT NULL DEFAULT 1200,
  retry_times int NOT NULL DEFAULT 2,
  status varchar(20) NOT NULL DEFAULT 'PENDING',
  current_page int NOT NULL DEFAULT 0,
  total_pages int NOT NULL DEFAULT 0,
  collected_count int NOT NULL DEFAULT 0,
  valid_count int NOT NULL DEFAULT 0,
  invalid_count int NOT NULL DEFAULT 0,
  imported_count int NOT NULL DEFAULT 0,
  error_message varchar(1000) NULL,
  request_headers json NULL,
  created_by bigint unsigned NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  started_at datetime(3) NULL,
  completed_at datetime(3) NULL,
  PRIMARY KEY (id),
  KEY idx_kongfz_collect_task_status_created_at (status, created_at),
  KEY idx_kongfz_collect_task_cat_created_at (cat_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='kongfz category collect task';

CREATE TABLE IF NOT EXISTS t_kongfz_collect_raw_row (
  id bigint unsigned NOT NULL AUTO_INCREMENT,
  task_id bigint unsigned NOT NULL,
  cat_id int NOT NULL,
  page int NOT NULL,
  row_index int NOT NULL,
  isbn varchar(20) NULL,
  normalized_isbn varchar(20) NULL,
  title text NULL,
  cover_url varchar(1000) NULL,
  valid tinyint(1) NOT NULL DEFAULT 0,
  error_message varchar(1000) NULL,
  raw_hash varchar(64) NOT NULL,
  raw_payload json NOT NULL,
  import_task_row_id bigint unsigned NULL,
  sync_status varchar(20) NOT NULL DEFAULT 'PENDING',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_kongfz_collect_raw_row_task_page_index (task_id, page, row_index),
  KEY idx_kongfz_collect_raw_row_cat_isbn (cat_id, normalized_isbn),
  KEY idx_kongfz_collect_raw_row_task_status (task_id, sync_status),
  KEY idx_kongfz_collect_raw_row_raw_hash (raw_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='kongfz collected raw row';

CREATE TABLE IF NOT EXISTS t_book (
  isbn varchar(20) NOT NULL COMMENT 'normalized ISBN primary key',
  title varchar(255) NOT NULL COMMENT 'book title',
  author varchar(255) NULL COMMENT 'author',
  publisher varchar(255) NULL COMMENT 'publisher',
  publish_year varchar(20) NULL COMMENT 'publish year',
  cover_url varchar(1000) NULL COMMENT 'cover URL',
  source varchar(50) NOT NULL DEFAULT 'manual' COMMENT 'data source: manual/mock/external/scan',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (isbn),
  KEY idx_book_title (title)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='book base information';

CREATE TABLE IF NOT EXISTS t_price_snapshot (
  id varchar(36) NOT NULL COMMENT 'primary key',
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
  id varchar(36) NOT NULL COMMENT 'primary key',
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
  price_snapshot_id varchar(36) NULL COMMENT 'price snapshot ID',
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

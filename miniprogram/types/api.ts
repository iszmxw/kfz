export type Decision = "ACCEPT" | "REJECT" | "NEED_REVIEW";

export type Confidence = "HIGH" | "MEDIUM" | "LOW" | "NONE" | string;

export interface EchoResponse<T> {
  code: number;
  data?: T;
  msg: string;
  reqId?: string;
}

export interface BookDTO {
  isbn: string;
  normalized_isbn: string;
  title: string;
  author: string;
  publisher: string;
  publish_year: string;
  cover_url: string;
}

export interface MarketPriceDTO {
  source: string;
  min: number | null;
  avg: number | null;
  max: number | null;
  sample_count: number;
  confidence: Confidence;
  collected_at: string | null;
}

export interface DuplicateInfoDTO {
  scanned_recently: boolean;
  duplicate_in_current_batch: boolean;
  last_scanned_at: string | null;
}

export interface BookCheckResponse {
  scan_log_id: string;
  book: BookDTO;
  decision: Decision;
  reason: string;
  suggested_recycle_price: number | null;
  market_price: MarketPriceDTO | null;
  duplicate: DuplicateInfoDTO;
  batch_action_required?: string;
  scanned_at: string;
}

export interface LatestDecisionDTO {
  decision: Decision;
  reason: string;
  suggested_recycle_price: number | null;
  updated_at: string;
}

export interface BookDetailResponse {
  book: BookDTO;
  latest_market_price: MarketPriceDTO | null;
  latest_decision: LatestDecisionDTO;
}

export interface ScanHistoryItem {
  scan_log_id: string;
  isbn: string;
  title: string;
  decision: Decision;
  reason: string;
  suggested_recycle_price: number | null;
  confidence: Confidence;
  scanned_at: string;
}

export interface ScanHistoryResponse {
  page: number;
  page_size: number;
  total: number;
  items: ScanHistoryItem[];
}

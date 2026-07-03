export interface AdminMenu {
  id: string;
  title: string;
  path: string;
  icon: string;
  permission_code: string;
}

export interface AdminUser {
  id: string;
  username: string;
  name: string;
  status: string;
  roles: string[];
}

export interface LoginResponse {
  token: string;
  user: AdminUser;
  permissions: string[];
  menus: AdminMenu[];
}

export interface PageResponse<T> {
  page: number;
  page_size: number;
  total: number;
  items: T[];
}

export interface Book {
  isbn: string;
  title: string;
  author: string;
  publisher: string;
  publishYear?: string;
  publish_year?: string;
  coverUrl?: string;
  cover_url?: string;
  source: string;
  updatedAt?: string;
  updated_at?: string;
}

export interface ScanLog {
  id: string;
  normalizedIsbn?: string;
  normalized_isbn?: string;
  decision: string;
  reason: string;
  confidence: string;
  suggestedRecyclePrice?: string;
  suggested_recycle_price?: string;
  scannedAt?: string;
  scanned_at?: string;
}

export interface PriceSnapshot {
  id: string;
  isbn: string;
  source: string;
  minPrice?: string;
  avgPrice?: string;
  maxPrice?: string;
  sampleCount?: number;
  confidence: string;
  collectedAt?: string;
}

export interface AdminMenu {
  id: number;
  title: string;
  path: string;
  icon: string;
  permission_code: string;
}

export interface AdminUser {
  id: number;
  username: string;
  name: string;
  status: string;
  roles: string[];
}

export interface AdminRole {
  id: number;
  code: string;
  name: string;
  description?: string;
  status: string;
}

export interface AdminUserListItem {
  id: number;
  username: string;
  name: string;
  status: string;
  roleIds?: number[];
  roles?: string[];
  lastLoginAt?: string | null;
  lastLoginIp?: string;
  createdAt?: string;
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
  id: number;
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
  id: number;
  isbn: string;
  source: string;
  minPrice?: string;
  avgPrice?: string;
  maxPrice?: string;
  sampleCount?: number;
  confidence: string;
  collectedAt?: string;
}

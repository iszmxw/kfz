export type CollectorRunStatus = "idle" | "running" | "paused" | "stopping" | "completed" | "failed" | "stopped"

export type SyncStatus = "pending" | "synced" | "failed"

export interface CollectorConfig {
  backendBaseUrl: string
  catId: number
  startPage: number
  endPage?: number
  delayMs: number
  retryTimes: number
}

export interface CollectorRun {
  id: string
  catId: number
  startPage: number
  endPage?: number
  delayMs: number
  status: CollectorRunStatus
  currentPage: number
  totalPages?: number
  collectedCount: number
  validCount: number
  invalidCount: number
  message: string
  startedAt: string
  completedAt?: string
  updatedAt: string
}

export interface CollectorStatus {
  runId?: string
  running: boolean
  paused: boolean
  status: CollectorRunStatus
  currentPage: number
  totalPages?: number
  collectedCount: number
  validCount: number
  invalidCount: number
  message: string
}

export interface CollectorItem {
  id: string
  runId: string
  catId: number
  page: number
  index: number
  valid: boolean
  errorMessage: string
  isbn: string
  normalizedIsbn: string
  title: string
  author: string
  publisher: string
  publishYear: string
  publishDate: string
  binding: string
  listPrice: string
  coverUrl: string
  kongfzId?: number
  mid?: number
  oldBookMinPrice?: string
  oldBookMinPriceText: string
  oldBookOnSaleNum: number
  newBookMinPrice?: string
  newBookMinPriceText: string
  newBookOnSaleNum: number
  onSaleProductNum: number
  score?: number
  saleNum?: number
  popularity?: number
  bookShowInfo: string[]
  rawPayload: unknown
  rawHash: string
  rawUrl: string
  syncStatus: SyncStatus
  syncError?: string
  syncedAt?: string
  createdAt: string
  updatedAt: string
}

export interface SyncBatch {
  id: string
  backendBaseUrl: string
  catId: number
  total: number
  success: number
  failed: number
  responsePayload?: unknown
  errorMessage?: string
  createdAt: string
}

export interface CollectorSummary {
  total: number
  valid: number
  invalid: number
  pending: number
  synced: number
  failed: number
}

export interface KongfzApiPager {
  page?: number
  size?: number
  total?: number
  pages?: number
}

export interface KongfzApiResponse {
  status: number
  errType?: string
  message?: string
  data?: {
    itemResponse?: {
      total?: number
      list?: unknown[]
      pager?: KongfzApiPager
    }
  }
}

export interface RuntimeMessage<T = unknown> {
  type: string
  payload?: T
}

export interface RuntimeResponse<T = unknown> {
  ok: boolean
  data?: T
  error?: string
}

export interface StartCollectionPayload {
  config: CollectorConfig
}

export interface SyncItemsPayload {
  backendBaseUrl: string
  token: string
  catId: number
  limit?: number
}

export const DEFAULT_COLLECTOR_CONFIG: CollectorConfig = {
  backendBaseUrl: "http://127.0.0.1:8888",
  catId: 43,
  startPage: 1,
  delayMs: 1200,
  retryTimes: 2
}

export const MESSAGE_TYPES = {
  GET_STATUS: "KONGFZ_GET_STATUS",
  START_COLLECTION: "KONGFZ_START_COLLECTION",
  PAUSE_COLLECTION: "KONGFZ_PAUSE_COLLECTION",
  RESUME_COLLECTION: "KONGFZ_RESUME_COLLECTION",
  STOP_COLLECTION: "KONGFZ_STOP_COLLECTION",
  SYNC_ITEMS: "KONGFZ_SYNC_ITEMS"
} as const

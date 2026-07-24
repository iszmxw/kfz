import { getUnsyncedValidItems, markItemsSyncStatus, putSyncBatch } from "./db"
import type { CollectorItem, SyncBatch } from "./types"

interface EchoResponse<T> {
  code: number
  msg: string
  data?: T
}

interface BackendSyncRow {
  isbn: string
  valid: boolean
  error_message?: string
}

interface BackendSyncData {
  summary: {
    total: number
    success: number
    failed: number
  }
  rows: BackendSyncRow[]
}

export async function loginAdmin(backendBaseUrl: string, username: string, password: string): Promise<string> {
  const response = await fetch(`${normalizeBackendBaseUrl(backendBaseUrl)}/admin/api/v1/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ username, password })
  })
  const payload = (await response.json()) as EchoResponse<{ token: string }>
  if (!response.ok || payload.code !== 1 || !payload.data?.token) {
    throw new Error(payload.msg || "后台登录失败")
  }
  return payload.data.token
}

export async function syncUnsyncedItems(params: {
  backendBaseUrl: string
  token: string
  catId: number
  limit?: number
}): Promise<BackendSyncData> {
  const items = await getUnsyncedValidItems(params.limit ?? 500)
  if (items.length === 0) {
    return { summary: { total: 0, success: 0, failed: 0 }, rows: [] }
  }

  const now = new Date().toISOString()
  const batch: SyncBatch = {
    id: crypto.randomUUID(),
    backendBaseUrl: params.backendBaseUrl,
    catId: params.catId,
    total: items.length,
    success: 0,
    failed: 0,
    createdAt: now
  }

  try {
    const response = await fetch(`${normalizeBackendBaseUrl(params.backendBaseUrl)}/admin/api/v1/import/kongfz-category/sync`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${params.token}`,
        "Content-Type": "application/json"
      },
      body: JSON.stringify({
        cat_id: params.catId,
        source: "kongfz_category",
        items: items.map(toBackendItem)
      })
    })
    const payload = (await response.json()) as EchoResponse<BackendSyncData>
    if (!response.ok || payload.code !== 1 || !payload.data) {
      throw new Error(payload.msg || "同步失败")
    }

    const succeeded = new Set(payload.data.rows.filter((row) => row.valid).map((row) => row.isbn))
    const failed = new Map(payload.data.rows.filter((row) => !row.valid).map((row) => [row.isbn, row.error_message || "同步失败"]))
    await markItemsSyncStatus(
      items.filter((item) => succeeded.has(item.normalizedIsbn)).map((item) => item.id),
      "synced"
    )
    for (const [isbn, message] of failed) {
      await markItemsSyncStatus(
        items.filter((item) => item.normalizedIsbn === isbn).map((item) => item.id),
        "failed",
        message
      )
    }

    batch.success = payload.data.summary.success
    batch.failed = payload.data.summary.failed
    batch.responsePayload = payload.data
    await putSyncBatch(batch)
    return payload.data
  } catch (error) {
    const message = error instanceof Error ? error.message : "同步失败"
    await markItemsSyncStatus(
      items.map((item) => item.id),
      "failed",
      message
    )
    batch.failed = items.length
    batch.errorMessage = message
    await putSyncBatch(batch)
    throw error
  }
}

function toBackendItem(item: CollectorItem): Record<string, unknown> {
  return {
    isbn: item.normalizedIsbn,
    title: item.title,
    book_name: item.title,
    author: item.author,
    publisher: item.publisher,
    publish_year: item.publishYear,
    publish_date: item.publishDate,
    binding: item.binding,
    list_price: item.listPrice,
    cover_url: item.coverUrl,
    raw_url: item.rawUrl,
    page: item.page,
    mid: item.mid ?? 0,
    kongfz_id: item.kongfzId ?? 0,
    old_book_min_price: item.oldBookMinPrice ?? "",
    old_book_min_price_text: item.oldBookMinPriceText,
    old_book_on_sale_num: item.oldBookOnSaleNum,
    book_show_info: item.bookShowInfo,
    raw_payload: item.rawPayload
  }
}

function normalizeBackendBaseUrl(value: string): string {
  return value.replace(/\/+$/, "")
}

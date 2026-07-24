import { sha256, stableStringify } from "./hash"
import { normalizeISBN } from "./isbn"
import type { CollectorItem, KongfzApiPager, KongfzApiResponse } from "./types"

export interface MapKongfzItemContext {
  runId: string
  catId: number
  page: number
  index: number
  rawUrl: string
}

export interface ValidatedKongfzPage {
  list: unknown[]
  pager: KongfzApiPager
}

export function validateKongfzResponse(payload: KongfzApiResponse): ValidatedKongfzPage {
  if (!payload || payload.status !== 1) {
    throw new Error(payload?.message || "孔夫子接口返回失败")
  }
  const itemResponse = payload.data?.itemResponse
  if (!itemResponse || !Array.isArray(itemResponse.list)) {
    throw new Error("孔夫子接口数据结构异常")
  }
  return {
    list: itemResponse.list,
    pager: itemResponse.pager ?? {}
  }
}

export async function mapKongfzItem(raw: unknown, context: MapKongfzItemContext): Promise<CollectorItem> {
  const record = asRecord(raw)
  const bookShowInfo = arrayOfString(record.bookShowInfo)
  const title = stringValue(record.bookName)
  const normalizedIsbn = normalizeISBN(record.isbn)
  const publishDate = bookShowInfo[2] ?? ""
  const now = new Date().toISOString()
  const errors: string[] = []
  if (!normalizedIsbn) {
    errors.push("ISBN格式无效")
  }
  if (!title) {
    errors.push("书名不能为空")
  }

  const rawString = stableStringify(raw)
  const id = normalizedIsbn || `invalid:${context.runId}:${context.page}:${context.index}`
  return {
    id,
    runId: context.runId,
    catId: context.catId,
    page: context.page,
    index: context.index,
    valid: errors.length === 0,
    errorMessage: errors.join("；"),
    isbn: stringValue(record.isbn),
    normalizedIsbn,
    title,
    author: bookShowInfo[0] ?? "",
    publisher: bookShowInfo[1] ?? "",
    publishYear: extractPublishYear(publishDate),
    publishDate,
    binding: bookShowInfo[3] ?? "",
    listPrice: bookShowInfo[4] ?? "",
    coverUrl: imageUrl(record.imgUrlEntity),
    kongfzId: numberValue(record.id),
    mid: numberValue(record.mid),
    oldBookMinPrice: decimalString(record.oldBookMinPrice),
    oldBookMinPriceText: stringValue(record.oldBookMinPriceText),
    oldBookOnSaleNum: numberValue(record.oldBookOnSaleNum) ?? 0,
    newBookMinPrice: decimalString(record.newBookMinPrice),
    newBookMinPriceText: stringValue(record.newBookMinPriceText),
    newBookOnSaleNum: numberValue(record.newBookOnSaleNum) ?? 0,
    onSaleProductNum: numberValue(record.onSaleProductNum) ?? 0,
    score: numberValue(record.score),
    saleNum: numberValue(record.saleNum),
    popularity: numberValue(record.popularity),
    bookShowInfo,
    rawPayload: raw,
    rawHash: await sha256(rawString),
    rawUrl: context.rawUrl,
    syncStatus: "pending",
    createdAt: now,
    updatedAt: now
  }
}

export function buildKongfzCategoryUrl(catId: number, page: number): string {
  const url = new URL("https://search.kongfz.com/pc-gw/search-web/client/pc/bookLib/category/list")
  url.searchParams.set("catId", String(catId))
  url.searchParams.set("page", String(page))
  return url.toString()
}

function asRecord(value: unknown): Record<string, unknown> {
  if (value && typeof value === "object") {
    return value as Record<string, unknown>
  }
  return {}
}

function arrayOfString(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return []
  }
  return value.map((entry) => String(entry ?? "").trim())
}

function stringValue(value: unknown): string {
  return String(value ?? "").trim()
}

function numberValue(value: unknown): number | undefined {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value
  }
  const parsed = Number(String(value ?? "").trim())
  if (!Number.isFinite(parsed)) {
    return undefined
  }
  return parsed
}

function decimalString(value: unknown): string | undefined {
  if (value === null || value === undefined || value === "") {
    return undefined
  }
  if (typeof value === "number" && Number.isFinite(value)) {
    return String(value)
  }
  const text = String(value).trim()
  return text || undefined
}

function imageUrl(value: unknown): string {
  const entity = asRecord(value)
  return stringValue(entity.bigImgUrl) || stringValue(entity.smallImgUrl)
}

function extractPublishYear(value: string): string {
  const text = value.trim()
  if (text.length < 4) {
    return text
  }
  const year = text.slice(0, 4)
  return /^\d{4}$/.test(year) ? year : text
}

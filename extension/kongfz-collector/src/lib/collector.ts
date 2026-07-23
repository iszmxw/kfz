import { buildKongfzCategoryUrl, mapKongfzItem, validateKongfzResponse } from "./kongfzMapper"
import type { CollectorConfig, CollectorItem, CollectorStatus, KongfzApiResponse } from "./types"

export interface CollectPagesDeps {
  runId: string
  fetchPage: (catId: number, page: number) => Promise<KongfzApiResponse>
  saveItems: (items: CollectorItem[]) => Promise<void>
  onProgress?: (status: CollectorStatus) => Promise<void> | void
  shouldPause?: () => boolean
  shouldStop?: () => boolean
  delay?: (ms: number) => Promise<void>
}

export class CollectorStopError extends Error {}

export async function collectPages(config: CollectorConfig, deps: CollectPagesDeps): Promise<CollectorStatus> {
  const delay = deps.delay ?? sleep
  const startPage = Math.max(1, config.startPage)
  let endPage = config.endPage && config.endPage >= startPage ? config.endPage : undefined
  let totalPages = endPage
  let collectedCount = 0
  let validCount = 0
  let invalidCount = 0
  let currentPage = startPage

  const emit = async (patch: Partial<CollectorStatus> = {}) => {
    await deps.onProgress?.({
      runId: deps.runId,
      running: true,
      paused: Boolean(deps.shouldPause?.()),
      status: Boolean(deps.shouldPause?.()) ? "paused" : "running",
      currentPage,
      totalPages,
      collectedCount,
      validCount,
      invalidCount,
      message: "",
      ...patch
    })
  }

  for (currentPage = startPage; ; currentPage += 1) {
    if (deps.shouldStop?.()) {
      await emit({ running: false, status: "stopped", message: "已停止" })
      return buildFinalStatus(deps.runId, "stopped", currentPage, totalPages, collectedCount, validCount, invalidCount, "已停止")
    }
    while (deps.shouldPause?.()) {
      await emit({ status: "paused", message: "已暂停" })
      await delay(500)
      if (deps.shouldStop?.()) {
        await emit({ running: false, status: "stopped", message: "已停止" })
        return buildFinalStatus(deps.runId, "stopped", currentPage, totalPages, collectedCount, validCount, invalidCount, "已停止")
      }
    }

    await emit({ message: `采集第 ${currentPage} 页` })
    const payload = await fetchPageWithRetry(config.catId, currentPage, config.retryTimes, deps.fetchPage, delay)
    const page = validateKongfzResponse(payload)
    if (!endPage) {
      const inferredPages = page.pager.pages && page.pager.pages > 0 ? page.pager.pages : currentPage
      totalPages = Math.min(inferredPages, 100)
      endPage = totalPages
    }
    if (page.list.length === 0) {
      throw new CollectorStopError(`第 ${currentPage} 页为空，采集已停止`)
    }

    const rawUrl = buildKongfzCategoryUrl(config.catId, currentPage)
    const mapped = await Promise.all(
      page.list.map((raw, index) =>
        mapKongfzItem(raw, {
          runId: deps.runId,
          catId: config.catId,
          page: currentPage,
          index,
          rawUrl
        })
      )
    )
    await deps.saveItems(mapped)
    collectedCount += mapped.length
    validCount += mapped.filter((item) => item.valid).length
    invalidCount += mapped.filter((item) => !item.valid).length
    await emit({ message: `第 ${currentPage} 页完成` })

    if (!endPage || currentPage >= endPage) {
      break
    }
    await delay(Math.max(0, config.delayMs))
  }

  const message = `采集完成，共 ${collectedCount} 条`
  await emit({ running: false, paused: false, status: "completed", message })
  return buildFinalStatus(deps.runId, "completed", currentPage, totalPages, collectedCount, validCount, invalidCount, message)
}

export async function fetchKongfzPage(catId: number, page: number): Promise<KongfzApiResponse> {
  const response = await fetch(buildKongfzCategoryUrl(catId, page), {
    credentials: "omit",
    headers: {
      Accept: "application/json,text/plain,*/*"
    }
  })
  if ([401, 403, 429].includes(response.status)) {
    throw new CollectorStopError(`孔夫子接口返回 ${response.status}，采集已停止`)
  }
  if (!response.ok) {
    throw new Error(`孔夫子接口返回 ${response.status}`)
  }
  const text = await response.text()
  if (looksBlocked(text)) {
    throw new CollectorStopError("孔夫子返回验证或风控页面，采集已停止")
  }
  try {
    return JSON.parse(text) as KongfzApiResponse
  } catch {
    throw new CollectorStopError("孔夫子接口返回非JSON内容，采集已停止")
  }
}

async function fetchPageWithRetry(
  catId: number,
  page: number,
  retryTimes: number,
  fetchPage: CollectPagesDeps["fetchPage"],
  delay: (ms: number) => Promise<void>
): Promise<KongfzApiResponse> {
  let lastError: unknown
  for (let attempt = 0; attempt <= retryTimes; attempt += 1) {
    try {
      return await fetchPage(catId, page)
    } catch (error) {
      if (error instanceof CollectorStopError) {
        throw error
      }
      lastError = error
      if (attempt < retryTimes) {
        await delay(500 * (attempt + 1))
      }
    }
  }
  throw lastError instanceof Error ? lastError : new Error("孔夫子接口请求失败")
}

function looksBlocked(text: string): boolean {
  const lower = text.toLowerCase()
  return lower.includes("captcha") || text.includes("验证码") || text.includes("访问过于频繁")
}

function buildFinalStatus(
  runId: string,
  status: CollectorStatus["status"],
  currentPage: number,
  totalPages: number | undefined,
  collectedCount: number,
  validCount: number,
  invalidCount: number,
  message: string
): CollectorStatus {
  return {
    runId,
    running: false,
    paused: false,
    status,
    currentPage,
    totalPages,
    collectedCount,
    validCount,
    invalidCount,
    message
  }
}

export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

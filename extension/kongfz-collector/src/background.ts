import { collectPages, CollectorStopError, fetchKongfzPage } from "./lib/collector"
import { getCollectorSummary, getLatestRun, patchRun, putCollectorItems, putRun } from "./lib/db"
import { syncUnsyncedItems } from "./lib/syncClient"
import {
  DEFAULT_COLLECTOR_CONFIG,
  MESSAGE_TYPES,
  type CollectorConfig,
  type CollectorRun,
  type CollectorStatus,
  type RuntimeMessage,
  type RuntimeResponse,
  type StartCollectionPayload,
  type SyncItemsPayload
} from "./lib/types"

let active = false
let paused = false
let stopping = false
let runtimeStatus: CollectorStatus = {
  running: false,
  paused: false,
  status: "idle",
  currentPage: 0,
  collectedCount: 0,
  validCount: 0,
  invalidCount: 0,
  message: ""
}

chrome.runtime.onMessage.addListener((message: RuntimeMessage, _sender, sendResponse) => {
  void handleMessage(message)
    .then((data) => sendResponse({ ok: true, data } satisfies RuntimeResponse))
    .catch((error) => sendResponse({ ok: false, error: error instanceof Error ? error.message : "操作失败" } satisfies RuntimeResponse))
  return true
})

async function handleMessage(message: RuntimeMessage): Promise<unknown> {
  switch (message.type) {
    case MESSAGE_TYPES.GET_STATUS:
      return getRuntimeStatus()
    case MESSAGE_TYPES.START_COLLECTION:
      void startCollection((message.payload as StartCollectionPayload | undefined)?.config)
      return getRuntimeStatus()
    case MESSAGE_TYPES.PAUSE_COLLECTION:
      paused = true
      runtimeStatus = { ...runtimeStatus, paused: true, status: "paused", message: "已暂停" }
      return getRuntimeStatus()
    case MESSAGE_TYPES.RESUME_COLLECTION:
      paused = false
      runtimeStatus = { ...runtimeStatus, paused: false, status: active ? "running" : runtimeStatus.status, message: active ? "继续采集" : runtimeStatus.message }
      return getRuntimeStatus()
    case MESSAGE_TYPES.STOP_COLLECTION:
      stopping = true
      runtimeStatus = { ...runtimeStatus, status: "stopping", message: "正在停止" }
      return getRuntimeStatus()
    case MESSAGE_TYPES.SYNC_ITEMS:
      return syncUnsyncedItems(message.payload as SyncItemsPayload)
    default:
      throw new Error("未知操作")
  }
}

async function startCollection(config?: CollectorConfig): Promise<void> {
  if (active) {
    runtimeStatus = { ...runtimeStatus, message: "已有采集任务运行中" }
    return
  }

  const normalizedConfig = {
    ...DEFAULT_COLLECTOR_CONFIG,
    ...config,
    catId: positiveInt(config?.catId, DEFAULT_COLLECTOR_CONFIG.catId),
    startPage: positiveInt(config?.startPage, DEFAULT_COLLECTOR_CONFIG.startPage),
    delayMs: positiveInt(config?.delayMs, DEFAULT_COLLECTOR_CONFIG.delayMs),
    retryTimes: Math.max(0, Number(config?.retryTimes ?? DEFAULT_COLLECTOR_CONFIG.retryTimes))
  }
  if (config?.endPage && config.endPage >= normalizedConfig.startPage) {
    normalizedConfig.endPage = Math.floor(config.endPage)
  } else {
    delete normalizedConfig.endPage
  }

  const now = new Date().toISOString()
  const run: CollectorRun = {
    id: crypto.randomUUID(),
    catId: normalizedConfig.catId,
    startPage: normalizedConfig.startPage,
    endPage: normalizedConfig.endPage,
    delayMs: normalizedConfig.delayMs,
    status: "running",
    currentPage: normalizedConfig.startPage,
    totalPages: normalizedConfig.endPage,
    collectedCount: 0,
    validCount: 0,
    invalidCount: 0,
    message: "开始采集",
    startedAt: now,
    updatedAt: now
  }

  active = true
  paused = false
  stopping = false
  runtimeStatus = runToStatus(run)
  await putRun(run)

  try {
    const finalStatus = await collectPages(normalizedConfig, {
      runId: run.id,
      fetchPage: fetchKongfzPage,
      saveItems: putCollectorItems,
      shouldPause: () => paused,
      shouldStop: () => stopping,
      onProgress: async (status) => {
        runtimeStatus = status
        await patchRun(run.id, {
          status: status.status,
          currentPage: status.currentPage,
          totalPages: status.totalPages,
          collectedCount: status.collectedCount,
          validCount: status.validCount,
          invalidCount: status.invalidCount,
          message: status.message
        })
      }
    })
    runtimeStatus = finalStatus
    await patchRun(run.id, {
      status: finalStatus.status,
      currentPage: finalStatus.currentPage,
      totalPages: finalStatus.totalPages,
      collectedCount: finalStatus.collectedCount,
      validCount: finalStatus.validCount,
      invalidCount: finalStatus.invalidCount,
      message: finalStatus.message,
      completedAt: new Date().toISOString()
    })
  } catch (error) {
    const message = error instanceof Error ? error.message : "采集失败"
    const status = error instanceof CollectorStopError || stopping ? "stopped" : "failed"
    runtimeStatus = { ...runtimeStatus, running: false, paused: false, status, message }
    await patchRun(run.id, {
      status,
      message,
      completedAt: new Date().toISOString()
    })
  } finally {
    active = false
    paused = false
    stopping = false
  }
}

async function getRuntimeStatus(): Promise<{ status: CollectorStatus; summary: Awaited<ReturnType<typeof getCollectorSummary>> }> {
  if (!active && runtimeStatus.status === "idle") {
    const latest = await getLatestRun()
    if (latest) {
      runtimeStatus = runToStatus(latest)
    }
  }
  return {
    status: runtimeStatus,
    summary: await getCollectorSummary()
  }
}

function runToStatus(run: CollectorRun): CollectorStatus {
  return {
    runId: run.id,
    running: run.status === "running" || run.status === "paused",
    paused: run.status === "paused",
    status: run.status,
    currentPage: run.currentPage,
    totalPages: run.totalPages,
    collectedCount: run.collectedCount,
    validCount: run.validCount,
    invalidCount: run.invalidCount,
    message: run.message
  }
}

function positiveInt(value: unknown, fallback: number): number {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback
  }
  return Math.floor(parsed)
}

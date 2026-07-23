import { deleteDB, openDB, type DBSchema, type IDBPDatabase } from "idb"
import type { CollectorItem, CollectorRun, CollectorSummary, SyncBatch, SyncStatus } from "./types"

const DB_NAME = "kongfz_collector"
const DB_VERSION = 1

interface KongfzCollectorDB extends DBSchema {
  collector_runs: {
    key: string
    value: CollectorRun
    indexes: {
      "by-updated-at": string
    }
  }
  collector_items: {
    key: string
    value: CollectorItem
    indexes: {
      "by-run-id": string
      "by-sync-status": SyncStatus
      "by-cat-id": number
    }
  }
  sync_batches: {
    key: string
    value: SyncBatch
    indexes: {
      "by-created-at": string
    }
  }
}

let dbPromise: Promise<IDBPDatabase<KongfzCollectorDB>> | undefined

export function getDB(): Promise<IDBPDatabase<KongfzCollectorDB>> {
  if (!dbPromise) {
    dbPromise = openDB<KongfzCollectorDB>(DB_NAME, DB_VERSION, {
      upgrade(db) {
        if (!db.objectStoreNames.contains("collector_runs")) {
          const runs = db.createObjectStore("collector_runs", { keyPath: "id" })
          runs.createIndex("by-updated-at", "updatedAt")
        }
        if (!db.objectStoreNames.contains("collector_items")) {
          const items = db.createObjectStore("collector_items", { keyPath: "id" })
          items.createIndex("by-run-id", "runId")
          items.createIndex("by-sync-status", "syncStatus")
          items.createIndex("by-cat-id", "catId")
        }
        if (!db.objectStoreNames.contains("sync_batches")) {
          const batches = db.createObjectStore("sync_batches", { keyPath: "id" })
          batches.createIndex("by-created-at", "createdAt")
        }
      }
    })
  }
  return dbPromise
}

export async function putRun(run: CollectorRun): Promise<void> {
  const db = await getDB()
  await db.put("collector_runs", run)
}

export async function patchRun(id: string, patch: Partial<CollectorRun>): Promise<CollectorRun | undefined> {
  const db = await getDB()
  const run = await db.get("collector_runs", id)
  if (!run) {
    return undefined
  }
  const next = { ...run, ...patch, updatedAt: new Date().toISOString() }
  await db.put("collector_runs", next)
  return next
}

export async function getLatestRun(): Promise<CollectorRun | undefined> {
  const db = await getDB()
  const runs = await db.getAll("collector_runs")
  return runs.sort((left, right) => right.updatedAt.localeCompare(left.updatedAt))[0]
}

export async function putCollectorItems(items: CollectorItem[]): Promise<void> {
  if (items.length === 0) {
    return
  }
  const db = await getDB()
  const tx = db.transaction("collector_items", "readwrite")
  for (const item of items) {
    const existing = await tx.store.get(item.id)
    if (existing) {
      const unchanged = existing.rawHash === item.rawHash
      await tx.store.put({
        ...existing,
        ...item,
        createdAt: existing.createdAt,
        syncStatus: unchanged ? existing.syncStatus : "pending",
        syncError: unchanged ? existing.syncError : undefined,
        syncedAt: unchanged ? existing.syncedAt : undefined,
        updatedAt: new Date().toISOString()
      })
      continue
    }
    await tx.store.put(item)
  }
  await tx.done
}

export async function listCollectorItems(limit = 200): Promise<CollectorItem[]> {
  const db = await getDB()
  const items = await db.getAll("collector_items")
  return items
    .sort((left, right) => right.updatedAt.localeCompare(left.updatedAt))
    .slice(0, limit)
}

export async function getUnsyncedValidItems(limit = 500): Promise<CollectorItem[]> {
  const db = await getDB()
  const items = await db.getAll("collector_items")
  return items
    .filter((item) => item.valid && item.syncStatus !== "synced")
    .sort((left, right) => left.updatedAt.localeCompare(right.updatedAt))
    .slice(0, limit)
}

export async function markItemsSyncStatus(ids: string[], syncStatus: SyncStatus, syncError = ""): Promise<void> {
  if (ids.length === 0) {
    return
  }
  const now = new Date().toISOString()
  const db = await getDB()
  const tx = db.transaction("collector_items", "readwrite")
  for (const id of ids) {
    const item = await tx.store.get(id)
    if (!item) {
      continue
    }
    await tx.store.put({
      ...item,
      syncStatus,
      syncError: syncError || undefined,
      syncedAt: syncStatus === "synced" ? now : item.syncedAt,
      updatedAt: now
    })
  }
  await tx.done
}

export async function getCollectorSummary(): Promise<CollectorSummary> {
  const db = await getDB()
  const items = await db.getAll("collector_items")
  return items.reduce<CollectorSummary>(
    (summary, item) => {
      summary.total += 1
      if (item.valid) {
        summary.valid += 1
      } else {
        summary.invalid += 1
      }
      summary[item.syncStatus] += 1
      return summary
    },
    { total: 0, valid: 0, invalid: 0, pending: 0, synced: 0, failed: 0 }
  )
}

export async function putSyncBatch(batch: SyncBatch): Promise<void> {
  const db = await getDB()
  await db.put("sync_batches", batch)
}

export async function resetCollectorDB(): Promise<void> {
  if (dbPromise) {
    const db = await dbPromise
    db.close()
    dbPromise = undefined
  }
  await deleteDB(DB_NAME)
}

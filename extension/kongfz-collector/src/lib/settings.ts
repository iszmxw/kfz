import { DEFAULT_COLLECTOR_CONFIG, type CollectorConfig } from "./types"

export interface ExtensionSettings extends CollectorConfig {
  adminToken: string
  adminUsername: string
}

export const DEFAULT_SETTINGS: ExtensionSettings = {
  ...DEFAULT_COLLECTOR_CONFIG,
  adminToken: "",
  adminUsername: ""
}

export async function loadSettings(): Promise<ExtensionSettings> {
  const stored = await chromeStorageGet<Partial<ExtensionSettings>>(DEFAULT_SETTINGS)
  return {
    ...DEFAULT_SETTINGS,
    ...stored,
    catId: toPositiveInt(stored.catId, DEFAULT_SETTINGS.catId),
    startPage: toPositiveInt(stored.startPage, DEFAULT_SETTINGS.startPage),
    endPage: stored.endPage ? toPositiveInt(stored.endPage, stored.endPage) : undefined,
    delayMs: toPositiveInt(stored.delayMs, DEFAULT_SETTINGS.delayMs),
    retryTimes: toPositiveInt(stored.retryTimes, DEFAULT_SETTINGS.retryTimes)
  }
}

export async function saveSettings(settings: Partial<ExtensionSettings>): Promise<void> {
  await chromeStorageSet(settings)
}

function toPositiveInt(value: unknown, fallback: number): number {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback
  }
  return Math.floor(parsed)
}

function chromeStorageGet<T>(defaults: T): Promise<T> {
  return new Promise((resolve) => {
    chrome.storage.local.get(defaults as Record<string, unknown>, (items) => resolve(items as T))
  })
}

function chromeStorageSet(values: Record<string, unknown>): Promise<void> {
  return new Promise((resolve) => {
    chrome.storage.local.set(values, () => resolve())
  })
}

import "fake-indexeddb/auto"

import { beforeEach, describe, expect, it } from "vitest"
import { getCollectorSummary, getUnsyncedValidItems, markItemsSyncStatus, putCollectorItems, resetCollectorDB } from "../src/lib/db"
import type { CollectorItem } from "../src/lib/types"

describe("collector db", () => {
  beforeEach(async () => {
    await resetCollectorDB()
  })

  it("deduplicates valid isbn rows and resets changed raw hash to pending", async () => {
    const first = sampleItem({ rawHash: "a", syncStatus: "pending" })
    await putCollectorItems([first])
    await markItemsSyncStatus([first.id], "synced")
    expect((await getCollectorSummary()).synced).toBe(1)

    await putCollectorItems([sampleItem({ rawHash: "a", title: "活着" })])
    expect((await getCollectorSummary()).synced).toBe(1)

    await putCollectorItems([sampleItem({ rawHash: "b", title: "活着新版" })])
    const summary = await getCollectorSummary()
    expect(summary.total).toBe(1)
    expect(summary.pending).toBe(1)
    expect(summary.synced).toBe(0)
    const pending = await getUnsyncedValidItems()
    expect(pending[0].title).toBe("活着新版")
  })
})

function sampleItem(patch: Partial<CollectorItem> = {}): CollectorItem {
  const now = new Date().toISOString()
  return {
    id: "9787506365437",
    runId: "run-1",
    catId: 43,
    page: 1,
    index: 0,
    valid: true,
    errorMessage: "",
    isbn: "9787506365437",
    normalizedIsbn: "9787506365437",
    title: "活着",
    author: "余华",
    publisher: "作家出版社",
    publishYear: "2012",
    publishDate: "2012-08",
    binding: "平装",
    listPrice: "20.00",
    coverUrl: "",
    oldBookMinPrice: "0.01",
    oldBookMinPriceText: "0.01",
    oldBookOnSaleNum: 4708,
    newBookMinPriceText: "",
    newBookOnSaleNum: 0,
    onSaleProductNum: 4708,
    bookShowInfo: [],
    rawPayload: {},
    rawHash: "a",
    rawUrl: "https://example.com/api",
    syncStatus: "pending",
    createdAt: now,
    updatedAt: now,
    ...patch
  }
}

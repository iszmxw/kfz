import { describe, expect, it } from "vitest"
import { collectPages } from "../src/lib/collector"
import type { CollectorItem, CollectorStatus, KongfzApiResponse } from "../src/lib/types"

describe("collector", () => {
  it("collects pages and retries transient failures", async () => {
    let attempts = 0
    const saved: CollectorItem[] = []
    const status = await collectPages(
      { backendBaseUrl: "http://127.0.0.1:8888", catId: 43, startPage: 1, endPage: 1, delayMs: 0, retryTimes: 1 },
      {
        runId: "run-1",
        fetchPage: async () => {
          attempts += 1
          if (attempts === 1) {
            throw new Error("network")
          }
          return responseWithOneItem()
        },
        saveItems: async (items) => {
          saved.push(...items)
        },
        delay: async () => {}
      }
    )

    expect(attempts).toBe(2)
    expect(saved).toHaveLength(1)
    expect(status.status).toBe("completed")
  })

  it("pauses and resumes from predicates", async () => {
    let paused = true
    const statuses: CollectorStatus[] = []
    await collectPages(
      { backendBaseUrl: "http://127.0.0.1:8888", catId: 43, startPage: 1, endPage: 1, delayMs: 0, retryTimes: 0 },
      {
        runId: "run-1",
        fetchPage: async () => responseWithOneItem(),
        saveItems: async () => {},
        shouldPause: () => paused,
        delay: async () => {
          paused = false
        },
        onProgress: (status) => {
          statuses.push(status)
        }
      }
    )

    expect(statuses.some((status) => status.status === "paused")).toBe(true)
    expect(statuses.at(-1)?.status).toBe("completed")
  })
})

function responseWithOneItem(): KongfzApiResponse {
  return {
    status: 1,
    data: {
      itemResponse: {
        pager: { page: 1, pages: 1, total: 1, size: 50 },
        list: [
          {
            bookName: "活着",
            isbn: "9787506365437",
            oldBookMinPrice: 0.01,
            oldBookOnSaleNum: 4708,
            bookShowInfo: ["余华  著", "作家出版社", "2012-08", "平装", "20.00"]
          }
        ]
      }
    }
  }
}

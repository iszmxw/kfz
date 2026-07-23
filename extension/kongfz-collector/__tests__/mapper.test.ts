import { describe, expect, it } from "vitest"
import { mapKongfzItem, validateKongfzResponse } from "../src/lib/kongfzMapper"

describe("kongfz mapper", () => {
  it("maps bookShowInfo and conservative price fields", async () => {
    const item = await mapKongfzItem(
      {
        bookName: "活着",
        isbn: "9787506365437",
        imgUrlEntity: { bigImgUrl: "https://example.com/cover.jpg" },
        id: 17,
        mid: 59456769,
        oldBookMinPrice: 0.01,
        oldBookMinPriceText: "0.01",
        oldBookOnSaleNum: 4708,
        bookShowInfo: ["余华  著", "作家出版社", "2012-08", "平装", "20.00"]
      },
      { runId: "run-1", catId: 43, page: 1, index: 0, rawUrl: "https://example.com/api" }
    )

    expect(item.id).toBe("9787506365437")
    expect(item.valid).toBe(true)
    expect(item.title).toBe("活着")
    expect(item.author).toBe("余华  著")
    expect(item.publisher).toBe("作家出版社")
    expect(item.publishYear).toBe("2012")
    expect(item.binding).toBe("平装")
    expect(item.listPrice).toBe("20.00")
    expect(item.oldBookMinPrice).toBe("0.01")
  })

  it("keeps invalid isbn rows", async () => {
    const item = await mapKongfzItem(
      { bookName: "坏数据", isbn: "bad-isbn", bookShowInfo: [] },
      { runId: "run-1", catId: 43, page: 1, index: 2, rawUrl: "https://example.com/api" }
    )

    expect(item.id).toBe("invalid:run-1:1:2")
    expect(item.valid).toBe(false)
    expect(item.errorMessage).toContain("ISBN格式无效")
  })

  it("validates response shape", () => {
    expect(() =>
      validateKongfzResponse({
        status: 1,
        data: { itemResponse: { list: [], pager: { pages: 100 } } }
      })
    ).not.toThrow()

    expect(() => validateKongfzResponse({ status: 0, message: "failed" })).toThrow("failed")
  })
})

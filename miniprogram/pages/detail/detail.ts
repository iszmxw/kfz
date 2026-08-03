import Toast from "@vant/weapp/toast/toast";
import type { BookDetailResponse } from "../../types/api";
import { getBookDetail } from "../../services/book";
import { decisionClass, decisionLabel, formatDateTime, formatNullableMoney } from "../../utils/format";

interface DetailData {
  isbn: string;
  loading: boolean;
  detail: BookDetailResponse | null;
  decisionTone: string;
  decisionText: string;
  decisionTagColor: string;
  coverText: string;
  bookTitleText: string;
  bookAuthorText: string;
  bookPublishText: string;
  marketMinText: string;
  marketAvgText: string;
  marketMaxText: string;
  marketSampleText: string;
  marketSourceText: string;
  confidenceText: string;
  confidenceDesc: string;
  suggestedPriceText: string;
}

const decisionColor: Record<string, string> = {
  accept: "#0f9f6e",
  reject: "#b93d3d",
  review: "#d97706"
};

Page<DetailData, WechatMiniprogram.Page.CustomOption>({
  data: {
    isbn: "",
    loading: false,
    detail: null,
    decisionTone: "",
    decisionText: "",
    decisionTagColor: "#969799",
    coverText: "样书",
    bookTitleText: "未知图书",
    bookAuthorText: "未知作者",
    bookPublishText: "未知出版社 · 年份缺失",
    marketMinText: "-",
    marketAvgText: "-",
    marketMaxText: "-",
    marketSampleText: "0 条",
    marketSourceText: "-",
    confidenceText: "NONE",
    confidenceDesc: "暂无价格数据，已按书籍存在默认建议回收",
    suggestedPriceText: "-"
  },

  onLoad(options: Record<string, string | undefined>) {
    const isbn = options.isbn || "";
    this.setData({ isbn });
    this.loadDetail();
  },

  async loadDetail() {
    if (!this.data.isbn) {
      Toast.fail("缺少 ISBN");
      return;
    }

    this.setData({ loading: true });
    try {
      const detail = await getBookDetail(this.data.isbn);
      this.setData({ loading: false, detail });
      this.refreshView(detail);
    } catch (error) {
      const message = error instanceof Error ? error.message : "网络异常，请重试";
      this.setData({ loading: false, detail: null });
      Toast.fail(message);
    }
  },

  refreshView(detail: BookDetailResponse) {
    const decision = detail.latest_decision.decision;
    const tone = decisionClass(decision);
    const price = detail.latest_market_price;
    const confidence = price?.confidence || "NONE";
    this.setData({
      decisionTone: tone,
      decisionText: decisionLabel(decision),
      decisionTagColor: decisionColor[tone],
      coverText: decision === "ACCEPT" ? "可回收\n样书" : decision === "REJECT" ? "未录入\n图书" : "待确认\n样书",
      bookTitleText: detail.book.title || "未知图书",
      bookAuthorText: detail.book.author || "未知作者",
      bookPublishText: `${detail.book.publisher || "未知出版社"} · ${detail.book.publish_year || "年份缺失"}`,
      marketMinText: formatNullableMoney(price?.min),
      marketAvgText: formatNullableMoney(price?.avg),
      marketMaxText: formatNullableMoney(price?.max),
      marketSampleText: `${price?.sample_count || 0} 条`,
      marketSourceText: price ? `${price.source} · ${formatDateTime(price.collected_at)}` : "-",
      confidenceText: confidence,
      confidenceDesc: price ? "价格数据仅用于建议价参考" : "暂无价格数据，已按书籍存在默认建议回收",
      suggestedPriceText: formatNullableMoney(detail.latest_decision.suggested_recycle_price)
    });
  },

  reload() {
    this.loadDetail();
  },

  goBack() {
    wx.switchTab({ url: "/pages/index/index" });
  }
});

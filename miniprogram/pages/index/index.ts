import Toast from "@vant/weapp/toast/toast";
import type { BookCheckResponse } from "../../types/api";
import { checkBook } from "../../services/book";
import { createClientRequestId, playDecisionVoice, vibrateShort } from "../../utils/feedback";
import { decisionClass, decisionLabel, formatDateTime, formatNullableMoney } from "../../utils/format";
import { extractISBN, normalizeISBN } from "../../utils/isbn";
import { defaultSettings, getRecentScan, getSettings, saveRecentScan, saveSettings } from "../../utils/storage";

interface IndexData {
  checking: boolean;
  manualISBN: string;
  currentResult: BookCheckResponse | null;
  recentResult: BookCheckResponse | null;
  settings: typeof defaultSettings;
  showErrorPopup: boolean;
  errorTitle: string;
  errorMessage: string;
  todayStats: {
    scanned: number;
    accepted: number;
    amount: string;
  };
  scanButtonText: string;
  detailDisabled: boolean;
  showSuggestedPrice: boolean;
  showManualActions: boolean;
  showRecentOnly: boolean;
  decisionTone: string;
  decisionLabel: string;
  decisionTitle: string;
  suggestedPriceText: string;
  coverText: string;
  bookTitleText: string;
  bookPublishText: string;
  marketMinText: string;
  marketAvgText: string;
  marketSampleText: string;
  recentTitleText: string;
  recentTimeText: string;
  recentPriceText: string;
  recentDecisionLabel: string;
  recentTagColor: string;
}

Page<IndexData, WechatMiniprogram.Page.CustomOption>({
  data: {
    checking: false,
    manualISBN: "",
    currentResult: null,
    recentResult: null,
    settings: defaultSettings,
    showErrorPopup: false,
    errorTitle: "",
    errorMessage: "",
    todayStats: {
      scanned: 12,
      accepted: 5,
      amount: "¥46"
    },
    scanButtonText: "扫码识别",
    detailDisabled: true,
    showSuggestedPrice: false,
    showManualActions: false,
    showRecentOnly: false,
    decisionTone: "",
    decisionLabel: "",
    decisionTitle: "",
    suggestedPriceText: "-",
    coverText: "样书",
    bookTitleText: "未知图书",
    bookPublishText: "未知出版社 · 年份缺失",
    marketMinText: "-",
    marketAvgText: "-",
    marketSampleText: "0",
    recentTitleText: "未知图书",
    recentTimeText: "-",
    recentPriceText: "-",
    recentDecisionLabel: "",
    recentTagColor: "#969799"
  },

  onShow() {
    const settings = getSettings();
    const recentResult = getRecentScan();
    this.setData({ settings, recentResult, detailDisabled: !recentResult });
    this.refreshRecentView(recentResult);
  },

  onScan() {
    if (this.data.checking) {
      return;
    }

    wx.scanCode({
      scanType: ["barCode"],
      success: (res) => {
        const isbn = extractISBN(res.result);
        this.submitISBN(isbn);
      },
      fail: () => {
        this.showError("扫码失败", "请重新对准图书 ISBN 条形码。");
      }
    });
  },

  onManualISBNChange(event: WechatMiniprogram.CustomEvent<{ value: string }>) {
    this.setData({ manualISBN: event.detail.value });
  },

  onManualCheck() {
    this.submitISBN(normalizeISBN(this.data.manualISBN));
  },

  async submitISBN(isbn: string) {
    if (!isbn) {
      Toast.fail("未识别到有效 ISBN");
      this.showError("识别失败", "未识别到有效 ISBN，请确认扫描的是图书 ISBN 条形码。");
      return;
    }

    this.setData({ checking: true, showErrorPopup: false, scanButtonText: "扫码中..." });
    Toast.loading({
      message: "查询中...",
      forbidClick: true,
      duration: 0
    });

    try {
      const result = await checkBook({
        isbn,
        clientRequestId: createClientRequestId()
      });
      saveRecentScan(result);
      this.setData({
        currentResult: result,
        recentResult: result,
        checking: false,
        scanButtonText: "扫码识别",
        detailDisabled: false,
        showRecentOnly: false
      });
      this.refreshResultView(result);
      this.refreshRecentView(result);
      this.afterDecision(result);
      Toast.clear();
    } catch (error) {
      const message = error instanceof Error ? error.message : "网络异常，请重试";
      this.setData({ checking: false, scanButtonText: "扫码识别" });
      Toast.clear();
      Toast.fail(message);
      this.showError(message === "网络异常，请重试" ? "网络异常" : "查询失败", message);
    }
  },

  afterDecision(result: BookCheckResponse) {
    if (this.data.settings.vibrateEnabled) {
      vibrateShort();
    }
    if (this.data.settings.voiceEnabled) {
      playDecisionVoice(result.decision);
    }
    if (result.duplicate.scanned_recently) {
      Toast("该 ISBN 近期扫过");
    }
  },

  refreshResultView(result: BookCheckResponse) {
    const tone = decisionClass(result.decision);
    const price = result.market_price;
    this.setData({
      decisionTone: tone,
      decisionLabel: decisionLabel(result.decision),
      decisionTitle: result.decision === "ACCEPT" ? "建议回收" : result.decision === "REJECT" ? "不建议回收" : "样本不足，人工判断",
      suggestedPriceText: formatNullableMoney(result.suggested_recycle_price),
      showSuggestedPrice: result.suggested_recycle_price !== null,
      showManualActions: result.decision === "NEED_REVIEW",
      coverText: result.decision === "ACCEPT" ? "可回收\n样书" : result.decision === "REJECT" ? "低价\n样书" : "待确认\n样书",
      bookTitleText: result.book.title || "未知图书",
      bookPublishText: `${result.book.publisher || "未知出版社"} · ${result.book.publish_year || "年份缺失"}`,
      marketMinText: formatNullableMoney(price?.min),
      marketAvgText: formatNullableMoney(price?.avg),
      marketSampleText: price ? String(price.sample_count) : "0"
    });
  },

  refreshRecentView(result: BookCheckResponse | null) {
    if (!result) {
      return;
    }
    const tone = decisionClass(result.decision);
    const colorMap: Record<string, string> = {
      accept: "#0f9f6e",
      reject: "#b93d3d",
      review: "#d97706"
    };
    this.setData({
      recentTitleText: result.book.title || "未知图书",
      recentTimeText: formatDateTime(result.scanned_at),
      recentPriceText: formatNullableMoney(result.suggested_recycle_price),
      recentDecisionLabel: decisionLabel(result.decision),
      recentTagColor: colorMap[tone],
      showRecentOnly: !this.data.currentResult
    });
  },

  restoreRecent() {
    if (!this.data.recentResult) {
      return;
    }
    this.setData({ currentResult: this.data.recentResult });
    this.refreshResultView(this.data.recentResult);
  },

  onToggleVoice(event: WechatMiniprogram.CustomEvent<{ value?: boolean }>) {
    const checked = typeof event.detail === "boolean" ? event.detail : Boolean(event.detail.value);
    const settings = {
      ...this.data.settings,
      voiceEnabled: checked
    };
    saveSettings(settings);
    this.setData({ settings });
  },

  onManualAccept() {
    Toast("原型展示：人工选择“收”不提交");
  },

  onManualReject() {
    Toast("原型展示：人工选择“不收”不提交");
  },

  goHistory() {
    wx.switchTab({ url: "/pages/history/history" });
  },

  goSettings() {
    wx.switchTab({ url: "/pages/settings/settings" });
  },

  goDetail() {
    const result = this.data.currentResult || this.data.recentResult;
    if (!result) {
      return;
    }
    wx.navigateTo({
      url: `/pages/detail/detail?isbn=${result.book.normalized_isbn}`
    });
  },

  showError(title: string, message: string) {
    this.setData({
      showErrorPopup: true,
      errorTitle: title,
      errorMessage: message
    });
  },

  closeErrorPopup() {
    this.setData({ showErrorPopup: false });
  }
});

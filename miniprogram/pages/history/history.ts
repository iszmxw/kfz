import Toast from "@vant/weapp/toast/toast";
import type { Decision, ScanHistoryItem } from "../../types/api";
import { getScanHistory } from "../../services/scan";
import { decisionClass, decisionLabel, formatDateTime, formatNullableMoney } from "../../utils/format";

type HistoryTab = "ALL" | Decision;

interface HistoryViewItem extends ScanHistoryItem {
  titleText: string;
  decisionText: string;
  scannedAtText: string;
  priceText: string;
  tagColor: string;
}

interface HistoryData {
  activeTab: HistoryTab;
  loading: boolean;
  items: HistoryViewItem[];
}

const decisionColor: Record<string, string> = {
  accept: "#0f9f6e",
  reject: "#b93d3d",
  review: "#d97706"
};

Page<HistoryData, WechatMiniprogram.Page.CustomOption>({
  data: {
    activeTab: "ALL",
    loading: false,
    items: []
  },

  onShow() {
    this.loadHistory();
  },

  onTabChange(event: WechatMiniprogram.CustomEvent<{ name: HistoryTab }>) {
    this.setData({ activeTab: event.detail.name });
    this.loadHistory();
  },

  async loadHistory() {
    this.setData({ loading: true });
    try {
      const activeTab = this.data.activeTab;
      const resp = await getScanHistory({
        page: 1,
        pageSize: 20,
        decision: activeTab === "ALL" ? undefined : activeTab
      });
      this.setData({
        loading: false,
        items: resp.items.map((item) => this.toViewItem(item))
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : "网络异常，请重试";
      this.setData({ loading: false, items: [] });
      Toast.fail(message);
    }
  },

  toViewItem(item: ScanHistoryItem): HistoryViewItem {
    const tone = decisionClass(item.decision);
    return {
      ...item,
      titleText: item.title || "未知图书",
      decisionText: decisionLabel(item.decision),
      scannedAtText: formatDateTime(item.scanned_at),
      priceText: `建议回收价 ${formatNullableMoney(item.suggested_recycle_price)}`,
      tagColor: decisionColor[tone]
    };
  },

  goDetail(event: WechatMiniprogram.TouchEvent) {
    const isbn = event.currentTarget.dataset.isbn as string | undefined;
    if (!isbn) {
      return;
    }
    wx.navigateTo({
      url: `/pages/detail/detail?isbn=${isbn}`
    });
  }
});

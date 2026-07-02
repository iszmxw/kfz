interface TabItem {
  pagePath: string;
  text: string;
  icon: string;
  activeIcon: string;
}

const tabs: TabItem[] = [
  {
    pagePath: "pages/index/index",
    text: "扫码",
    icon: "qr",
    activeIcon: "scan"
  },
  {
    pagePath: "pages/history/history",
    text: "历史",
    icon: "records-o",
    activeIcon: "records"
  },
  {
    pagePath: "pages/settings/settings",
    text: "设置",
    icon: "setting-o",
    activeIcon: "setting"
  }
];

Component({
  data: {
    selected: "pages/index/index",
    hidden: false,
    color: "#6f7c8e",
    selectedColor: "#1677ff",
    tabs
  },

  pageLifetimes: {
    show() {
      this.syncSelected();
    }
  },

  methods: {
    onSwitchTab(event: WechatMiniprogram.TouchEvent) {
      const pagePath = event.currentTarget.dataset.path as string | undefined;
      if (!pagePath || pagePath === this.data.selected) {
        return;
      }

      wx.switchTab({
        url: `/${pagePath}`
      });
    },

    syncSelected() {
      const pages = getCurrentPages();
      const currentPage = pages[pages.length - 1];
      const route = currentPage?.route || "pages/index/index";

      if (route !== this.data.selected) {
        this.setData({ selected: route });
      }
    }
  }
});

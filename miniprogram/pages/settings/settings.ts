import Toast from "@vant/weapp/toast/toast";
import { defaultSettings, getSettings, saveSettings, type AppSettings } from "../../utils/storage";

interface SettingsData {
  settings: AppSettings;
}

Page<SettingsData, WechatMiniprogram.Page.CustomOption>({
  data: {
    settings: defaultSettings
  },

  onShow() {
    this.setData({ settings: getSettings() });
  },

  onSwitchChange(event: WechatMiniprogram.CustomEvent<{ value?: boolean }>) {
    const key = event.currentTarget.dataset.key as keyof AppSettings | undefined;
    if (!key) {
      return;
    }
    const checked = typeof event.detail === "boolean" ? event.detail : Boolean(event.detail.value);
    const settings = {
      ...this.data.settings,
      [key]: checked
    };
    saveSettings(settings);
    this.setData({ settings });
    Toast.success("已保存");
  }
});

import type { BookCheckResponse } from "../types/api";

export interface AppSettings {
  voiceEnabled: boolean;
  vibrateEnabled: boolean;
  showPriceDetail: boolean;
}

const settingsKey = "kfz_settings";
const recentScanKey = "kfz_recent_scan";

export const defaultSettings: AppSettings = {
  voiceEnabled: true,
  vibrateEnabled: true,
  showPriceDetail: true
};

export function getSettings(): AppSettings {
  const settings = wx.getStorageSync(settingsKey) as Partial<AppSettings> | "";
  return {
    ...defaultSettings,
    ...(settings || {})
  };
}

export function saveSettings(settings: AppSettings): void {
  wx.setStorageSync(settingsKey, settings);
}

export function getRecentScan(): BookCheckResponse | null {
  return (wx.getStorageSync(recentScanKey) as BookCheckResponse | "") || null;
}

export function saveRecentScan(scan: BookCheckResponse): void {
  wx.setStorageSync(recentScanKey, scan);
}

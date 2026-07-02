import { config } from "../config/index";
import type { EchoResponse } from "../types/api";

export interface RequestOptions {
  url: string;
  method?: WechatMiniprogram.RequestOption["method"];
  data?: WechatMiniprogram.IAnyObject;
}

export function request<T>(options: RequestOptions): Promise<T> {
  return new Promise((resolve, reject) => {
    wx.request<EchoResponse<T>>({
      url: `${config.baseUrl}${options.url}`,
      method: options.method || "GET",
      data: options.data,
      success: (res) => {
        const body = res.data;
        if (res.statusCode < 200 || res.statusCode >= 300) {
          reject(new Error("网络异常，请重试"));
          return;
        }
        if (!body || body.code !== 1) {
          reject(new Error(body?.msg || "请求失败，请重试"));
          return;
        }
        resolve(body.data as T);
      },
      fail: () => {
        reject(new Error("网络异常，请重试"));
      }
    });
  });
}

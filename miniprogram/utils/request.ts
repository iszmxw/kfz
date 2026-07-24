import { config } from "../config/index";
import type { EchoResponse } from "../types/api";

export interface RequestOptions {
  url: string;
  method?: WechatMiniprogram.RequestOption["method"];
  data?: WechatMiniprogram.IAnyObject;
}

function requestFailMessage(errMsg: string): string {
  const detail = errMsg ? `（${errMsg}）` : "";

  if (errMsg.includes("url not in domain list") || errMsg.includes("不在以下 request 合法域名列表中")) {
    return `请求域名未配置到小程序合法域名：${config.baseUrl}${detail}`;
  }
  if (errMsg.includes("ssl") || errMsg.includes("TLS") || errMsg.includes("certificate")) {
    return `HTTPS 证书校验失败，请检查后端域名证书：${config.baseUrl}${detail}`;
  }
  if (errMsg.includes("timeout")) {
    return `接口请求超时，请确认网络和后端状态：${config.baseUrl}${detail}`;
  }
  if (errMsg.includes("ERR_NAME_NOT_RESOLVED") || errMsg.includes("resolve host")) {
    return `接口域名解析失败，请检查域名或网络：${config.baseUrl}${detail}`;
  }
  if (errMsg.includes("connect") || errMsg.includes("refused") || errMsg.includes("fail")) {
    return `接口不可达，请确认后端已启动：${config.baseUrl}${detail}`;
  }
  return `网络异常，请重试${detail}`;
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
      fail: (error) => {
        const errMsg = typeof error?.errMsg === "string" ? error.errMsg : "";
        reject(new Error(requestFailMessage(errMsg)));
      }
    });
  });
}

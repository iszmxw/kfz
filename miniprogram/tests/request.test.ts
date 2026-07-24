import { describe, expect, it, vi } from "vitest";
import { config } from "../config/index";
import { request } from "../utils/request";

describe("request", () => {
  it("unwraps successful echo response data", async () => {
    const wxMock = {
      request: vi.fn((options) => {
        options.success({
          statusCode: 200,
          data: {
            code: 1,
            msg: "success.",
            reqId: "req_001",
            data: { ok: true }
          }
        });
      })
    };
    vi.stubGlobal("wx", wxMock);

    await expect(request<{ ok: boolean }>({ url: "/ping" })).resolves.toEqual({ ok: true });
    expect(wxMock.request).toHaveBeenCalledWith(expect.objectContaining({
      method: "GET",
      url: `${config.baseUrl}/ping`
    }));
  });

  it("rejects business errors with backend message", async () => {
    vi.stubGlobal("wx", {
      request: vi.fn((options) => {
        options.success({
          statusCode: 200,
          data: { code: 0, msg: "未识别到有效 ISBN", reqId: "req_002" }
        });
      })
    });

    await expect(request({ url: "/bad" })).rejects.toThrow("未识别到有效 ISBN");
  });

  it("rejects request failures with endpoint hint", async () => {
    vi.stubGlobal("wx", {
      request: vi.fn((options) => {
        options.fail({ errMsg: "request:fail" });
      })
    });

    await expect(request({ url: "/offline" })).rejects.toThrow("接口不可达，请确认后端已启动");
  });

  it("rejects domain check failures with domain setup hint", async () => {
    vi.stubGlobal("wx", {
      request: vi.fn((options) => {
        options.fail({
          errMsg: "request:fail url not in domain list"
        });
      })
    });

    await expect(request({ url: "/domain-check" })).rejects.toThrow("请求域名未配置到小程序合法域名");
  });
});

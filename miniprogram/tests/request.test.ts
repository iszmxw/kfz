import { describe, expect, it, vi } from "vitest";
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
      url: "http://127.0.0.1:8888/ping"
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

  it("rejects network failures with default network message", async () => {
    vi.stubGlobal("wx", {
      request: vi.fn((options) => {
        options.fail({ errMsg: "request:fail" });
      })
    });

    await expect(request({ url: "/offline" })).rejects.toThrow("网络异常，请重试");
  });
});

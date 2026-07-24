import type { RuntimeMessage, RuntimeResponse } from "./types"

export function sendRuntimeMessage<TPayload, TResult>(type: string, payload?: TPayload): Promise<RuntimeResponse<TResult>> {
  const message: RuntimeMessage<TPayload> = { type, payload }
  return new Promise((resolve) => {
    chrome.runtime.sendMessage(message, (response?: RuntimeResponse<TResult>) => {
      if (chrome.runtime.lastError) {
        resolve({ ok: false, error: chrome.runtime.lastError.message })
        return
      }
      resolve(response ?? { ok: false, error: "后台没有返回结果" })
    })
  })
}

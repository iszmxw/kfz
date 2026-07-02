import type { Decision } from "../types/api";

export function playDecisionVoice(_decision: Decision): void {
  // Audio assets are not ready in V1. Keep this seam for shou.mp3/bushou.mp3 later.
}

export function vibrateShort(): void {
  wx.vibrateShort({ type: "medium" });
}

export function createClientRequestId(): string {
  return `mini_${Date.now()}_${Math.random().toString(16).slice(2)}`;
}

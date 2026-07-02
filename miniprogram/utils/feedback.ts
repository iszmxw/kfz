import type { Decision } from "../types/api";

const decisionAudioMap: Record<Decision, string> = {
  ACCEPT: "/assets/audio/shou.mp3",
  REJECT: "/assets/audio/bushou.mp3",
  NEED_REVIEW: "/assets/audio/xuqueren.mp3"
};

export function playDecisionVoice(decision: Decision): void {
  const src = decisionAudioMap[decision];
  if (!src) {
    return;
  }

  const audio = wx.createInnerAudioContext();
  audio.src = src;
  audio.obeyMuteSwitch = false;
  audio.onEnded(() => audio.destroy());
  audio.onError(() => audio.destroy());
  audio.play();
}

export function vibrateShort(): void {
  wx.vibrateShort({ type: "medium" });
}

export function createClientRequestId(): string {
  return `mini_${Date.now()}_${Math.random().toString(16).slice(2)}`;
}

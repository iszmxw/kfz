import type { Decision } from "../types/api";

const decisionAudioMap: Record<Decision, string> = {
  ACCEPT: "/assets/audio/shou.mp3",
  REJECT: "/assets/audio/bushou.mp3",
  NEED_REVIEW: "/assets/audio/xuqueren.mp3"
};

const activeAudios: WechatMiniprogram.InnerAudioContext[] = [];
let appInForeground = true;
let pendingDecision: Decision | null = null;
let pendingTimer: number | null = null;

function releaseAudio(audio: WechatMiniprogram.InnerAudioContext): void {
  const index = activeAudios.indexOf(audio);
  if (index >= 0) {
    activeAudios.splice(index, 1);
  }
  audio.destroy();
}

function schedulePendingVoice(delay = 300): void {
  if (!pendingDecision || pendingTimer !== null) {
    return;
  }

  pendingTimer = setTimeout(() => {
    pendingTimer = null;
    const decision = pendingDecision;
    pendingDecision = null;
    if (decision) {
      playDecisionVoice(decision);
    }
  }, delay) as unknown as number;
}

function queueDecisionVoice(decision: Decision, delay = 300): void {
  pendingDecision = decision;
  if (appInForeground) {
    schedulePendingVoice(delay);
  }
}

export function initAudioFeedback(): void {
  if (typeof wx.setInnerAudioOption !== "function") {
    return;
  }

  wx.setInnerAudioOption({
    obeyMuteSwitch: false,
    mixWithOther: true,
    fail: (error) => {
      console.warn("setInnerAudioOption failed", error);
    }
  });
}

export function setAudioFeedbackForeground(isForeground: boolean): void {
  appInForeground = isForeground;
  if (isForeground) {
    schedulePendingVoice();
  }
}

export function playDecisionVoice(decision: Decision): void {
  if (!appInForeground) {
    queueDecisionVoice(decision);
    return;
  }

  const src = decisionAudioMap[decision];
  if (!src) {
    return;
  }

  const audio = wx.createInnerAudioContext();
  audio.src = src;
  audio.obeyMuteSwitch = false;
  activeAudios.push(audio);
  audio.onEnded(() => releaseAudio(audio));
  audio.onError((error) => {
    console.warn("playDecisionVoice failed", src, error);
    releaseAudio(audio);
    const errMsg = typeof error?.errMsg === "string" ? error.errMsg : "";
    if (errMsg.includes("runningState=background")) {
      queueDecisionVoice(decision, 600);
    }
  });
  audio.play();
}

export function vibrateShort(): void {
  wx.vibrateShort({ type: "medium" });
}

export function createClientRequestId(): string {
  return `mini_${Date.now()}_${Math.random().toString(16).slice(2)}`;
}

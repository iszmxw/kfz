import type { Decision } from "../types/api";

const decisionLabels: Record<Decision, string> = {
  ACCEPT: "收",
  REJECT: "不收",
  NEED_REVIEW: "需确认"
};

const decisionClasses: Record<Decision, string> = {
  ACCEPT: "accept",
  REJECT: "reject",
  NEED_REVIEW: "review"
};

export function formatMoney(value: number): string {
  return `¥${value.toFixed(2)}`;
}

export function formatNullableMoney(value: number | null | undefined): string {
  if (value === null || value === undefined) {
    return "-";
  }
  return formatMoney(value);
}

export function decisionLabel(decision: Decision): string {
  return decisionLabels[decision];
}

export function decisionClass(decision: Decision): string {
  return decisionClasses[decision];
}

export function isAcceptedDecision(decision: Decision): boolean {
  return decision === "ACCEPT";
}

export function formatDateTime(value: string | null | undefined): string {
  if (!value) {
    return "-";
  }

  const match = value.match(/^(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2})/);
  if (match) {
    return `${match[1]} ${match[2]}`;
  }

  return value;
}

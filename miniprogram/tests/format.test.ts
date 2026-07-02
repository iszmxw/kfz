import { describe, expect, it } from "vitest";
import {
  decisionClass,
  decisionLabel,
  formatDateTime,
  formatMoney,
  formatNullableMoney,
  isAcceptedDecision
} from "../utils/format";

describe("formatMoney", () => {
  it("formats numbers as yuan with two decimals", () => {
    expect(formatMoney(7.35)).toBe("¥7.35");
    expect(formatMoney(7.4)).toBe("¥7.40");
  });

  it("formats null money as placeholder", () => {
    expect(formatNullableMoney(null)).toBe("-");
    expect(formatNullableMoney(undefined)).toBe("-");
  });
});

describe("decision helpers", () => {
  it("maps decisions to display labels and classes", () => {
    expect(decisionLabel("ACCEPT")).toBe("收");
    expect(decisionLabel("REJECT")).toBe("不收");
    expect(decisionLabel("NEED_REVIEW")).toBe("需确认");
    expect(decisionClass("ACCEPT")).toBe("accept");
    expect(decisionClass("REJECT")).toBe("reject");
    expect(decisionClass("NEED_REVIEW")).toBe("review");
  });

  it("identifies only ACCEPT as accepted decision", () => {
    expect(isAcceptedDecision("ACCEPT")).toBe(true);
    expect(isAcceptedDecision("REJECT")).toBe(false);
    expect(isAcceptedDecision("NEED_REVIEW")).toBe(false);
  });
});

describe("formatDateTime", () => {
  it("keeps local readable date and time", () => {
    expect(formatDateTime("2026-06-30T10:30:00+08:00")).toBe("2026-06-30 10:30");
  });

  it("returns placeholder for empty time", () => {
    expect(formatDateTime("")).toBe("-");
  });
});

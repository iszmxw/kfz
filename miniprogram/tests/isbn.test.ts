import { describe, expect, it } from "vitest";
import { extractISBN, normalizeISBN } from "../utils/isbn";

describe("normalizeISBN", () => {
  it("normalizes ISBN-13 by trimming separators", () => {
    expect(normalizeISBN(" 978-7-111-12806-9 ")).toBe("9787111128069");
  });

  it("accepts ISBN-10 with trailing X", () => {
    expect(normalizeISBN("7-111-12806-X")).toBe("711112806X");
  });

  it("rejects values that are not ISBN-10 or ISBN-13", () => {
    expect(normalizeISBN("978-invalid")).toBe("");
    expect(normalizeISBN("123456789")).toBe("");
  });
});

describe("extractISBN", () => {
  it("extracts ISBN from scan text", () => {
    expect(extractISBN("ISBN: 978-7-111-12806-9")).toBe("9787111128069");
  });

  it("returns empty string when no valid ISBN exists", () => {
    expect(extractISBN("qr-code:abc")).toBe("");
  });
});

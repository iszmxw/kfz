export function normalizeISBN(value: unknown): string {
  const normalized = String(value ?? "")
    .trim()
    .toUpperCase()
    .replaceAll("-", "")
    .replaceAll(" ", "")

  if (normalized.length === 10) {
    for (let index = 0; index < normalized.length; index += 1) {
      const ch = normalized[index]
      const isDigit = ch >= "0" && ch <= "9"
      if (isDigit || (index === 9 && ch === "X")) {
        continue
      }
      return ""
    }
    return normalized
  }

  if (normalized.length === 13) {
    for (const ch of normalized) {
      if (ch < "0" || ch > "9") {
        return ""
      }
    }
    return normalized
  }

  return ""
}

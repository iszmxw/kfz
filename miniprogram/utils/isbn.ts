export function normalizeISBN(value: string): string {
  const normalized = value.trim().toUpperCase().replace(/[-\s]/g, "");

  if (/^\d{13}$/.test(normalized)) {
    return normalized;
  }

  if (/^\d{9}[\dX]$/.test(normalized)) {
    return normalized;
  }

  return "";
}

export function extractISBN(value: string): string {
  const compact = value.toUpperCase().replace(/[-\s:：]/g, "");
  const isbn13 = compact.match(/\d{13}/);
  if (isbn13) {
    return normalizeISBN(isbn13[0]);
  }

  const isbn10 = compact.match(/\d{9}[\dX]/);
  if (isbn10) {
    return normalizeISBN(isbn10[0]);
  }

  return "";
}

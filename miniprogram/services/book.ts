import type { BookCheckResponse, BookDetailResponse } from "../types/api";
import { request } from "../utils/request";

export interface CheckBookParams {
  isbn: string;
  clientRequestId: string;
}

export function checkBook(params: CheckBookParams): Promise<BookCheckResponse> {
  return request<BookCheckResponse>({
    url: "/app/v1/book/check.json",
    data: {
      isbn: params.isbn,
      client_request_id: params.clientRequestId
    }
  });
}

export function getBookDetail(isbn: string): Promise<BookDetailResponse> {
  return request<BookDetailResponse>({
    url: "/app/v1/book/detail.json",
    data: { isbn }
  });
}

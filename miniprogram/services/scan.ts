import type { Decision, ScanHistoryResponse } from "../types/api";
import { request } from "../utils/request";

export interface ScanHistoryParams {
  page?: number;
  pageSize?: number;
  decision?: Decision;
}

export function getScanHistory(params: ScanHistoryParams = {}): Promise<ScanHistoryResponse> {
  return request<ScanHistoryResponse>({
    url: "/app/v1/scan/history.json",
    data: {
      page: params.page || 1,
      page_size: params.pageSize || 20,
      decision: params.decision || ""
    }
  });
}

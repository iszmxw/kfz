import { Progress, Tag } from "react-vant"
import type { CollectorStatus, CollectorSummary } from "../lib/types"

interface Props {
  status: CollectorStatus
  summary: CollectorSummary
}

export function StatusPanel({ status, summary }: Props) {
  const percent = status.totalPages ? Math.min(100, Math.round((status.currentPage / status.totalPages) * 100)) : 0
  return (
    <div className="panel">
      <div className="panel-title">
        <span>任务状态</span>
        <Tag type={tagType(status.status)}>{status.status}</Tag>
      </div>
      <Progress percentage={percent} strokeWidth={8} />
      <div className="muted" style={{ marginTop: 8 }}>
        {status.message || "空闲"} {status.totalPages ? ` · ${status.currentPage}/${status.totalPages}` : ""}
      </div>
      <div className="stat-grid">
        <div className="stat">
          <span>本次采集</span>
          <strong>{status.collectedCount}</strong>
        </div>
        <div className="stat">
          <span>有效</span>
          <strong>{status.validCount}</strong>
        </div>
        <div className="stat">
          <span>无效</span>
          <strong>{status.invalidCount}</strong>
        </div>
        <div className="stat">
          <span>本地总量</span>
          <strong>{summary.total}</strong>
        </div>
        <div className="stat">
          <span>待同步</span>
          <strong>{summary.pending + summary.failed}</strong>
        </div>
        <div className="stat">
          <span>已同步</span>
          <strong>{summary.synced}</strong>
        </div>
      </div>
    </div>
  )
}

function tagType(status: CollectorStatus["status"]): "primary" | "success" | "warning" | "danger" | "default" {
  if (status === "completed") {
    return "success"
  }
  if (status === "failed") {
    return "danger"
  }
  if (status === "running") {
    return "primary"
  }
  if (status === "paused" || status === "stopping") {
    return "warning"
  }
  return "default"
}

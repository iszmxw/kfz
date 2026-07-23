import "react-vant/lib/index.css"
import "./style.css"

import { Pause, Play, RotateCcw, Settings, Square } from "lucide-react"
import { useEffect, useState } from "react"
import { Button, Toast } from "react-vant"
import { StatusPanel } from "./components/StatusPanel"
import { loadSettings } from "./lib/settings"
import { sendRuntimeMessage } from "./lib/runtime"
import { MESSAGE_TYPES, type CollectorStatus, type CollectorSummary } from "./lib/types"

const emptyStatus: CollectorStatus = {
  running: false,
  paused: false,
  status: "idle",
  currentPage: 0,
  collectedCount: 0,
  validCount: 0,
  invalidCount: 0,
  message: ""
}

const emptySummary: CollectorSummary = {
  total: 0,
  valid: 0,
  invalid: 0,
  pending: 0,
  synced: 0,
  failed: 0
}

export default function Popup() {
  const [status, setStatus] = useState(emptyStatus)
  const [summary, setSummary] = useState(emptySummary)

  useEffect(() => {
    void refresh()
    const timer = window.setInterval(refresh, 1200)
    return () => window.clearInterval(timer)
  }, [])

  async function refresh() {
    const response = await sendRuntimeMessage<void, { status: CollectorStatus; summary: CollectorSummary }>(MESSAGE_TYPES.GET_STATUS)
    if (response.ok && response.data) {
      setStatus(response.data.status)
      setSummary(response.data.summary)
    }
  }

  async function start() {
    const settings = await loadSettings()
    const response = await sendRuntimeMessage(MESSAGE_TYPES.START_COLLECTION, { config: settings })
    if (!response.ok) {
      Toast.fail(response.error || "启动失败")
      return
    }
    Toast.success("已启动")
    await refresh()
  }

  async function control(type: string, message: string) {
    const response = await sendRuntimeMessage(type)
    if (!response.ok) {
      Toast.fail(response.error || "操作失败")
      return
    }
    Toast.success(message)
    await refresh()
  }

  return (
    <main className="app-shell">
      <div className="topbar">
        <div className="title-block">
          <h1>孔夫子采集</h1>
          <p>小说分类</p>
        </div>
        <Button size="small" icon={<Settings size={16} />} onClick={() => chrome.runtime.openOptionsPage()} />
      </div>
      <div className="content">
        <StatusPanel status={status} summary={summary} />
        <div className="panel">
          <div className="panel-title">快速控制</div>
          <div className="button-row">
            <Button type="primary" disabled={status.running} onClick={start}>
              <span className="icon-button-label"><Play size={15} />开始</span>
            </Button>
            <Button disabled={!status.running || status.paused} onClick={() => control(MESSAGE_TYPES.PAUSE_COLLECTION, "已暂停")}>
              <span className="icon-button-label"><Pause size={15} />暂停</span>
            </Button>
            <Button disabled={!status.paused} onClick={() => control(MESSAGE_TYPES.RESUME_COLLECTION, "继续采集")}>
              <span className="icon-button-label"><RotateCcw size={15} />恢复</span>
            </Button>
            <Button disabled={!status.running && status.status !== "paused"} onClick={() => control(MESSAGE_TYPES.STOP_COLLECTION, "正在停止")}>
              <span className="icon-button-label"><Square size={15} />停止</span>
            </Button>
          </div>
        </div>
      </div>
    </main>
  )
}

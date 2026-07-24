import "react-vant/lib/index.css"
import "./style.css"

import { Pause, Play, RefreshCw, RotateCcw, Save, Square, Upload } from "lucide-react"
import { useEffect, useState } from "react"
import { Button, Field, Tabs, Toast } from "react-vant"
import { StatusPanel } from "./components/StatusPanel"
import { listCollectorItems } from "./lib/db"
import { sendRuntimeMessage } from "./lib/runtime"
import { DEFAULT_SETTINGS, loadSettings, saveSettings, type ExtensionSettings } from "./lib/settings"
import { loginAdmin } from "./lib/syncClient"
import { MESSAGE_TYPES, type CollectorItem, type CollectorStatus, type CollectorSummary } from "./lib/types"

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

export default function OptionsPage() {
  const [settings, setSettings] = useState<ExtensionSettings>(DEFAULT_SETTINGS)
  const [password, setPassword] = useState("")
  const [status, setStatus] = useState(emptyStatus)
  const [summary, setSummary] = useState(emptySummary)
  const [items, setItems] = useState<CollectorItem[]>([])

  useEffect(() => {
    void init()
    const timer = window.setInterval(refreshStatus, 1500)
    return () => window.clearInterval(timer)
  }, [])

  async function init() {
    setSettings(await loadSettings())
    await refreshStatus()
    await refreshItems()
  }

  async function refreshStatus() {
    const response = await sendRuntimeMessage<void, { status: CollectorStatus; summary: CollectorSummary }>(MESSAGE_TYPES.GET_STATUS)
    if (response.ok && response.data) {
      setStatus(response.data.status)
      setSummary(response.data.summary)
    }
  }

  async function refreshItems() {
    setItems(await listCollectorItems(200))
  }

  async function persistSettings(next = settings) {
    await saveSettings(next)
    setSettings(next)
    Toast.success("已保存")
  }

  async function start() {
    await saveSettings(settings)
    const response = await sendRuntimeMessage(MESSAGE_TYPES.START_COLLECTION, { config: settings })
    if (!response.ok) {
      Toast.fail(response.error || "启动失败")
      return
    }
    Toast.success("已启动")
    await refreshStatus()
  }

  async function control(type: string, message: string) {
    const response = await sendRuntimeMessage(type)
    if (!response.ok) {
      Toast.fail(response.error || "操作失败")
      return
    }
    Toast.success(message)
    await refreshStatus()
  }

  async function login() {
    try {
      const token = await loginAdmin(settings.backendBaseUrl, settings.adminUsername, password)
      const next = { ...settings, adminToken: token }
      await saveSettings(next)
      setSettings(next)
      setPassword("")
      Toast.success("已登录")
    } catch (error) {
      Toast.fail(error instanceof Error ? error.message : "登录失败")
    }
  }

  async function sync() {
    if (!settings.adminToken) {
      Toast.fail("请先登录后台")
      return
    }
    const response = await sendRuntimeMessage(MESSAGE_TYPES.SYNC_ITEMS, {
      backendBaseUrl: settings.backendBaseUrl,
      token: settings.adminToken,
      catId: settings.catId,
      limit: 500
    })
    if (!response.ok) {
      Toast.fail(response.error || "同步失败")
      await refreshItems()
      await refreshStatus()
      return
    }
    Toast.success("同步完成")
    await refreshItems()
    await refreshStatus()
  }

  return (
    <main className="options-shell">
      <div className="topbar">
        <div className="title-block">
          <h1>孔夫子分类采集</h1>
          <p>小说分类</p>
        </div>
        <Button icon={<RefreshCw size={16} />} onClick={() => void init()}>
          刷新
        </Button>
      </div>

      <div className="content">
        <StatusPanel status={status} summary={summary} />
        <Tabs>
          <Tabs.TabPane title="任务" name="task">
            <div className="panel">
              <div className="panel-title">采集参数</div>
              <div className="form-grid">
                <Field className="form-grid-wide" label="后端地址" value={settings.backendBaseUrl} onChange={(value) => setSettings({ ...settings, backendBaseUrl: String(value) })} />
                <Field label="分类ID" type="number" value={String(settings.catId)} onChange={(value) => setSettings({ ...settings, catId: toPositiveInt(value, 43) })} />
                <Field label="起始页" type="number" value={String(settings.startPage)} onChange={(value) => setSettings({ ...settings, startPage: toPositiveInt(value, 1) })} />
                <Field label="结束页" type="number" value={settings.endPage ? String(settings.endPage) : ""} onChange={(value) => setSettings({ ...settings, endPage: String(value).trim() ? toPositiveInt(value, 1) : undefined })} />
                <Field label="限速ms" type="number" value={String(settings.delayMs)} onChange={(value) => setSettings({ ...settings, delayMs: toPositiveInt(value, 1200) })} />
                <Field label="重试" type="number" value={String(settings.retryTimes)} onChange={(value) => setSettings({ ...settings, retryTimes: toPositiveInt(value, 2) })} />
              </div>
              <div className="button-row">
                <Button type="primary" icon={<Save size={15} />} onClick={() => void persistSettings()}>
                  保存
                </Button>
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
          </Tabs.TabPane>

          <Tabs.TabPane title="同步" name="sync">
            <div className="panel">
              <div className="panel-title">后台账号</div>
              <div className="form-grid">
                <Field label="用户名" value={settings.adminUsername} onChange={(value) => setSettings({ ...settings, adminUsername: String(value) })} />
                <Field label="密码" type="password" value={password} onChange={(value) => setPassword(String(value))} />
              </div>
              <div className="button-row">
                <Button type="primary" onClick={login}>登录</Button>
                <Button type="primary" icon={<Upload size={15} />} onClick={sync}>同步待处理</Button>
              </div>
            </div>
          </Tabs.TabPane>

          <Tabs.TabPane title="数据" name="items">
            <div className="panel">
              <div className="panel-title">
                <span>本地数据</span>
                <Button size="small" onClick={refreshItems}>刷新列表</Button>
              </div>
              <div className="scroll-table">
                <table className="items-table">
                  <thead>
                    <tr>
                      <th>图书</th>
                      <th>ISBN</th>
                      <th>旧书低价</th>
                      <th>在售</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    {items.map((item) => (
                      <tr key={item.id}>
                        <td>
                          <div className="book-cell">
                            {item.coverUrl ? <img src={item.coverUrl} alt="" /> : <span />}
                            <div>
                              <div className="book-title">{item.title || "无标题"}</div>
                              <div className="muted">{[item.author, item.publisher, item.publishYear].filter(Boolean).join(" · ")}</div>
                              {!item.valid ? <div className="muted">{item.errorMessage}</div> : null}
                            </div>
                          </div>
                        </td>
                        <td>{item.normalizedIsbn || item.isbn}</td>
                        <td>{item.oldBookMinPriceText || item.oldBookMinPrice || "-"}</td>
                        <td>{item.oldBookOnSaleNum}</td>
                        <td>{item.valid ? item.syncStatus : "invalid"}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </Tabs.TabPane>
        </Tabs>
      </div>
    </main>
  )
}

function toPositiveInt(value: unknown, fallback: number): number {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return fallback
  }
  return Math.floor(parsed)
}

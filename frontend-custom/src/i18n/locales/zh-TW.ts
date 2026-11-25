import common from './zh-TW/common.ts'
import auth from './zh-TW/auth.ts'
import chat from './zh-TW/chat.ts'
import agent from './zh-TW/agent.ts'
import knowledge from './zh-TW/knowledge.ts'
import datasource from './zh-TW/datasource.ts'
import model from './zh-TW/model.ts'
import settings from './zh-TW/settings.ts'
import system from './zh-TW/system.ts'

export default {
  ...common,
  ...auth,
  ...chat,
  ...agent,
  ...knowledge,
  ...datasource,
  ...model,
  ...settings,
  ...system,
  "knowledgeBase": {
    ...knowledge["knowledgeBase"],
    "processingOverview": {
      "title": "全庫處理進度",
      "total": "共 {count} 份文件（含所有資料夾）",
      "complete": "全流程完成",
      "parse": "文件解析完成",
      "summary": "摘要產生完成",
      "summaryDetails": "摘要待處理 {pending} · 產生中 {processing} · 失敗 {failed}",
      "optimize": "後處理完成",
      "details": "待解析 {pending} · 解析中 {processing} · 最後處理 {finalizing} · 失敗 {failed} · 已取消 {cancelled}",
      "denominators": "摘要僅統計已安排摘要任務的文件；後處理僅統計已解析文件。",
      "refresh": "重新整理進度",
      "loadFailed": "暫時無法取得全庫進度。",
      "pdfIntegrity": "PDF 全頁檢查：已通過 {verified} · 歷史未檢查 {unverified} · 待檢查 {pending} · 失敗 {failed}",
      "failureList": "檢視失敗文件與原因",
      "failureLoading": "正在載入失敗記錄…",
      "failureLoadFailed": "失敗記錄載入失敗，請重試。",
      "pageMismatch": "原始 PDF {expected} 頁，實際傳回 {actual} 頁"
    },
  },
  "model": {
    ...model["model"],
    "editor": {
      ...model["model"]["editor"],
      "contextWindowDesc": "該模型一次請求能容納的 token 數。智慧代理壓縮對話歷史會按此上限工作。留空則使用預設 {value}（{label}）。請按廠商文件填寫真實值，填大會導致壓縮不觸發、上遊直接拒絕請求。",
    },
  },
}

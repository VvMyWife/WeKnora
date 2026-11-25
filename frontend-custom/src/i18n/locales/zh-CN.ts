import common from './zh-CN/common.ts'
import auth from './zh-CN/auth.ts'
import chat from './zh-CN/chat.ts'
import agent from './zh-CN/agent.ts'
import knowledge from './zh-CN/knowledge.ts'
import datasource from './zh-CN/datasource.ts'
import model from './zh-CN/model.ts'
import settings from './zh-CN/settings.ts'
import system from './zh-CN/system.ts'

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
      "title": "全库处理进度",
      "total": "共 {count} 份文档（含所有文件夹）",
      "complete": "全流程完成",
      "parse": "文档解析完成",
      "summary": "摘要生成完成",
      "summaryDetails": "摘要待处理 {pending} · 生成中 {processing} · 失败 {failed}",
      "optimize": "后处理完成",
      "details": "待解析 {pending} · 解析中 {processing} · 优化中 {finalizing} · 失败 {failed} · 已取消 {cancelled}",
      "denominators": "摘要仅统计已安排摘要任务的文档；后处理仅统计已解析文档。",
      "refresh": "刷新进度",
      "loadFailed": "全库进度暂时无法获取。",
      "pdfIntegrity": "PDF 全页校验：已通过 {verified} · 历史未校验 {unverified} · 待校验 {pending} · 失败 {failed}",
      "failureList": "查看失败文档与原因",
      "failureLoading": "正在加载失败记录…",
      "failureLoadFailed": "失败记录加载失败，请重试。",
      "pageMismatch": "源 PDF {expected} 页，实际返回 {actual} 页"
    },
    "pdfDocFormat": "PDF、DOCX 等文档，单文件不超过 {size} MB",
    "textMarkdownFormat": "TXT、Markdown 等文本，单文件不超过 {size} MB",
  },
  "model": {
    ...model["model"],
    "editor": {
      ...model["model"]["editor"],
      "contextWindowDesc": "该模型一次请求能容纳的 token 数。智能体压缩对话历史会按此上限工作。留空则使用默认 {value}（{label}）。请按厂商文档填写真实值，填大会导致压缩不触发、上游直接拒绝请求。",
    },
  },
  "createChat": {
    ...chat["createChat"],
    "title": "档案智查，让每份档案都可检索、可追溯",
  },
  "auth": {
    ...auth["auth"],
    "firstTime": "首次使用档案智查？",
    "registerSubtitle": "创建账户并开始使用档案智查",
  },
  "newUserGuide": {
    ...common["newUserGuide"],
    "steps": {
      ...common["newUserGuide"]["steps"],
      "knowledge": {
        ...common["newUserGuide"]["steps"]["knowledge"],
        "desc": "知识库是一切的起点：上传档案、网页或 FAQ，系统会自动解析并建立索引。点击这里进入知识库。",
      },
      "welcome": {
        ...common["newUserGuide"]["steps"]["welcome"],
        "title": "欢迎使用档案智查",
      },
    },
  },
}

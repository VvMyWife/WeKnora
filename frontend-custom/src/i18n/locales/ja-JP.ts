import common from './ja-JP/common.ts'
import auth from './ja-JP/auth.ts'
import chat from './ja-JP/chat.ts'
import agent from './ja-JP/agent.ts'
import knowledge from './ja-JP/knowledge.ts'
import datasource from './ja-JP/datasource.ts'
import model from './ja-JP/model.ts'
import settings from './ja-JP/settings.ts'
import system from './ja-JP/system.ts'

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
      "title": "ナレッジベースの処理状況",
      "total": "全フォルダーの文書数：{count}",
      "complete": "全処理完了",
      "parse": "解析完了",
      "summary": "要約完了",
      "summaryDetails": "要約待機 {pending} · 生成中 {processing} · 失敗 {failed}",
      "optimize": "後処理完了",
      "details": "待機 {pending} · 解析中 {processing} · 最終処理 {finalizing} · 失敗 {failed} · キャンセル {cancelled}",
      "denominators": "要約は予定されたタスクのみ、後処理は解析済み文書のみを集計します。",
      "refresh": "状況を更新",
      "loadFailed": "処理状況を取得できません。",
      "pdfIntegrity": "PDFページ検証：確認済み {verified} · 過去の未確認 {unverified} · 待機 {pending} · 失敗 {failed}",
      "failureList": "失敗した文書と原因を表示",
      "failureLoading": "失敗記録を読み込み中…",
      "failureLoadFailed": "失敗記録を読み込めません。再試行してください。",
      "pageMismatch": "元のPDFは{expected}ページ、サービスの結果は{actual}ページ"
    },
  },
  "model": {
    ...model["model"],
    "editor": {
      ...model["model"]["editor"],
      "contextWindowDesc": "1回のリクエストでこのモデルが受け付けられるトークン数です。エージェントの履歴圧縮はこの上限を基準にします。空欄の場合はデフォルト値の{value}（{label}）が使われます。プロバイダの実際のウィンドウサイズを指定してください。大きすぎる値を指定すると圧縮が働かず、プロバイダにリクエストを拒否されます。",
    },
  },
}

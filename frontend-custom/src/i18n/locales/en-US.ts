import common from './en-US/common.ts'
import auth from './en-US/auth.ts'
import chat from './en-US/chat.ts'
import agent from './en-US/agent.ts'
import knowledge from './en-US/knowledge.ts'
import datasource from './en-US/datasource.ts'
import model from './en-US/model.ts'
import settings from './en-US/settings.ts'
import system from './en-US/system.ts'

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
      "title": "Knowledge base progress",
      "total": "{count} documents across all folders",
      "complete": "Fully processed",
      "parse": "Parsing complete",
      "summary": "Summaries complete",
      "summaryDetails": "Summary queued {pending} · Generating {processing} · Failed {failed}",
      "optimize": "Post-processing complete",
      "details": "Queued {pending} · Parsing {processing} · Finalizing {finalizing} · Failed {failed} · Cancelled {cancelled}",
      "denominators": "Summaries count scheduled jobs only; post-processing counts parsed documents only.",
      "refresh": "Refresh progress",
      "loadFailed": "Knowledge base progress is temporarily unavailable.",
      "pdfIntegrity": "PDF page check: verified {verified} · historical unverified {unverified} · pending {pending} · failed {failed}",
      "failureList": "View failed documents and reasons",
      "failureLoading": "Loading failures…",
      "failureLoadFailed": "Could not load failures. Please retry.",
      "pageMismatch": "Source PDF {expected} pages, service returned {actual}"
    },
    "pdfDocFormat": "PDF, DOCX and other documents, max {size} MB per file",
    "textMarkdownFormat": "TXT, Markdown and other text files, max {size} MB per file",
  },
  "model": {
    ...model["model"],
    "editor": {
      ...model["model"]["editor"],
      "contextWindowDesc": "How many tokens this model can take in one request. Agent history compaction uses this limit. Leave empty for the default {value} ({label}). Use the provider’s real window — a larger guess means compaction never fires and the provider rejects the request.",
    },
  },
}

import common from './ko-KR/common.ts'
import auth from './ko-KR/auth.ts'
import chat from './ko-KR/chat.ts'
import agent from './ko-KR/agent.ts'
import knowledge from './ko-KR/knowledge.ts'
import datasource from './ko-KR/datasource.ts'
import model from './ko-KR/model.ts'
import settings from './ko-KR/settings.ts'
import system from './ko-KR/system.ts'

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
      "title": "지식베이스 처리 진행률",
      "total": "모든 폴더의 문서 {count}개",
      "complete": "전체 처리 완료",
      "parse": "파싱 완료",
      "summary": "요약 완료",
      "summaryDetails": "요약 대기 {pending} · 생성 중 {processing} · 실패 {failed}",
      "optimize": "후처리 완료",
      "details": "대기 {pending} · 파싱 {processing} · 마무리 {finalizing} · 실패 {failed} · 취소 {cancelled}",
      "denominators": "요약은 예약된 작업만, 후처리는 파싱된 문서만 집계합니다.",
      "refresh": "진행률 새로 고침",
      "loadFailed": "지식베이스 진행률을 불러올 수 없습니다.",
      "pdfIntegrity": "PDF 페이지 검사: 확인 {verified} · 이전 미확인 {unverified} · 대기 {pending} · 실패 {failed}",
      "failureList": "실패 문서와 원인 보기",
      "failureLoading": "실패 기록을 불러오는 중…",
      "failureLoadFailed": "실패 기록을 불러올 수 없습니다. 다시 시도하세요.",
      "pageMismatch": "원본 PDF {expected}페이지, 서비스 반환 {actual}페이지"
    },
    "pdfDocFormat": "PDF, DOCX 등 문서 파일당 최대 {size}MB",
    "textMarkdownFormat": "TXT, Markdown 등 텍스트 파일당 최대 {size}MB",
  },
  "model": {
    ...model["model"],
    "editor": {
      ...model["model"]["editor"],
      "contextWindowDesc": "모델이 한 요청에 수용할 수 있는 토큰 수입니다. 에이전트 대화 압축이 이 한도를 사용합니다. 비워 두면 기본값 {value}({label})을 사용합니다. 공급자 문서의 실제 값을 입력하세요. 더 크게 설정하면 압축이 발생하지 않고 요청이 거부될 수 있습니다.",
    },
  },
}

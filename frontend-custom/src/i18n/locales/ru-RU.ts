import common from './ru-RU/common.ts'
import auth from './ru-RU/auth.ts'
import chat from './ru-RU/chat.ts'
import agent from './ru-RU/agent.ts'
import knowledge from './ru-RU/knowledge.ts'
import datasource from './ru-RU/datasource.ts'
import model from './ru-RU/model.ts'
import settings from './ru-RU/settings.ts'
import system from './ru-RU/system.ts'

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
      "title": "Ход обработки базы знаний",
      "total": "Документов во всех папках: {count}",
      "complete": "Полностью обработано",
      "parse": "Разбор завершён",
      "summary": "Сводки готовы",
      "summaryDetails": "Сводки: в очереди {pending} · создаются {processing} · ошибки {failed}",
      "optimize": "Постобработка завершена",
      "details": "В очереди {pending} · разбор {processing} · завершение {finalizing} · ошибки {failed} · отменено {cancelled}",
      "denominators": "Сводки учитывают только запланированные задания, постобработка — только разобранные документы.",
      "refresh": "Обновить прогресс",
      "loadFailed": "Прогресс базы знаний временно недоступен.",
      "pdfIntegrity": "Проверка страниц PDF: подтверждено {verified} · ранее не проверено {unverified} · ожидают {pending} · ошибки {failed}",
      "failureList": "Показать документы с ошибками и причины",
      "failureLoading": "Загрузка записей об ошибках…",
      "failureLoadFailed": "Не удалось загрузить ошибки. Повторите попытку.",
      "pageMismatch": "В исходном PDF {expected} страниц, сервис вернул {actual}"
    },
    "pdfDocFormat": "PDF, DOCX и другие документы: не более {size} МБ на файл",
    "textMarkdownFormat": "TXT, Markdown и другие текстовые файлы: не более {size} МБ на файл",
  },
  "model": {
    ...model["model"],
    "editor": {
      ...model["model"]["editor"],
      "contextWindowDesc": "Сколько токенов модель принимает за один запрос. Сжатие истории агента использует этот лимит. Пустое значение — по умолчанию {value} ({label}). Укажите реальное окно провайдера: завышенное значение не запускает сжатие, и провайдер отклоняет запрос.",
    },
  },
}

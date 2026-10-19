export interface GraphDocument {
  id: string
  title?: string
  file_name?: string
  folder_path?: string
  archive_source?: { file_name?: string; relative_path?: string }
}

export function documentGraphLabel(doc: GraphDocument): string {
  return doc.archive_source?.file_name || doc.file_name || doc.title || doc.id
}

export function documentGraphPath(doc: GraphDocument): string {
  return doc.archive_source?.relative_path || [doc.folder_path, doc.file_name || doc.title].filter(Boolean).join('/')
}

export function documentScopeParams(scope: 'kb' | 'documents', docs: GraphDocument[], view: 'union' | 'connections') {
  if (scope === 'kb') return { scope: 'kb' as const }
  const ids = [...new Set(docs.map(doc => doc.id))]
  if (!ids.length || ids.length > 50) throw new Error('请选择 1 至 50 份档案')
  return { scope: 'documents' as const, knowledge_ids: ids, view }
}

// Both KB identity and selected scope are part of request identity. Old network
// responses must never paint a new KB's graph or overwrite a new selection.
export function documentGraphRequestKey(kb: string, scope: string, ids: string[], view: string, types: string[]) {
  return JSON.stringify([kb, scope, [...ids].sort(), view, [...types].sort()])
}

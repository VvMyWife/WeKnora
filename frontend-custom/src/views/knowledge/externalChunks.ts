export interface ImportRecord {
  id: string; external_document_id: string; knowledge_id: string; relative_path: string;
  status: string; stage: string; error: string;
}
export function importCounts(rows: ImportRecord[]) {
  return {
    total: rows.length,
    completed: rows.filter(r => r.status === 'completed').length,
    failed: rows.filter(r => r.status === 'failed').length,
    waiting: rows.filter(r => r.status === 'preview').length,
    active: rows.filter(r => ['preparing','queued','pending','processing','finalizing','postprocessing'].includes(r.status)).length,
  };
}

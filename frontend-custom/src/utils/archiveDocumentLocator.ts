import type { SourceLocateRequest } from './sourceLocator';

// The OCR page ordinal is deterministic in the archived PDF. It establishes
// an exact page mapping, without asserting an unverified text box.
export function archiveDocumentLocate(pageIndex: unknown, content: string, token: number): SourceLocateRequest {
  const valid = typeof pageIndex === 'number' && Number.isInteger(pageIndex) && pageIndex >= 0;
  return {
    locators: valid ? [{ type: 'pdf', page: pageIndex + 1, mapping: 'exact' }] : [],
    quotes: [], token, sourceMarkdown: content || '', unavailable: !valid,
  };
}

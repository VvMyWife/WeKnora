import { test } from 'node:test';
import assert from 'node:assert/strict';
import { archiveDocumentLocate } from './archiveDocumentLocator';
import { resolvePdfSource } from './pdfSourceLocate';

test('an archived scanned page locates page 2 without a PDF text layer', async () => {
  const result = await resolvePdfSource(archiveDocumentLocate(1, 'OCR text', 1), 45, async () => '');
  assert.deepEqual(result.targets, [{ page: 2, granularity: 'page' }]);
});
test('missing or invalid OCR page never invents a source page', () => {
  for (const index of [undefined, -1, 1.5, '2']) {
    const result = archiveDocumentLocate(index, 'OCR text', 1);
    assert.equal(result.unavailable, true);
    assert.deepEqual(result.locators, []);
  }
});

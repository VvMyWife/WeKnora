import assert from 'node:assert/strict';
import test from 'node:test';
import { buildProcessingProgress } from './processingOverview';

test('whole-KB progress distinguishes parsing, summary and finalization', () => {
  const result = buildProcessingProgress({
    total: 10,
    parse: { pending: 1, processing: 1, finalizing: 3, completed: 4, failed: 1 },
    summary: { none: 4, pending: 1, processing: 1, completed: 3, failed: 1 },
  });
  assert.equal(result.overallPercent, 40);
  assert.equal(result.parsePercent, 70);
  assert.equal(result.summaryPercent, 50);
  assert.equal(result.postProcessPercent, 57);
  assert.equal(result.active, true);
  assert.equal(result.summaryEligible, 6);
});

test('empty and skipped-summary KBs do not claim 100 percent summary progress', () => {
  const result = buildProcessingProgress({ total: 2, parse: { completed: 2 }, summary: { none: 2 } });
  assert.equal(result.summaryEligible, 0);
  assert.equal(result.summaryPercent, 0);
  assert.equal(result.active, false);
});

import type { KnowledgeProcessingOverview } from '@/api/knowledge-base/index';

export function buildProcessingProgress(overview: KnowledgeProcessingOverview) {
  const count = (values: Record<string, number>, key: string) => Math.max(0, values[key] || 0);
  const total = Math.max(0, overview.total || 0);
  const parse = overview.parse || {};
  const summary = overview.summary || {};
  const completed = count(parse, 'completed');
  const finalizing = count(parse, 'finalizing');
  const parsed = completed + finalizing;
  const summaryEligible = ['pending', 'processing', 'completed', 'failed']
    .reduce((sum, status) => sum + count(summary, status), 0);
  const percent = (done: number, denominator: number) =>
    denominator > 0 ? Math.min(100, Math.round(100 * done / denominator)) : 0;
  return {
    total,
    completed,
    pending: count(parse, 'pending'),
    processing: count(parse, 'processing'),
    finalizing,
    failed: count(parse, 'failed'),
    cancelled: count(parse, 'cancelled'),
    overallPercent: percent(completed, total),
    parsed,
    parsePercent: percent(parsed, total),
    summaryEligible,
    summaryCompleted: count(summary, 'completed'),
    summaryPending: count(summary, 'pending'),
    summaryProcessing: count(summary, 'processing'),
    summaryFailed: count(summary, 'failed'),
    summaryPercent: percent(count(summary, 'completed'), summaryEligible),
    postProcessPercent: percent(completed, parsed),
    active: count(parse, 'pending') + count(parse, 'processing') + finalizing > 0,
  };
}

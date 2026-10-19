import type { KnowledgeProcessingControl } from '@/api/knowledge-base/index';

export function processingControlNotice(control: KnowledgeProcessingControl | null): string {
  if (!control?.supported) return '';
  if (control.state === 'pausing') {
    return `正在暂停：等待 ${control.active_tasks} 个已启动任务完成；尚未启动的任务已暂停，其他知识库可继续处理。请等显示“已暂停”后再修改服务参数。`;
  }
  if (control.state === 'paused') {
    return '已暂停当前知识库后台处理，已有结果和排队任务均保留。可修改参数后点击“继续处理”；排队任务通常在恢复后约 30～45 秒内继续。';
  }
  return '';
}

import assert from 'node:assert/strict';
import test from 'node:test';
import { processingControlNotice } from './processingControl';

test('draining distinguishes live tasks from a completed pause', () => {
  const message = processingControlNotice({ supported:true, requested_paused:true, state:'pausing', active_tasks:3 });
  assert.match(message, /3 个已启动任务/);
  assert.match(message, /其他知识库可继续/);
  assert.match(message, /请等显示/);
});

test('paused retains queue and describes resume delay', () => {
  const message = processingControlNotice({ supported:true, requested_paused:true, state:'paused', active_tasks:0 });
  assert.match(message, /^已暂停/);
  assert.match(message, /排队任务均保留/);
  assert.match(message, /30～45 秒/);
});

test('running and unsupported states do not claim processing is paused', () => {
  assert.equal(processingControlNotice(null), '');
  assert.equal(processingControlNotice({ supported:true, requested_paused:false, state:'running', active_tasks:2 }), '');
  assert.equal(processingControlNotice({ supported:false, requested_paused:false, state:'unsupported', active_tasks:0 }), '');
});

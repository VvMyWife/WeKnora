import { test } from 'node:test';
import assert from 'node:assert/strict';
import { importCounts, type ImportRecord } from './externalChunks';
test('external chunks: queued/finalizing are not completion', () => {
    const rows = ['preview','queued','finalizing','failed','completed'].map((status,id) => ({id:String(id),status}) as ImportRecord);
    assert.deepEqual(importCounts(rows),{total:5,completed:1,failed:1,waiting:1,active:2});
});

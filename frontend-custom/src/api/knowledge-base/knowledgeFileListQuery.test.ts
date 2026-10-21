import assert from 'node:assert/strict';
import test from 'node:test';

import { buildListKnowledgeFilesQuery } from './knowledgeFileListQuery';

test('知识文件列表请求会传递排序字段和方向', () => {
  const query = new URLSearchParams(buildListKnowledgeFilesQuery({
    page: 2,
    page_size: 35,
    sort_by: 'file_name',
    sort_order: 'asc',
    folder_path: '',
  }));

  assert.equal(query.get('page'), '2');
  assert.equal(query.get('page_size'), '35');
  assert.equal(query.get('sort_by'), 'file_name');
  assert.equal(query.get('sort_order'), 'asc');
  assert.equal(query.get('folder_path'), '');
});

test('原件映射信息与官方排序和目录过滤同时保留', () => {
  const query = new URLSearchParams(buildListKnowledgeFilesQuery({
    page: 1,
    page_size: 20,
    sort_by: 'updated_at',
    sort_order: 'desc',
    folder_path: 'archive',
    folder_recursive: true,
    include_archive_source: true,
  }));
  assert.equal(query.get('include_archive_source'), 'true');
  assert.equal(query.get('sort_by'), 'updated_at');
  assert.equal(query.get('folder_path'), 'archive');
  assert.equal(query.get('folder_recursive'), 'true');
});

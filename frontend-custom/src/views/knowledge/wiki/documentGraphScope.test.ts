import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { documentGraphLabel, documentGraphPath, documentScopeParams, documentGraphRequestKey } from './documentGraphScope'

test('document selection uses IDs and mapped original labels', () => {
  const doc = { id: 'a', file_name: 'parsed.md', folder_path: 'folder', archive_source: { file_name: 'original.pdf', relative_path: 'original/folder.pdf' } }
  assert.equal(documentGraphLabel(doc), 'original.pdf')
  assert.equal(documentGraphPath(doc), 'original/folder.pdf')
  assert.deepEqual(documentScopeParams('documents', [doc, doc, {id:'b',file_name:'original.pdf'}], 'connections').knowledge_ids, ['a','b'])
  assert.deepEqual(documentScopeParams('kb', [doc], 'connections'), {scope:'kb'})
  assert.throws(() => documentScopeParams('documents', [], 'union'))
  assert.throws(() => documentScopeParams('documents', Array.from({length:51},(_,i)=>({id:String(i)})), 'union'))
})
test('request keys guard KB, scope, type and selection changes', () => {
  const key = documentGraphRequestKey('kb','documents',['b','a'],'union',['entity'])
  assert.equal(key, documentGraphRequestKey('kb','documents',['a','b'],'union',['entity']))
  assert.notEqual(key, documentGraphRequestKey('new','documents',['a','b'],'union',['entity']))
  assert.notEqual(key, documentGraphRequestKey('kb','documents',['a'],'union',['entity']))
  assert.notEqual(key, documentGraphRequestKey('kb','documents',['a','b'],'connections',['entity']))
  assert.notEqual(key, documentGraphRequestKey('kb','documents',['a','b'],'union',['summary']))
})
test('document picker escapes the graph empty-state stacking context', () => {
  const selector = readFileSync(new URL('./DocumentGraphSelector.vue', import.meta.url), 'utf8')
  assert.match(selector, /<Teleport to="body">/)
  const toolbarZ = Number(selector.match(/\.document-graph-toolbar\s*\{[^}]*z-index:(\d+)/)?.[1])
  const browser = readFileSync(new URL('./WikiBrowser.vue', import.meta.url), 'utf8')
  const emptyZ = Number(browser.match(/\.wiki-graph-empty\s*\{[^}]*z-index:\s*(\d+)/)?.[1])
  assert.ok(toolbarZ > emptyZ, 'empty state must not cover the scope controls')
})

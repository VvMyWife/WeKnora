import assert from 'node:assert/strict'
import test from 'node:test'
import { relativeImageSourceHtml, knowledgeSourcePreviewID } from './relativeImageSource.ts'
import { createChatMarkdownRenderer, renderChatMarkdown } from './chatMarkdownRenderer.ts'

const kid = '221babfa-9c83-4c66-b845-7ec07e72dc06'
const kb = '8575ca20-5e7e-48c4-9fc0-786d979fdf5d'
const ref = { knowledge_id: kid, knowledge_base_id: kb, content: '![](images/page_9_image_1.jpg)' }

test('unique exact evidence source links to the existing mapped-original preview', () => {
  const html = relativeImageSourceHtml('images/page_9_image_1.jpg', [ref, ref])!
  assert.ok(html.includes(`/platform/knowledge-bases/${kb}?knowledge_id=${kid}`))
  assert.ok(!html.includes('<img'))
})
test('never guess across documents or match a basename from another directory', () => {
  assert.ok(!relativeImageSourceHtml('images/page_9_image_1.jpg', [ref, {...ref, knowledge_id: kb}])!.includes('<a'))
  assert.ok(!relativeImageSourceHtml('other/page_9_image_1.jpg', [ref])!.includes('<a'))
})
test('existing provider, absolute and artifact URLs remain unchanged; unsafe paths never mapped', () => {
  for (const path of ['https://example.org/a.jpg', '/api/file/a.jpg', 'resource://image', 'artifact://x', '../x.jpg', '%2e%2e/x.jpg', '//host/x.jpg']) {
    assert.equal(relativeImageSourceHtml(path, [ref]), null)
  }
})
test('unmatched path is escaped and not left as a broken browser-relative image', () => {
  const html = relativeImageSourceHtml('images/<bad>.jpg')!
  assert.ok(html.includes('&lt;bad&gt;'))
  assert.ok(!html.includes('<img'))
})
test('both standard and agent renderers share the repair, code examples stay literal', () => {
  for (const renderer of [createChatMarkdownRenderer(), createChatMarkdownRenderer({imageRenderer: ({href}) => `<img src="${href}">`})]) {
    const render = (text:string) => renderChatMarkdown(text, {renderer, escapeMarkdown: x=>x, sanitizeHtml:x=>x, knowledgeReferences:[ref]})
    assert.ok(render('![](images/page_9_image_1.jpg)').includes(`knowledge_id=${kid}`))
    assert.ok(!render('`![](images/page_9_image_1.jpg)`').includes(`knowledge_id=${kid}`))
    assert.ok(render('![](https://example.org/a.jpg)').includes('<img'))
  }
})

test('model-mislabelled sandbox image uses uniquely evidenced source, not a missing artifact card', () => {
  const html = relativeImageSourceHtml('sandbox:page_9_image_1.jpg', [ref])!
  assert.ok(html.includes(`knowledge_id=${kid}`))
  assert.ok(html.includes('预览对应原件'))
  assert.ok(!relativeImageSourceHtml('sandbox:page_9_image_1.jpg', [ref, {...ref, knowledge_id:kb}])!.includes('<a'))
})
test('real sandbox artifacts retain precedence and unrelated output is not misattributed', () => {
  assert.equal(relativeImageSourceHtml('sandbox:page_9_image_1.jpg', [ref], [{index:0,file_name:'page_9_image_1.jpg'}]), null)
  assert.ok(!relativeImageSourceHtml('sandbox:unknown.jpg', [ref])!.includes('<a'))
  assert.equal(relativeImageSourceHtml('sandbox:report.xlsx', [ref]), null)
})
test('sandbox repair is applied before image renderer but genuine artifacts still reach it', () => {
  const renderer = createChatMarkdownRenderer({imageRenderer: () => '<b>real-artifact</b>'})
  const render = (artifacts: {index:number;file_name:string}[]) => renderChatMarkdown('![](sandbox:page_9_image_1.jpg)',
    {renderer,escapeMarkdown:x=>x,sanitizeHtml:x=>x,knowledgeReferences:[ref],artifacts})
  assert.ok(render([]).includes(`knowledge_id=${kid}`))
  assert.ok(render([{index:0,file_name:'page_9_image_1.jpg'}]).includes('real-artifact'))
})
test('drawer route parser rejects external, malformed and unscoped destinations', () => {
  const origin='http://10.20.0.15:18180'
  const url=`/platform/knowledge-bases/${kb}?knowledge_id=${kid}`
  assert.equal(knowledgeSourcePreviewID(url,origin),kid)
  assert.equal(knowledgeSourcePreviewID('https://other.example'+url,origin),null)
  assert.equal(knowledgeSourcePreviewID('/platform/knowledge-bases/no?knowledge_id='+kid,origin),null)
})

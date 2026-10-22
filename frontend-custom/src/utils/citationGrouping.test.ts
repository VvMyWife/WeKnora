import assert from 'node:assert/strict'
import test from 'node:test'
import { DOMParser, XMLSerializer } from '@xmldom/xmldom'
import { groupCitationTree, hideUnboundCitationTree } from './citationGrouping.ts'

const ref = (id='a') => `<span class="citation citation-archival" data-citation-id="${id.repeat(64)}">[${id==='a'?1:2}]</span>`
function grouped(html: string, streaming=false) {
  const doc = new DOMParser().parseFromString(`<div>${html}</div>`, 'text/xml')
  groupCitationTree(doc.documentElement as unknown as Element, streaming)
  return new XMLSerializer().serializeToString(doc)
}
const count = (html:string) => (html.match(/class="citation citation-archival"/g)||[]).length

function displayed(html:string, consolidate=true) {
  const doc=new DOMParser().parseFromString(`<div>${html}</div>`,'text/xml')
  const root=doc.documentElement as unknown as Element
  if(consolidate)groupCitationTree(root)
  hideUnboundCitationTree(root)
  return new XMLSerializer().serializeToString(doc)
}
test('hide current and historical missing-binding notices without changing saved text or valid refs',()=>{
 const source=`<p>引言<span class="citation-unbound" title="missing">〔引用待核验〕</span></p><p>事实${ref()}</p><p>过渡<span class="citation-unbound">〔未绑定来源〕</span></p>`
 for(const consolidate of [true,false]){
  const out=displayed(source,consolidate)
  assert.doesNotMatch(out,/citation-unbound|引用待核验|未绑定来源/)
  assert.match(out,/引言/);assert.match(out,/过渡/);assert.equal(count(out),1)
 }
 assert.match(source,/引用待核验/)
})
test('hidden notices still prevent false same-source consolidation across unsupported text',()=>{
 const out=displayed(`<p>甲${ref()}过渡<span class="citation-unbound">〔引用待核验〕</span>乙${ref()}</p><ul><li>丙${ref()}</li><li>过渡<span class="citation-unbound">〔未绑定来源〕</span></li><li>丁${ref()}</li></ul>`)
 assert.equal(count(out),4)
 assert.doesNotMatch(out,/citation-unbound|data-citation-group-size/)
})
test('hide only UI notice spans, not prose, escaped code, other warnings or code examples',()=>{
 const out=displayed('<p>正文提及引用待核验</p><code>&lt;span class="citation-unbound"&gt;例子&lt;/span&gt;</code><pre><span class="citation-unbound">代码例子</span></pre><div class="archival-warning">来源不可访问</div>')
 assert.match(out,/正文提及引用待核验/);assert.match(out,/代码例子/)
 assert.match(out,/&lt;span/);assert.match(out,/来源不可访问/)
})

test('same-source sentences display a single final ref; canonical input stays intact', () => {
  const source = `<p>甲。${ref()}乙。${ref()}丙。${ref()}</p>`
  const out = grouped(source)
  assert.equal(count(out),1)
  assert.match(out,/本段以上 3 处连续同源内容/)
  assert.match(out,/甲。乙。丙。<span/)
  assert.equal(count(source),3)
})
test('source change divides runs, never merges chunks by PDF name', () => {
  const out=grouped(`<p>甲${ref()}乙${ref()}丙${ref('b')}丁${ref('b')}戊${ref()}</p>`)
  assert.equal(count(out),3)
})
test('same-level list items share last ref; missing item and nested level break grouping', () => {
  const out=grouped(`<ul><li>甲${ref()}</li><li>乙${ref()}</li><li>缺引</li><li>丙${ref()}</li><li>丁${ref()}</li></ul>`)
  assert.equal(count(out),2)
  assert.match(out,/以上 2 个连续同源条目/)
  assert.equal(count(grouped(`<ul><li>甲${ref()}<ul><li>子${ref()}</li></ul></li><li>乙${ref()}</li></ul>`)),3)
})
test('headings, paragraphs and tables never merge their sources across boundaries', () => {
  assert.equal(count(grouped(`<p>甲${ref()}</p><h2>标题</h2><p>乙${ref()}</p>`)),2)
  assert.equal(count(grouped(`<table><tbody><tr><td>甲${ref()}</td></tr><tr><td>乙${ref()}</td></tr></tbody></table>`)),2)
})
test('mixed source sets, uncited tails, code and review markers stay conservative', () => {
  assert.equal(count(grouped(`<p>甲${ref()} ${ref('b')}乙${ref()}</p>`)),3)
  assert.equal(count(grouped(`<p>甲${ref()}乙${ref()}未绑定事实</p>`)),2)
  assert.equal(count(grouped(`<p>甲${ref()}<span class="citation-unbound">待核验</span>乙${ref()}</p>`)),2)
  assert.equal(count(grouped(`<p>甲${ref()}<code>example</code>乙${ref()}</p>`)),2)
})
test('matching multisource sets merge without losing either last link', () => {
  const out=grouped(`<ul><li>甲${ref()} ${ref('b')}</li><li>乙${ref('b')} ${ref()}</li></ul>`)
  assert.equal(count(out),2)
  assert.match(out,/对应以上 2 个连续同源条目/)
})
test('streaming keeps last branch unchanged; finished render and reload are identical', () => {
  const closed=`<p>甲${ref()}乙${ref()}</p>`
  const open=`<ul><li>丙${ref()}</li><li>丁${ref()}</li></ul>`
  assert.equal(count(grouped(closed+open,true)),3)
  assert.equal(count(grouped(closed+open,false)),2)
  assert.equal(grouped(closed+open,false),grouped(closed+open,false))
})

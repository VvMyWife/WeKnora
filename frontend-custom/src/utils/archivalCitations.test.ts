import assert from 'node:assert/strict'
import test from 'node:test'
import {citationInitialPage,citationPageLabel,citationPageTarget,citationSourceGroups,type ArchivalCitation} from './archivalCitations.ts'
import {preprocessCitationTags} from './citationMarkdown.ts'
test('known physical pages and no-page fallback',()=>{
 assert.equal(citationInitialPage({status:'known',pages:[9,10]}),9)
 assert.equal(citationInitialPage({status:'missing',pages:[]}),1)
 assert.equal(citationInitialPage({status:'version_mismatch',pages:[9]}),1)
 assert.equal(citationInitialPage({status:'known',pages:[-1]}),1)
 assert.equal(citationPageLabel({status:'missing',pages:[]}), '未提供页码')
})

test('archival footnotes require a stable evidence ID and preserve legacy tags',()=>{
 const id='a'.repeat(64)
 const tag=`<kb doc="档案" chunk_id="11111111-1111-4111-8111-111111111111" citation_id="${id}" footnote="2" />`
 const html=preprocessCitationTags(tag)
 assert.match(html,/data-citation-id=/);assert.match(html,/\[2\]/)
 assert.doesNotMatch(preprocessCitationTags(tag.replace(id,'bad')),/data-citation-id=/)
 assert.match(preprocessCitationTags('<kb doc="旧档案" chunk_id="old" />'),/旧档案/)
})
test('loaded PDF bounds, missing pages and version mismatch never fake a location',()=>{
 assert.deepEqual(citationPageTarget({status:'known',pages:[2,3]},16),{page:2,warning:''})
 assert.equal(citationPageTarget({status:'known',pages:[2,17]},16).page,1)
 assert.match(citationPageTarget({status:'known',pages:[2,17]},16).warning,/越界/)
 assert.equal(citationPageTarget({status:'known',pages:[]},16).page,1)
 assert.match(citationPageTarget({status:'version_mismatch',pages:[],reason:'原件版本不匹配'},16).warning,/版本不匹配/)
})
test('answer directory groups only its cited evidence by exact original version',()=>{
 const base:ArchivalCitation={id:'a',number:1,chunk_id:'c1',knowledge_id:'k',knowledge_base_id:'kb',title:'parsed.md',archive_file_name:'original.pdf',content_revision:1,content_sha256:'hash',archive_version_id:'v1',source_kind:'chunk',location:{status:'known',pages:[2,3]}}
 const manifest={version:1 as const,state:'verified',repair_attempted:false,unbound_blocks:[],citations:[base,{...base,id:'b',number:2,chunk_id:'c2',location:{status:'known' as const,pages:[3,5]}},{...base,id:'c',number:3,archive_version_id:'v2'},{...base,id:'d',number:4,chunk_id:'c4',location:{status:'missing' as const,pages:[]}},base]}
 const groups=citationSourceGroups(manifest)
 assert.equal(groups.length,2);assert.equal(groups[0].title,'original.pdf')
 assert.deepEqual(groups[0].pages,[2,3,5]);assert.deepEqual(groups[0].withoutPages,[4]);assert.deepEqual(groups[0].chunkIds,['c1','c2','c4'])
 assert.equal(groups[0].citations.length,3);assert.equal(groups[1].version,'v2')
 assert.deepEqual(citationSourceGroups({...manifest,citations:[{...base,source_kind:'derived'}]})[0].pages,[])
 assert.deepEqual(citationSourceGroups(null),[])
})
test('derived Wiki page versions never merge into a false PDF source',()=>{
 const base:ArchivalCitation={id:'a',number:1,chunk_id:'wiki',wiki_page_id:'wiki',knowledge_id:'',knowledge_base_id:'kb',title:'派生页面',content_revision:1,content_sha256:'hash1',source_kind:'derived',location:{status:'missing',pages:[]}}
 const groups=citationSourceGroups({version:1,state:'verified',repair_attempted:false,unbound_blocks:[],citations:[base,{...base,id:'b',number:2,content_revision:2,content_sha256:'hash2'}]})
 assert.equal(groups.length,2);assert.equal(groups[0].derived,true)
 assert.deepEqual(groups[0].pages,[]);assert.deepEqual(groups[0].withoutPages,[1])
})

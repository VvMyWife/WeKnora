export type CitationLocation = {pages:number[];status:'known'|'missing'|'version_mismatch'|'invalid';reason?:string}
export type ArchivalCitation = {id:string;number:number;chunk_id:string;knowledge_id:string;knowledge_base_id:string;title:string;content_revision:number;content_sha256:string;archive_version_id?:string;archive_file_name?:string;archive_sha256?:string;source_kind:string;wiki_page_id?:string;origins?:{knowledge_id:string;title:string;chunk_ids:string[]}[];location:CitationLocation}
export type CitationManifest = {version:1;state:string;citations:ArchivalCitation[];unbound_blocks:{index:number;sha256:string;reason:string}[];repair_attempted:boolean}
export function citationInitialPage(location:CitationLocation):number {
  return location.status==='known'&&Number.isSafeInteger(location.pages[0])&&location.pages[0]>0 ? location.pages[0] : 1
}
export function citationPageLabel(location:CitationLocation):string {
  return location.status==='known' ? `PDF 第 ${location.pages.join('、')} 页` : location.reason||'未提供页码'
}
export function citationPageTarget(location:CitationLocation,total:number):{page:number;warning:string} {
 if(location.status!=='known')return {page:1,warning:`${citationPageLabel(location)}，打开首页；可继续上下浏览`}
 if(!location.pages.length||location.pages.some(p=>!Number.isSafeInteger(p)||p<1||p>total))return {page:1,warning:`来源页码越界或无效，无法准确定位；本 PDF 共 ${total} 页，已打开首页`}
 return {page:location.pages[0],warning:''}
}
export type CitationSourceGroup={key:string;knowledgeId:string;version:string;title:string;derived:boolean;pages:number[];withoutPages:number[];citations:ArchivalCitation[];chunkIds:string[]}
/** Takes only the validated, actually cited manifest; never retrieval results. */
export function citationSourceGroups(manifest?:CitationManifest|null):CitationSourceGroup[]{
 const groups=new Map<string,CitationSourceGroup>(),seen=new Set<string>()
 for(const c of manifest?.citations||[]){
  if(seen.has(c.id))continue;seen.add(c.id)
  const derived=c.source_kind==='derived',key=JSON.stringify([c.wiki_page_id||c.knowledge_id,c.wiki_page_id?[c.content_revision,c.content_sha256]:c.archive_version_id||'',derived])
  let group=groups.get(key)
  if(!group){group={key,knowledgeId:c.knowledge_id,version:c.archive_version_id||'',title:c.archive_file_name||c.title,derived,pages:[],withoutPages:[],citations:[],chunkIds:[]};groups.set(key,group)}
  group.citations.push(c);if(!group.chunkIds.includes(c.chunk_id))group.chunkIds.push(c.chunk_id)
  if(!derived&&c.location.status==='known'&&c.location.pages.length)group.pages.push(...c.location.pages.filter(p=>Number.isSafeInteger(p)&&p>0));else group.withoutPages.push(c.number)
 }
 for(const group of groups.values()){group.pages=[...new Set(group.pages)].sort((a,b)=>a-b);group.citations.sort((a,b)=>a.number-b.number)}
 return [...groups.values()]
}

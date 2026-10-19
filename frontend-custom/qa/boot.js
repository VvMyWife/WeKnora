async (page) => {
 const source=(letter,number)=>({id:letter.repeat(64),number,chunk_id:`33333333-3333-4333-8333-33333333333${number}`,knowledge_id:'44444444-4444-4444-8444-444444444444',knowledge_base_id:'55555555-5555-4555-8555-555555555555',title:`隔离档案 ${number}`,content_revision:1,content_sha256:'c'.repeat(64),archive_version_id:'66666666-6666-4666-8666-666666666666',source_kind:'chunk',location:{status:number===1?'known':'missing',pages:number===1?[2,3]:[],reason:number===1?'':'未提供页码'}})
 const sources=[source('a',1),source('b',2)]
 await page.route('**/api/v1/messages/**',route=>{
  const id=route.request().url().split('/').pop()
  if(id==='citations')return route.fulfill({json:{success:true,data:{version:1,state:'verified',citations:sources,unbound_blocks:[],repair_attempted:false}}})
  const citation=sources.find(c=>c.id===id)
  return citation?route.fulfill({json:{success:true,data:{...citation,snapshot:'<table><tr><th rowspan="2">档案</th><th>数量</th></tr><tr><td>12</td></tr></table>\n\n原始证据正文 '+citation.number,current_revision:2,current_changed:true,current_content:'已修改的新正文',snapshot_at:'2026-10-19T06:00:00Z'}}}):route.fulfill({status:404,json:{error:'not found'}})
 })
 await page.goto('http://10.20.0.15:18185/qa/archival.html')
}

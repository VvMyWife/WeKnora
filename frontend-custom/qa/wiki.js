async (page) => {
 const source={id:'a'.repeat(64),number:1,chunk_id:'wiki-page',wiki_page_id:'wiki-page',knowledge_id:'',knowledge_base_id:'kb',title:'档案管理派生页面',content_revision:3,content_sha256:'c'.repeat(64),source_kind:'derived',location:{status:'missing',pages:[],reason:'派生内容，尚无可验证的原始页来源'},origins:[{knowledge_id:'doc',title:'原始档案',chunk_ids:['origin-chunk']}]}
 await page.route('**/api/v1/messages/**',route=>{
  const url=route.request().url()
  if(url.endsWith('/citations'))return route.fulfill({json:{data:{version:1,state:'verified',citations:[source],unbound_blocks:[],repair_attempted:false}}})
  if(url.includes('/origins/'))return route.fulfill({json:{data:{content:'当前来源分块，与 Wiki 快照严格区分',content_revision:4,state:'current_source_not_snapshot'}}})
  return route.fulfill({json:{data:{...source,snapshot:'引用时的 Wiki 摘录',current_revision:4,current_changed:true,current_content:'当前 Wiki 全文',snapshot_at:'2026-10-19T06:00:00Z'}}})
 })
 await page.reload()
}

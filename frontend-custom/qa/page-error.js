async (page) => {
 await page.route('**/citations/'+ 'a'.repeat(64),route=>route.fulfill({json:{data:{id:'a'.repeat(64),number:1,chunk_id:'33333333-3333-4333-8333-333333333331',knowledge_id:'44444444-4444-4444-8444-444444444444',title:'越界页码样例',content_revision:1,content_sha256:'c'.repeat(64),archive_version_id:'66666666-6666-4666-8666-666666666666',source_kind:'chunk',location:{status:'known',pages:[999]},snapshot:'原始证据正文',snapshot_at:'2026-10-19T06:00:00Z',current_revision:1,current_changed:false}}}))
}

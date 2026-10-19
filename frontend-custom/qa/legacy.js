async (page) => {
 await page.route('**/citations',route=>route.fulfill({json:{data:{version:1,state:'legacy',citations:[],unbound_blocks:[],repair_attempted:false}}}))
 await page.reload()
}

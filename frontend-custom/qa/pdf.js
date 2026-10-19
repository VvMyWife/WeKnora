async (page) => {
 await page.route('**/api/v1/knowledge/**',route=>route.fulfill({path:'E:/desktop/文档RAG/deploy/weknora_citations/runtime/pdf-fixtures/sixteen.pdf',contentType:'application/pdf'}))
}

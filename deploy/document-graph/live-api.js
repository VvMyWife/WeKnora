async(page) => {
  // Read-only acceptance against the currently opened, authorized KB.
  const result = await page.evaluate(async () => {
    const kb = location.pathname.split('/').pop()
    const headers = { Authorization: 'Bearer ' + localStorage.getItem('weknora_token') }
    const read = async path => {
      const response = await fetch('/api/v1/' + path, {headers})
      return {status: response.status, body: await response.json()}
    }
    const ensure = (value, message) => { if (!value) throw new Error(message) }
    const listed = await read(`knowledge-bases/${kb}/knowledge?page=1&page_size=50&include_archive_source=true`)
    ensure(listed.status === 200 && Array.isArray(listed.body.data), 'document list unavailable')
    const docs = listed.body.data
    ensure(docs.length >= 2, 'acceptance requires two existing documents')
    const ids = docs.slice(0, 3).map(doc => doc.id)
    const scope = ids.map(id => 'knowledge_ids=' + id).join('&')
    const base = `knowledgebase/${kb}/wiki/graph?scope=documents&${scope}`
    const union = await read(base + '&view=union&limit=500')
    ensure(union.status === 200 && union.body.meta.scope === 'documents', 'scoped endpoint not active: ' + JSON.stringify(union))
    ensure(union.body.nodes.length > 0, 'fixture documents have no Wiki source nodes')
    for (const node of union.body.nodes) {
      ensure(node.source_knowledge_ids.length > 0 && node.source_knowledge_ids.every(id => ids.includes(id)), 'outside source leaked')
      ensure(node.shared === (node.source_knowledge_ids.length >= 2) || (!node.shared && node.source_knowledge_ids.length === 1), 'incorrect shared node')
    }
    const connections = await read(base + '&view=connections&limit=500')
    ensure(connections.status === 200, 'connections query failed')
    const links = connections.body.edges || []
    for (const link of links) ensure(link.cross_source && (union.body.meta.truncated || union.body.edges.some(edge => edge.source === link.source && edge.target === link.target)), 'invented or non-cross-source link')
    for (const node of connections.body.nodes) ensure(node.shared || links.some(link => link.source === node.slug || link.target === node.slug), 'unconnected private node in connections view')
    const one = await read(`knowledgebase/${kb}/wiki/graph?scope=documents&knowledge_ids=${ids[0]}&view=connections`)
    ensure(one.status === 200 && one.body.nodes.length === 0, 'single source invented connections')
    const limited = await read(base + '&view=union&limit=1')
    ensure(limited.status === 200 && limited.body.nodes.length <= 1 && limited.body.meta.total === union.body.meta.total, 'limit before source filtering or wrong count')
    const invalid = await read(`knowledgebase/${kb}/wiki/graph?scope=documents&knowledge_ids=00000000-0000-0000-0000-000000000001`)
    const empty = await read(`knowledgebase/${kb}/wiki/graph?scope=documents`)
    ensure(invalid.status === 400 && empty.status === 400, 'invalid selection accepted')
    const anonymous = await fetch('/api/v1/' + base, {credentials: 'omit'})
    ensure(anonymous.status === 401, 'anonymous graph access accepted')
    const node = union.body.nodes[0]
    const ego = await read(base + '&mode=ego&center=' + encodeURIComponent(node.slug) + '&depth=2')
    ensure(ego.status === 200 && ego.body.nodes.every(item => item.source_knowledge_ids.every(id => ids.includes(id))), 'ego escaped scope')
    const search = await read(`knowledgebase/${kb}/wiki/search?scope=documents&${scope}&view=union&q=${encodeURIComponent(node.title.slice(0, 8))}&limit=50`)
    ensure(search.status === 200 && search.body.pages.length > 0 && search.body.pages.every(item => item.source_knowledge_ids.every(id => ids.includes(id))), 'search escaped scope')
    for (const candidate of union.body.nodes.slice(0, 5)) {
      const full = await read(`knowledgebase/${kb}/wiki/pages/${candidate.slug.split('/').map(encodeURIComponent).join('/')}`)
      ensure(full.status === 200, 'source Wiki page unavailable')
      {
        const wiki = full.body.data || full.body
        const sourceIDs = (wiki.source_refs || []).map(ref => ref.split('|')[0].trim())
        ensure(candidate.source_knowledge_ids.every(id => sourceIDs.includes(id)), 'graph source provenance mismatch')
      }
    }
    return {kb, selected: ids, documents: docs.length, mapped: docs.filter(doc => doc.archive_source).length, union: union.body.meta, connections: connections.body.meta, edges: links.length, ego:ego.body.nodes.length, search:search.body.pages.length, invalid:invalid.status, empty:empty.status, anonymous:anonymous.status, model_calls:0, requests:'GET only'}
  })
  return result
}

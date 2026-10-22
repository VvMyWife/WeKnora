/** Presentation only. Never mutate the saved answer or citation manifest. */
const children = (node: Node): Node[] => Array.from(node.childNodes)
const element = (node: Node): node is Element => node.nodeType === 1
const tag = (node: Node) => element(node) ? node.tagName.toLowerCase() : ''
const classes = (node: Element) => (node.getAttribute('class') || '').split(/\s+/)
const isCitation = (node: Node): node is Element => element(node)
  && classes(node).includes('citation-archival')
  && /^[a-f0-9]{64}$/.test(node.getAttribute('data-citation-id') || '')
const meaningful = (text: string) => /[\p{L}\p{N}]/u.test(text)

type Cluster = { nodes: Element[]; key: string }
type Scan = { clusters: Cluster[]; blocked: boolean; tail: string }
function scan(root: Element): Scan {
  const result: Scan = { clusters: [], blocked: false, tail: '' }
  let between = ''
  const walk = (node: Node) => {
    if (isCitation(node)) {
      const last = result.clusters.at(-1)
      if (last && !meaningful(between)) last.nodes.push(node)
      else result.clusters.push({ nodes: [node], key: '' })
      between = ''
      return
    }
    if (element(node)) {
      if (['code','pre','table','ul','ol','h1','h2','h3','h4','h5','h6'].includes(tag(node))
        || classes(node).some(c => c === 'citation-unbound' || c === 'citation-kb' || c === 'citation-web')) {
        result.blocked = true
        return
      }
    }
    if (node.nodeType === 3) between += node.nodeValue || ''
    else children(node).forEach(walk)
  }
  children(root).forEach(walk)
  result.tail = between
  for (const cluster of result.clusters) {
    cluster.key = [...new Set(cluster.nodes.map(n => n.getAttribute('data-citation-id')!))].sort().join(',')
  }
  return result
}
function remove(nodes: Element[]) { nodes.forEach(n => n.parentNode?.removeChild(n)) }
function scopeHint(nodes: Element[], size: number, list = false) {
  for (const node of nodes) {
    const hint = list ? `对应以上 ${size} 个连续同源条目` : `对应本段以上 ${size} 处连续同源内容`
    node.setAttribute('title', hint)
    node.setAttribute('data-citation-group-size', String(size))
    node.setAttribute('aria-label', `查看脚注 ${node.textContent || ''}；${hint}`)
  }
}

/** DOM operations only; accepts an injected DOM in Node regression tests. */
export function groupCitationTree(root: Element, streaming = false): void {
  const lastBranch = children(root).filter(element).at(-1)
  const openTail = (node: Node): boolean => {
    if (!streaming || !lastBranch) return false
    for (let p: Node | null = node; p; p = p.parentNode) if (p === lastBranch) return true
    return false
  }
  const all: Element[] = []
  const collect = (node: Node) => { if (element(node)) all.push(node); children(node).forEach(collect) }
  children(root).forEach(collect)
  for (const unit of all.filter(n => tag(n) === 'p' || tag(n) === 'li')) {
    if (openTail(unit)) continue
    const s = scan(unit)
    if (s.blocked || meaningful(s.tail)) continue
    for (let i = 0; i < s.clusters.length;) {
      let end = i + 1
      while (end < s.clusters.length && s.clusters[end]!.key === s.clusters[i]!.key) end++
      if (end - i > 1) {
        for (let j = i; j < end - 1; j++) remove(s.clusters[j]!.nodes)
        scopeHint(s.clusters[end - 1]!.nodes, end - i)
      }
      i = end
    }
  }
  for (const list of all.filter(n => tag(n) === 'ul' || tag(n) === 'ol')) {
    const items = children(list).filter(element)
    for (let i = 0; i < items.length;) {
      const first = items[i]!, s = scan(first)
      const key = !openTail(first) && tag(first) === 'li' && !s.blocked && !meaningful(s.tail)
        && s.clusters.length === 1 ? s.clusters[0]!.key : ''
      if (!key) { i++; continue }
      let end = i + 1
      for (; end < items.length; end++) {
        const next = items[end]!, n = scan(next)
        if (openTail(next) || tag(next) !== 'li' || n.blocked || meaningful(n.tail)
          || n.clusters.length !== 1 || n.clusters[0]!.key !== key) break
      }
      if (end - i > 1) {
        for (let j = i; j < end - 1; j++) remove(scan(items[j]!).clusters.flatMap(c => c.nodes))
        scopeHint(scan(items[end - 1]!).clusters.flatMap(c => c.nodes), end - i, true)
      }
      i = end
    }
  }
}

export function groupCitationHtml(html: string, streaming = false): string {
  if (typeof DOMParser === 'undefined' || !html.includes('citation-archival')) return html
  const doc = new DOMParser().parseFromString(html, 'text/html')
  groupCitationTree(doc.body, streaming)
  return doc.body.innerHTML
}

/** Remove presentation notices only, after grouping has used them as boundaries. */
export function hideUnboundCitationTree(root: Element): void {
  const walk = (node: Node) => {
    // Literal examples and document content in code blocks are not UI notices.
    if (['pre', 'code'].includes(tag(node))) return
    if (element(node) && tag(node) === 'span' && classes(node).includes('citation-unbound')) {
      node.parentNode?.removeChild(node)
      return
    }
    children(node).forEach(walk)
  }
  children(root).forEach(walk)
}

export function hideUnboundCitationHtml(html: string): string {
  if (typeof DOMParser === 'undefined' || !html.includes('citation-unbound')) return html
  const doc = new DOMParser().parseFromString(html, 'text/html')
  hideUnboundCitationTree(doc.body)
  return doc.body.innerHTML
}

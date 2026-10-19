package service

import (
	"errors"
	"sort"

	"github.com/Tencent/WeKnora/internal/types"
)

// scopeWikiGraph clones graph projections, never mutating stored pages. Filtering
// precedes ranking, BFS and limits, so off-scope links cannot influence traversal.
func scopeWikiGraph(pages []*types.WikiPage, req *types.WikiGraphRequest) ([]*types.WikiPage, map[string][]string, error) {
	if len(req.KnowledgeIDs) == 0 || len(req.KnowledgeIDs) > 50 {
		return nil, nil, errors.New("select between 1 and 50 documents")
	}
	if req.View != "" && req.View != "union" && req.View != "connections" {
		return nil, nil, errors.New("invalid document graph view")
	}
	selected := map[string]struct{}{}
	for _, id := range req.KnowledgeIDs {
		selected[id] = struct{}{}
	}
	sources := map[string][]string{}
	bySlug := map[string]*types.WikiPage{}
	for _, p := range pages {
		ids := map[string]bool{}
		for _, id := range p.SourceKnowledgeIDs() {
			if _, ok := selected[id]; ok {
				ids[id] = true
			}
		}
		if len(ids) == 0 {
			continue
		}
		for id := range ids {
			sources[p.Slug] = append(sources[p.Slug], id)
		}
		sort.Strings(sources[p.Slug])
		copyPage := *p
		bySlug[p.Slug] = &copyPage
	}
	keep := map[string]bool{}
	for slug, p := range bySlug {
		if req.View != "connections" || len(sources[slug]) > 1 {
			keep[slug] = true
		}
		for _, target := range p.OutLinks {
			if bySlug[target] != nil && crossWikiSources(sources[slug], sources[target]) {
				keep[slug], keep[target] = true, true
			}
		}
	}
	result := make([]*types.WikiPage, 0, len(keep))
	for slug, p := range bySlug {
		if !keep[slug] {
			continue
		}
		out := types.StringArray{}
		for _, target := range p.OutLinks {
			if keep[target] && (req.View != "connections" || crossWikiSources(sources[slug], sources[target])) {
				out = append(out, target)
			}
		}
		p.OutLinks, p.InLinks = out, types.StringArray{}
		result = append(result, p)
	}
	// Derive inbound links from the retained outbound links, not stale/global degrees.
	for _, p := range result {
		for _, target := range p.OutLinks {
			bySlug[target].InLinks = append(bySlug[target].InLinks, p.Slug)
		}
	}
	return result, sources, nil
}

func crossWikiSources(left, right []string) bool {
	for _, a := range left {
		for _, b := range right {
			if a != b {
				return true
			}
		}
	}
	return false
}

func decorateDocumentGraph(graph *types.WikiGraphData, pages []*types.WikiPage, sources map[string][]string, req *types.WikiGraphRequest) {
	graph.Meta.Scope, graph.Meta.View = "documents", req.View
	if graph.Meta.View == "" {
		graph.Meta.View = "union"
	}
	visible := map[string]bool{}
	for i := range graph.Nodes {
		n := &graph.Nodes[i]
		n.SourceKnowledgeIDs = sources[n.Slug]
		n.Shared = len(n.SourceKnowledgeIDs) > 1
		visible[n.Slug] = true
	}
	for i := range graph.Edges {
		e := &graph.Edges[i]
		e.CrossSource = crossWikiSources(sources[e.Source], sources[e.Target])
	}
	seen := map[string]bool{}
	for _, id := range req.KnowledgeIDs {
		if seen[id] {
			continue
		}
		seen[id] = true
		coverage := types.WikiGraphDocumentCoverage{KnowledgeID: id}
		for _, p := range pages {
			for _, source := range sources[p.Slug] {
				if source == id {
					coverage.Total++
					if visible[p.Slug] {
						coverage.Returned++
					}
					break
				}
			}
		}
		graph.Meta.Documents = append(graph.Meta.Documents, coverage)
	}
}

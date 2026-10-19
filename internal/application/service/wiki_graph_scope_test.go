package service

import (
	"github.com/Tencent/WeKnora/internal/types"
	"testing"
)

func TestDocumentGraphScopeBeforeLimitAndExactSources(t *testing.T) {
	pages := []*types.WikiPage{
		{Slug: "outside", SourceRefs: types.StringArray{"ab|not a"}, OutLinks: types.StringArray{"outside", "outside", "outside"}},
		{Slug: "a", SourceRefs: types.StringArray{"a|First", "a"}, OutLinks: types.StringArray{"b", "outside"}},
		{Slug: "b", SourceRefs: types.StringArray{"b|Second"}},
		{Slug: "shared", SourceRefs: types.StringArray{"a", "b|Second"}},
	}
	req := &types.WikiGraphRequest{Scope: "documents", KnowledgeIDs: []string{"a", "b"}, Limit: 1}
	g, err := computeGraphSubset(pages, req)
	if err != nil || g.Meta.Total != 3 || len(g.Nodes) != 1 || !g.Meta.Truncated {
		t.Fatalf("%+v %v", g, err)
	}
	if g.Nodes[0].Slug == "outside" {
		t.Fatal("scope applied after limit")
	}
	req.Limit = 500
	g, err = computeGraphSubset(pages, req)
	if err != nil || len(g.Edges) != 1 || !g.Edges[0].CrossSource {
		t.Fatalf("%+v %v", g, err)
	}
	for _, n := range g.Nodes {
		if n.Slug == "shared" && (!n.Shared || len(n.SourceKnowledgeIDs) != 2) {
			t.Fatal(n)
		}
	}
	if len(pages[1].OutLinks) != 2 {
		t.Fatal("mutated repository page")
	}
}

func TestDocumentGraphConnectionsNeverInventLinks(t *testing.T) {
	pages := []*types.WikiPage{{Slug: "a", SourceRefs: types.StringArray{"a"}}, {Slug: "b", SourceRefs: types.StringArray{"b"}}}
	req := &types.WikiGraphRequest{Scope: "documents", KnowledgeIDs: []string{"a", "b"}, View: "connections", Limit: 200}
	g, err := computeGraphSubset(pages, req)
	if err != nil || len(g.Nodes) != 0 || len(g.Edges) != 0 {
		t.Fatalf("%+v %v", g, err)
	}
	req.KnowledgeIDs = nil
	if _, err = computeGraphSubset(pages, req); err == nil {
		t.Fatal("empty scope accepted")
	}
}

func TestDocumentGraphEgoCannotTraverseOutside(t *testing.T) {
	pages := []*types.WikiPage{
		{Slug: "a", SourceRefs: types.StringArray{"a"}, OutLinks: types.StringArray{"outside"}},
		{Slug: "outside", OutLinks: types.StringArray{"b"}},
		{Slug: "b", SourceRefs: types.StringArray{"a"}},
	}
	g, err := computeGraphSubset(pages, &types.WikiGraphRequest{Scope: "documents", KnowledgeIDs: []string{"a"}, Mode: "ego", Center: "a", Depth: 3, Limit: 500})
	if err != nil || len(g.Nodes) != 1 {
		t.Fatalf("%+v %v", g, err)
	}
}

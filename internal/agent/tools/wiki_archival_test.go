package tools

import (
	"github.com/Tencent/WeKnora/internal/types"
	"strings"
	"testing"
)

func TestWikiArchivalEvidenceUsesOnlyBudgetedBodies(t *testing.T) {
	t.Setenv("ARCHIVAL_CITATIONS_ENABLED", "true")
	pages := []pendingWikiPage{}
	for _, id := range []string{"one", "two", "three"} {
		pages = append(pages, pendingWikiPage{page: &types.WikiPage{ID: id, Slug: id, Title: id, Version: 2}, kbID: "kb", body: strings.Repeat(id+" ", 2000)})
	}
	out, _, omitted, rows := renderWikiPagesWithEvidence(pages, 1600)
	if len(rows) == 0 || len(rows) > len(pages)-len(omitted) {
		t.Fatal("missing or omitted evidence")
	}
	for _, row := range rows {
		body, _ := row["content"].(string)
		if body == "" || !strings.Contains(out, body) || row["wiki_version"] != 2 {
			t.Fatal("non-supplied content registered")
		}
	}
	t.Setenv("ARCHIVAL_CITATIONS_ENABLED", "false")
	_, _, _, rows = renderWikiPagesWithEvidence(pages, 1600)
	if len(rows) != 0 {
		t.Fatal("feature-off changed legacy tool data")
	}
}

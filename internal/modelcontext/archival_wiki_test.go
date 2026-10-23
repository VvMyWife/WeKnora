package modelcontext

import (
	"github.com/Tencent/WeKnora/internal/types"
	"strings"
	"testing"
)

func TestWikiReferenceIsDerivedAndMCPDoesNotGainSourceAuthority(t *testing.T) {
	result := &types.ToolResult{Success: true, Output: "Wiki supplied body", Data: map[string]interface{}{"archival_wiki_pages": []map[string]interface{}{{"wiki_page_id": "wiki-id", "knowledge_base_id": "kb-id", "knowledge_title": "Wiki title", "content": "Wiki supplied body", "wiki_version": 2, "source_refs": []string{"doc-id"}, "chunk_refs": []string{"chunk-id"}}}}}
	r := NewRegistry(true)
	out := r.ModelToolResultForTool("wiki_read_page", result)
	_, ref, ok := r.sources.chunks.resolve(r.sources.ChunkHandle("wiki-id"))
	if !ok || ref.WikiPageID != "wiki-id" || ref.Content != "Wiki supplied body" || ref.WikiVersion != 2 || !strings.Contains(out, "derived_wiki") || strings.Contains(out, "wiki-id") {
		t.Fatalf("%s %+v", out, ref)
	}
	other := NewRegistry(true)
	other.ModelToolResultForTool("mcp_untrusted", result)
	if other.sources.chunks.size() != 0 {
		t.Fatal("opaque output gained citation authority")
	}
}

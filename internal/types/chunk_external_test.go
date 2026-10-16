package types

import "testing"

func TestExternalChunkEmbeddingKeepsEvidenceAndContextSeparate(t *testing.T) {
	c := &Chunk{Content: "原文", Metadata: JSON(`{"external_import":{"retrieval_text":"检索上下文\n原文"}}`)}
	if c.EmbeddingContent() != "检索上下文\n原文" || c.Content != "原文" {
		t.Fatal("lost external content separation")
	}
	c.ContentRevision = 1
	c.Content = "修改后"
	if c.EmbeddingContent() != "修改后" {
		t.Fatal("manual edits must not use stale retrieval text")
	}
}

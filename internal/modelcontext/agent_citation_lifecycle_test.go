package modelcontext

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/Tencent/WeKnora/internal/archivalcitations"
	"github.com/Tencent/WeKnora/internal/models/chat"
	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAgentCitationToolsToPersistedEvidence(t *testing.T) {
	t.Setenv("ARCHIVAL_CITATIONS_ENABLED", "true")
	for _, kind := range []string{"search_results", "grep_results", "graph_query_results", "knowledge_chunks_list"} {
		t.Run(kind, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if err = db.AutoMigrate(&archivalcitations.Record{}, &archivalcitations.ManifestRecord{}); err != nil {
				t.Fatal(err)
			}
			for _, sql := range []string{
				"CREATE TABLE chunks(id TEXT,knowledge_id TEXT,knowledge_base_id TEXT,content TEXT,chunk_type TEXT,metadata TEXT,relation_chunks TEXT,tenant_id INTEGER,content_revision INTEGER,deleted_at DATETIME)",
				"CREATE TABLE knowledges(id TEXT,tenant_id INTEGER,knowledge_base_id TEXT,title TEXT,deleted_at DATETIME)",
				"CREATE TABLE archive_mapping_links(tenant_id INTEGER,knowledge_id TEXT,version_id TEXT)",
				"CREATE TABLE archive_mapping_versions(id TEXT,tenant_id INTEGER,knowledge_id TEXT,sha256 TEXT,file_name TEXT)",
				"INSERT INTO chunks VALUES('chunk-real','doc-real','kb-real','original table row','text','{}','[]',1,1,NULL)",
				"INSERT INTO knowledges VALUES('doc-real',1,'kb-real','document',NULL)",
			} {
				if err = db.Exec(sql).Error; err != nil {
					t.Fatal(err)
				}
			}
			store := archivalcitations.NewStore(db)
			ctx := archivalcitations.WithScope(context.Background(), store, 1, "session", "message")
			r := NewRegistry(true)
			row := map[string]interface{}{"chunk_id": "chunk-real", "id": "chunk-real", "knowledge_id": "doc-real", "knowledge_base_id": "kb-real", "content": "original table row", "knowledge_title": "document", "chunk_type": "text"}
			data := map[string]interface{}{"display_type": kind, "results": []map[string]interface{}{row}, "chunk_results": []map[string]interface{}{row}, "chunks": []map[string]interface{}{row}, "knowledge_id": "doc-real", "knowledge_base_id": "kb-real", "total_chunks": 1, "fetched_chunks": 1}
			tool := "search_knowledge"
			if kind == "grep_results" {
				tool = "grep_chunks"
			}
			if kind == "graph_query_results" {
				tool = "query_knowledge_graph"
			}
			if kind == "knowledge_chunks_list" {
				tool = "read_document"
			}
			result := &types.ToolResult{Success: true, Output: "canonical tool output", Data: data}
			output := r.ModelToolResultForTool(tool, result)
			if !strings.Contains(output, `id="c1"`) {
				t.Fatalf("no adjacent handle: %s", output)
			}
			messages := r.WithAgentCitationReminder(r.EncodeMessages([]chat.Message{{Role: "tool", Name: tool, Content: output}}))
			if !strings.Contains(messages[0].Content, "<cite>") || result.Output != "canonical tool output" {
				t.Fatal("projection changed canonical tool output")
			}
			r.CaptureArchival(ctx)
			decoder := r.StreamDecoder()
			answer := decoder.Feed(`Fact<ref id="c1"/>`) + decoder.Flush()
			out, manifest, err := archivalcitations.FromContext(ctx).Finalize(ctx, answer)
			if err != nil || len(manifest.Citations) != 1 || !strings.Contains(out, `footnote="1"`) {
				t.Fatalf("citation not persisted: %s %+v %v", out, manifest, err)
			}
			evidence, _, err := store.Evidence(ctx, 1, "session", "message", manifest.Citations[0].ID)
			if err != nil || evidence.Snapshot != "original table row" {
				t.Fatal("wrong snapshot", err)
			}
		})
	}
}

func TestAgentHandlesSurviveParallelToolsAndCompactedHistory(t *testing.T) {
	r := NewRegistry(true)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.RegisterChunk(ChunkReference{ChunkID: "same-chunk", Content: "fact", KnowledgeID: "doc", KnowledgeBaseID: "kb"})
		}()
	}
	wg.Wait()
	h := r.sources.ChunkHandle("same-chunk")
	if h == "" {
		t.Fatal("parallel registration lost handle")
	}
	// The same registry spans the compaction boundary; the summary is not a source.
	r.EncodeMessages([]chat.Message{{Role: "assistant", Content: "Summary: fact from " + h}})
	if r.sources.ChunkHandle("same-chunk") != h {
		t.Fatal("compaction renumbered source")
	}
	d := r.StreamDecoder()
	out := d.Feed(`fact<ref id="`+h+`"/>`) + d.Flush()
	if !strings.Contains(out, `chunk_id="same-chunk"`) {
		t.Fatal(out)
	}
	newRequest := NewRegistry(true)
	newRequest.EncodeMessages([]chat.Message{{Role: "assistant", Content: out}})
	newDecoder := newRequest.StreamDecoder()
	if got := newDecoder.Feed(`<ref id="c1"/>`) + newDecoder.Flush(); got != "" {
		t.Fatal("history authorized new evidence", got)
	}
}

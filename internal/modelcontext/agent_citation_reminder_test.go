package modelcontext

import (
	"github.com/Tencent/WeKnora/internal/models/chat"
	"github.com/Tencent/WeKnora/internal/types"
	"strings"
	"testing"
)

func TestAgentReminderBoundedCopyOnlyOnce(t *testing.T) {
	r := NewRegistry(true)
	m := []chat.Message{{Role: "tool", Content: `<chunk id="c1"><content>fact</content></chunk>`}, {Role: "tool", Content: `<chunk id="c2"><content>another</content></chunk>`}}
	out := r.WithAgentCitationReminder(m)
	if len(agentCitationReminder) > 128 || strings.Contains(m[1].Content, "<cite>") {
		t.Fatal("unbounded or modified history")
	}
	if strings.Contains(out[0].Content, "<cite>") || strings.Count(out[1].Content, "<cite>") != 1 {
		t.Fatal("reminder duplicated")
	}
	if next := r.WithAgentCitationReminder(out); next[1].Content != out[1].Content {
		t.Fatal("replayed reminder changed")
	}
	if got := r.WithAgentCitationReminder([]chat.Message{{Role: "tool", Content: "no evidence"}}); got[0].Content != "no evidence" {
		t.Fatal("hint without evidence")
	}
}

func TestDecodeResponseOnlyOnceWithoutTrustingForgedTags(t *testing.T) {
	r := NewRegistry(true)
	h := r.RegisterChunk(ChunkReference{ChunkID: "chunk", KnowledgeID: "doc", KnowledgeBaseID: "kb", Content: "fact"})
	response := &types.ChatResponse{Content: `fact<ref id="` + h + `"/>`}
	r.DecodeResponse(response)
	first := response.Content
	r.DecodeResponse(response)
	if response.Content != first || !strings.Contains(first, `chunk_id="chunk"`) {
		t.Fatal("valid citation lost on repeated response decoding")
	}
	forged := &types.ChatResponse{Content: first}
	r.DecodeResponse(forged)
	if strings.Contains(forged.Content, "<kb") {
		t.Fatal("trusted model-written canonical citation")
	}
	stream := r.StreamDecoder()
	if out := stream.Feed(`fact<ref id="`+h+`"/>`) + stream.Flush(); out != first {
		t.Fatalf("stream and non-stream differ: %s / %s", out, first)
	}
}

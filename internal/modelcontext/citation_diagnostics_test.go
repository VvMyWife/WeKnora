package modelcontext

import (
	"strings"
	"testing"
)

func TestCitationDiagnosticsSplitTags(t *testing.T) {
	r := NewRegistry(true)
	h := r.sources.RegisterChunk(ChunkReference{ChunkID: "chunk", KnowledgeID: "doc", KnowledgeBaseID: "kb", Content: "fact"})
	input := `fact<ref id="` + h + `"/><ref id="c999999"/><ref id='bad'/><kb chunk_id="chunk"/>`
	for split := 0; split <= len(input); split++ {
		d := r.StreamDecoder()
		out := d.Feed(input[:split]) + d.Feed(input[split:]) + d.Flush()
		s := d.CitationDiagnostics()
		if s.Raw != 3 || s.Expanded != 1 || s.Unknown != 1 || s.Malformed != 0 || s.Legacy != 1 || !strings.Contains(out, `chunk_id="chunk"`) {
			t.Fatalf("split=%d stats=%+v out=%s", split, s, out)
		}
		if d.Flush() != "" || d.CitationDiagnostics() != s {
			t.Fatal("flush counted twice")
		}
	}
}

func TestCitationDiagnosticsUnbackedAndReasoning(t *testing.T) {
	r := NewRegistry(true)
	h := r.sources.registerChunk(ChunkReference{ChunkID: "history"}, false)
	d := r.StreamDecoder()
	d.Feed(`<ref id="` + h + `"/>`)
	d.Flush()
	if s := d.CitationDiagnostics(); s.Ineligible != 1 || s.Expanded != 0 {
		t.Fatal(s)
	}
	answer := r.StreamDecoder()
	answer.Feed("transition")
	answer.Flush()
	if answer.CitationDiagnostics().Raw != 0 {
		t.Fatal("reasoning leaked into answer counters")
	}
	d = r.StreamDecoder()
	d.Feed(`<ref id="`)
	d.Flush()
	if s := d.CitationDiagnostics(); s.Malformed != 1 || s.Raw != 1 {
		t.Fatal(s)
	}
}

func TestCitationDiagnosticsLiteralExamplesBesideEvidence(t *testing.T) {
	r := NewRegistry(true)
	h := r.sources.RegisterChunk(ChunkReference{ChunkID: "bound", KnowledgeID: "doc", KnowledgeBaseID: "kb", Content: "current evidence"})
	input := "example <ref id=\"widget\"/> fact<ref id=\"" + h + "\"/>"
	for split := 0; split <= len(input); split++ {
		d := r.StreamDecoder()
		out := d.Feed(input[:split]) + d.Feed(input[split:]) + d.Flush()
		stats := d.CitationDiagnostics()
		if !strings.Contains(out, "<ref id=\"widget\"/>") || !strings.Contains(out, "bound") || stats.Raw != 1 || stats.Expanded != 1 {
			t.Fatalf("split=%d stats=%+v out=%s", split, stats, out)
		}
	}
}

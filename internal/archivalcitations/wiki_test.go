package archivalcitations

import (
	"context"
	"strings"
	"testing"
)

func TestWikiPinsDerivedBodyAndOnlyExistingSameKBOrigins(t *testing.T) {
	s := requestFixture(t)
	db := s.store.DB
	for _, q := range []string{
		"CREATE TABLE wiki_pages(id TEXT,knowledge_base_id TEXT,title TEXT,tenant_id INTEGER,deleted_at DATETIME)",
		"CREATE TABLE knowledges(id TEXT,title TEXT,tenant_id INTEGER,knowledge_base_id TEXT,deleted_at DATETIME)",
		"CREATE TABLE chunks(id TEXT,knowledge_id TEXT,tenant_id INTEGER,knowledge_base_id TEXT,deleted_at DATETIME)",
		"INSERT INTO wiki_pages VALUES('wiki','kb','Wiki title',2,NULL)",
		"INSERT INTO knowledges VALUES('doc','原文',2,'kb',NULL),('foreign','不可泄露',3,'other',NULL)",
		"INSERT INTO chunks VALUES('origin','doc',2,'kb',NULL),('bad','foreign',3,'other',NULL)",
	} {
		if e := db.Exec(q).Error; e != nil {
			t.Fatal(e)
		}
	}
	if e := s.OfferWiki(context.Background(), "wiki", "kb", "实际提供的派生页面摘录", 3, []string{"doc|原文", "foreign|不可泄露"}, []string{"origin", "bad"}); e != nil {
		t.Fatal(e)
	}
	if e := s.OfferWiki(context.Background(), "wiki", "wrong", "wrong body", 1, nil, nil); e == nil {
		t.Fatal("wrong KB accepted")
	}
	// The wrong-KB offer above must be a fresh identity to exercise lookup.
	out, m, e := s.Finalize(context.Background(), `派生结论<kb chunk_id="wiki" />`)
	if e != nil {
		t.Fatal(e)
	}
	c := m.Citations[0]
	if c.WikiPageID != "wiki" || c.SourceKind != "derived" || c.ArchiveVersionID != "" || c.Location.Status != PageMissing || len(c.Origins) != 1 || c.Origins[0].KnowledgeID != "doc" || len(c.Origins[0].ChunkIDs) != 1 {
		t.Fatalf("%s %+v", out, c)
	}
	row, _, e := s.store.Evidence(context.Background(), 1, "s", "m", c.ID)
	if e != nil || row.Snapshot != "实际提供的派生页面摘录" {
		t.Fatal("lost wiki snapshot", e)
	}
}
func TestNoKnowledgeAnswerAndLongUncitedAnswerStayLocal(t *testing.T) {
	s := requestFixture(t)
	s.evidence = map[string]offeredEvidence{}
	out, m, e := s.Finalize(context.Background(), "你好")
	if e != nil || out != "你好" || m.State != "not_applicable" {
		t.Fatal(out, m, e)
	}
	s = requestFixture(t)
	answer := strings.Repeat("long answer ", 3000)
	out, m, e = s.Finalize(context.Background(), answer)
	if e != nil || m.RepairAttempted || !strings.Contains(out, answer) || m.State != "needs_review" {
		t.Fatal("unexpected synthesized binding", e)
	}
}
func TestLiteralCodeIsNotRewrittenAsAnAnswerCitation(t *testing.T) {
	s := requestFixture(t)
	answer := "```xml\n<kb chunk_id=\"literal\" />\n```\n\n正文<kb chunk_id=\"c1\" />"
	out, m, e := s.Finalize(context.Background(), answer)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(out, "```xml\n<kb chunk_id=\"literal\" />\n```") || len(m.Citations) != 1 || len(m.Unbound) != 0 {
		t.Fatal(out, m)
	}
}

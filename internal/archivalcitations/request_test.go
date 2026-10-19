package archivalcitations

import (
	"context"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"testing"
)

func requestFixture(t *testing.T) *Scope {
	t.Helper()
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&Record{}, &ManifestRecord{}); e != nil {
		t.Fatal(e)
	}
	s := &Scope{store: NewStore(db), tenant: 1, session: "s", message: "m", evidence: map[string]offeredEvidence{}}
	for _, id := range []string{"c1", "c2"} {
		snapshot := "evidence " + id
		sum := Summary{ID: EvidenceID(1, "s", "m", id, snapshot, "v1", 1), ChunkID: id, KnowledgeID: "k", KnowledgeBaseID: "kb", Title: "title", ContentRevision: 1, ContentSHA256: Hash(snapshot), ArchiveVersionID: "v1", SourceKind: "chunk", Location: Location{Pages: []int{2, 3}, Status: PageKnown}}
		s.evidence[id] = offeredEvidence{sum, Record{ID: sum.ID, TenantID: 1, KnowledgeTenantID: 1, SessionID: "s", MessageID: "m", KnowledgeID: "k", ChunkID: id, Snapshot: snapshot}}
	}
	return s
}
func TestOnlyActuallyCitedEvidenceIsStored(t *testing.T) {
	s := requestFixture(t)
	answer := `事实<kb chunk_id="c1" />` + "\n\n" + `同一事实<kb chunk_id="c1" />` + "\n\n" + `未知事实<kb chunk_id="invented" />`
	out, m, e := s.Finalize(context.Background(), answer)
	if e != nil {
		t.Fatal(e)
	}
	if len(m.Citations) != 1 || strings.Count(out, `footnote="1"`) != 2 || strings.Contains(out, "invented") || len(m.Unbound) != 1 || m.State != "needs_review" {
		t.Fatalf("%s %+v", out, m)
	}
	if _, _, e = s.store.Evidence(context.Background(), 1, "s", "m", s.evidence["c2"].summary.ID); e == nil {
		t.Fatal("uncited source persisted")
	}
	if _, _, e = s.Finalize(context.Background(), out); e != nil {
		t.Fatal(e)
	}
}
func TestRepairAddsBindingsWithoutRewritingFactsAndRunsOnlyOnce(t *testing.T) {
	s := requestFixture(t)
	calls := 0
	s.SetRepair(func(context.Context, string) (string, error) {
		calls++
		return `{"rewritten_answer":"fake","additions":[{"line_index":2,"chunk_id":"c1"},{"line_index":3,"chunk_id":"unoffered"},{"line_index":99,"chunk_id":"c2"}]}`, nil
	})
	answer := "| 名称 | 数量 |\n| --- | --- |\n| 表格事实 | 12 |\n未支持的内容"
	out, m, e := s.Finalize(context.Background(), answer)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(out, "表格事实 | 12") || strings.Contains(out, "fake") || !strings.HasSuffix(strings.Split(out, "\n")[2], " |") || len(m.Citations) != 1 || len(m.Unbound) != 1 || !m.RepairAttempted || calls != 1 {
		t.Fatalf("%s %+v", out, m)
	}
	if _, _, e = s.Finalize(context.Background(), out); e != nil || calls != 1 {
		t.Fatal("repair repeated")
	}
}

func TestOfferPinsFullTextBeforeEditAndNeverGuessesPages(t *testing.T) {
	s := requestFixture(t)
	db := s.store.DB
	for _, query := range []string{
		"CREATE TABLE chunks(id TEXT,knowledge_id TEXT,knowledge_base_id TEXT,content TEXT,chunk_type TEXT,metadata TEXT,tenant_id INTEGER,content_revision INTEGER,deleted_at DATETIME)",
		"CREATE TABLE knowledges(id TEXT,tenant_id INTEGER,knowledge_base_id TEXT,title TEXT,deleted_at DATETIME)",
		"CREATE TABLE archive_mapping_links(tenant_id INTEGER,knowledge_id TEXT,version_id TEXT)",
		"CREATE TABLE archive_mapping_versions(id TEXT,tenant_id INTEGER,knowledge_id TEXT,sha256 TEXT,file_name TEXT)",
		"INSERT INTO chunks VALUES('real','k','kb','original table','text','{}',2,1,NULL)",
		"INSERT INTO knowledges VALUES('k',2,'kb','document',NULL)",
	} {
		if e := db.Exec(query).Error; e != nil {
			t.Fatal(e)
		}
	}
	if e := s.Offer(context.Background(), "real", "k", "kb", "heading\noriginal table"); e != nil {
		t.Fatal(e)
	}
	if e := db.Exec("UPDATE chunks SET content='edited table',content_revision=2").Error; e != nil {
		t.Fatal(e)
	}
	out, m, e := s.Finalize(context.Background(), `Fact<kb chunk_id="real" />`)
	if e != nil {
		t.Fatal(e)
	}
	if len(m.Citations) != 1 || m.Citations[0].ContentRevision != 1 || m.Citations[0].Location.Status != PageMissing {
		t.Fatalf("%+v", m)
	}
	row, _, e := s.store.Evidence(context.Background(), 1, "s", "m", m.Citations[0].ID)
	if e != nil || row.Snapshot != "original table" || row.KnowledgeTenantID != 2 {
		t.Fatal("snapshot changed", e)
	}
	changed, e := s.store.Manifest(context.Background(), 1, "s", "m", out+"edited")
	if e != nil || changed.State != "answer_changed" || len(changed.Citations) != 0 {
		t.Fatal("stale answer binding")
	}
}

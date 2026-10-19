package archivalcitations

import (
	"context"
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestImmutableScopedEvidence(t *testing.T) {
	db, e := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&Record{}, &ManifestRecord{}); e != nil {
		t.Fatal(e)
	}
	s := NewStore(db)
	ctx := context.Background()
	sum := Summary{ID: EvidenceID(1, "s", "m", "c", "original", "v", 0), Number: 1, ChunkID: "c", KnowledgeID: "k", KnowledgeBaseID: "b", ContentSHA256: Hash("original"), ArchiveVersionID: "v", Location: Location{Pages: []int{}, Status: PageMissing}}
	m := Manifest{Version: 1, State: "verified", Citations: []Summary{sum}}
	rows := []Record{{ID: sum.ID, TenantID: 1, KnowledgeTenantID: 2, SessionID: "s", MessageID: "m", KnowledgeID: "k", ChunkID: "c", Snapshot: "original"}}
	if e = s.Save(ctx, 1, "s", "m", "answer", m, rows); e != nil {
		t.Fatal(e)
	}
	if e = s.Save(ctx, 1, "s", "m", "answer", m, rows); e != nil {
		t.Fatal(e)
	}
	if e = s.Save(ctx, 1, "s", "m", "changed answer", m, rows); !errors.Is(e, ErrImmutable) {
		t.Fatal(e)
	}
	for _, scope := range []struct {
		tenant           uint64
		session, message string
	}{{2, "s", "m"}, {1, "other", "m"}, {1, "s", "other"}} {
		if _, _, e = s.Evidence(ctx, scope.tenant, scope.session, scope.message, sum.ID); !errors.Is(e, gorm.ErrRecordNotFound) {
			t.Fatal("cross-scope disclosure")
		}
	}
	row, _, e := s.Evidence(ctx, 1, "s", "m", sum.ID)
	if e != nil || row.Snapshot != "original" || row.KnowledgeTenantID != 2 {
		t.Fatal("snapshot lost")
	}
	rows[0].Snapshot = "edited"
	if e = s.Save(ctx, 1, "s", "m2", "answer", m, rows); e == nil {
		t.Fatal("invalid evidence accepted")
	}
}

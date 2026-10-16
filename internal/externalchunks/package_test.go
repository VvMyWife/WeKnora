package externalchunks

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func fixture() map[string][]byte {
	md := []byte("# 测试\n原文")
	sum := sha256.Sum256(md)
	d, _ := json.Marshal(Document{ID: "doc", Title: "测试", RelativePath: "目录/测试.md", Version: 1, SourceSHA256: hex.EncodeToString(sum[:]), ParsedFile: "files/parsed.md"})
	c, _ := json.Marshal(Chunk{ID: "chunk", DocumentID: "doc", Order: 0, Content: "原文", RetrievalText: "原文", TokenCount: 2, SourceAnchors: []map[string]any{{"unit_id": "u1"}}})
	u, _ := json.Marshal(SourceUnit{ID: "doc:u1", DocumentID: "doc", UnitID: "u1", Text: "原文"})
	return map[string][]byte{"files/parsed.md": md, "documents.jsonl": append(d, '\n'), "chunks.jsonl": append(c, '\n'), "source_units.jsonl": append(u, '\n')}
}
func writePackage(t *testing.T, mutate func(map[string][]byte), corrupt bool) string {
	t.Helper()
	files := fixture()
	if mutate != nil {
		mutate(files)
	}
	m := Manifest{Format: Format, PackageID: "test", SourceSystem: "archive-rag", DatasetID: "fixture", GenerationID: "g1", GenerationConfig: json.RawMessage(`{}`), Counts: map[string]int{"documents": 1, "chunks": 1, "source_units": 1}, Files: map[string]Entry{}}
	for n, b := range files {
		h := sha256.Sum256(b)
		m.Files[n] = Entry{hex.EncodeToString(h[:]), int64(len(b))}
	}
	raw, _ := json.Marshal(m)
	files["manifest.json"] = raw
	if corrupt {
		files["chunks.jsonl"] = []byte("corrupt")
	}
	filename := filepath.Join(t.TempDir(), "fixture.zip")
	f, e := os.Create(filename)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	for n, b := range files {
		w, e := z.Create(n)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = w.Write(b); e != nil {
			t.Fatal(e)
		}
	}
	if e = z.Close(); e != nil {
		t.Fatal(e)
	}
	f.Close()
	return filename
}
func TestPackageValidation(t *testing.T) {
	p, e := Read(writePackage(t, nil, false))
	if e != nil {
		t.Fatal(e)
	}
	if p.Chunks[0].Content != "原文" {
		t.Fatal("content changed")
	}
	if _, e = Read(writePackage(t, nil, true)); e == nil {
		t.Fatal("accepted corrupt member")
	}
	for _, name := range []string{"../escape", "/etc/passwd", "files/../x", "x\\y", "a//b", "C:/x"} {
		t.Run(name, func(t *testing.T) {
			if _, e := Read(writePackage(t, func(f map[string][]byte) { f[name] = []byte("x") }, false)); e == nil {
				t.Fatal("accepted unsafe path")
			}
		})
	}
	for _, bad := range []string{
		`{"id":"c","document_id":"unknown","order":0,"content":"a","token_count":1}`,
		`{"id":"c","document_id":"doc","order":1,"content":"a","token_count":1}`,
		`{"id":"c","document_id":"doc","order":0,"content":"a","token_count":1,"source_anchors":[{"unit_id":"unknown"}]}`,
	} {
		if _, e := Read(writePackage(t, func(f map[string][]byte) { f["chunks.jsonl"] = []byte(bad + "\n") }, false)); e == nil {
			t.Fatal("accepted invalid chunk")
		}
	}
}
func TestCrossLanguageActualExport(t *testing.T) {
	filename := os.Getenv("EXTERNAL_CHUNKS_TEST_ZIP")
	if filename == "" {
		t.Skip("set EXTERNAL_CHUNKS_TEST_ZIP to a real RAG export")
	}
	p, e := Read(filename)
	if e != nil {
		t.Fatal(e)
	}
	t.Logf("accepted RAG export: %d documents, %d chunks, %d source units", len(p.Documents), len(p.Chunks), len(p.SourceUnits))
}
func TestStoreIdentityIsolationConflictAndResume(t *testing.T) {
	db, e := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&Batch{}, &Record{}, &Membership{}); e != nil {
		t.Fatal(e)
	}
	s := &Store{DB: db}
	p, e := Read(writePackage(t, nil, false))
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	b, rows, e := s.Plan(ctx, 1, "kb", "/private/test.zip", p)
	if e != nil {
		t.Fatal(e)
	}
	again, repeated, e := s.Plan(ctx, 1, "kb", "/another.zip", p)
	if e != nil || again.ID != b.ID || repeated[0].ID != rows[0].ID || again.PackagePath != "/private/test.zip" {
		t.Fatalf("not idempotent: %v", e)
	}
	_, other, e := s.Plan(ctx, 2, "kb", "/private/test.zip", p)
	if e != nil || other[0].KnowledgeID == rows[0].KnowledgeID {
		t.Fatal("cross-tenant ID collision")
	}
	if _, _, e = s.Batch(ctx, 2, "kb", b.ID); e == nil {
		t.Fatal("cross-tenant read")
	}
	p.Chunks[0].Content = "changed"
	p.ManifestHash = "changed-hash"
	if _, _, e = s.Plan(ctx, 1, "kb", "/changed.zip", p); e == nil {
		t.Fatal("allowed silent source overwrite")
	}
	p.Chunks[0].Content = "原文"
	p.ManifestHash = b.ManifestHash
	if e = s.Update(ctx, &rows[0], "failed", "embedding", "too long"); e != nil {
		t.Fatal(e)
	}
	_, actual, e := s.Batch(ctx, 1, "kb", b.ID)
	if e != nil || actual[0].Error != "too long" {
		t.Fatal("lost durable error")
	}
}

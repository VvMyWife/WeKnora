package externalchunks

import (
	"context"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"testing"
)

func TestRuntimeQuotaAndParsedAttachmentConflict(t *testing.T) {
	db, e := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&Batch{}, &Record{}, &Membership{}); e != nil {
		t.Fatal(e)
	}
	for _, sql := range []string{
		`CREATE TABLE tenants(id INTEGER PRIMARY KEY, storage_used INTEGER NOT NULL, storage_quota INTEGER NOT NULL, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE knowledges(id TEXT PRIMARY KEY, tenant_id INTEGER, knowledge_base_id TEXT, folder_path TEXT, file_name TEXT, deleted_at DATETIME)`,
		`INSERT INTO tenants(id,storage_used,storage_quota) VALUES(1,0,10000000),(2,0,1)`,
	} {
		if e = db.Exec(sql).Error; e != nil {
			t.Fatal(e)
		}
	}
	filename := writePackage(t, nil, false)
	p, e := Read(filename)
	if e != nil {
		t.Fatal(e)
	}
	s := &Store{DB: db, RuntimeChecks: true}
	ctx := context.Background()
	if _, _, e = s.Plan(ctx, 2, "kb", filename, p); e == nil {
		t.Fatal("quota bypass")
	}
	if _, _, e = s.Plan(ctx, 1, "kb", filename, p); e != nil {
		t.Fatal(e)
	}
	var used, again int64
	db.Raw("SELECT storage_used FROM tenants WHERE id=1").Scan(&used)
	if _, _, e = s.Plan(ctx, 1, "kb", filename, p); e != nil {
		t.Fatal(e)
	}
	db.Raw("SELECT storage_used FROM tenants WHERE id=1").Scan(&again)
	if used <= 0 || again != used {
		t.Fatal("duplicate quota charge")
	}
	entry := p.Manifest.Files[p.Documents[0].ParsedFile]
	entry.SHA256 = strings.Repeat("a", 64)
	p.Manifest.Files[p.Documents[0].ParsedFile] = entry
	p.ManifestHash = "new-manifest"
	if _, _, e = s.Plan(ctx, 1, "kb", filename, p); e == nil {
		t.Fatal("changed parsed preview silently reused source ID")
	}
}

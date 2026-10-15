package archivemapping

import (
	"context"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestPathRules(t *testing.T) {
	for _, bad := range []string{"../政策.pdf", "a/../b.pdf", "/a.pdf", "C:\\政策.pdf", "a//b.pdf", "a/./b.pdf", "a\x00.pdf"} {
		_, e := NormalizePath(bad, false)
		require.Error(t, e, bad)
	}
	p, e := NormalizePath("法规\\国家\\办法.PDF", false)
	require.NoError(t, e)
	require.Equal(t, "法规/国家/办法.PDF", p)
	require.Equal(t, MatchKey("法规/国家/办法.md"), MatchKey(p))
	require.NotEqual(t, MatchKey("另一个/办法.md"), MatchKey(p))
	_, ok := UnderRoot("政策_MD/法规/办法.md", "政策")
	require.False(t, ok)
}
func fixture(t *testing.T) *Store {
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, e)
	sql, e := db.DB()
	require.NoError(t, e)
	sql.SetMaxOpenConns(1)
	t.Cleanup(func() { sql.Close() })
	require.NoError(t, db.AutoMigrate(&Batch{}, &Item{}, &Version{}, &Link{}))
	require.NoError(t, db.Exec("CREATE TABLE tenants (id INTEGER PRIMARY KEY, storage_used INTEGER NOT NULL, storage_quota INTEGER NOT NULL, updated_at DATETIME, deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO tenants(id,storage_used,storage_quota) VALUES(1,0,1000)").Error)
	require.NoError(t, db.Exec("CREATE TABLE knowledges (id TEXT PRIMARY KEY, tenant_id INTEGER, knowledge_base_id TEXT, file_name TEXT, folder_path TEXT, file_type TEXT, title TEXT, deleted_at DATETIME)").Error)
	for _, row := range []struct{ id, folder string }{{"a", "政策_MD/甲"}, {"b", "政策_MD/乙"}} {
		require.NoError(t, db.Exec("INSERT INTO knowledges(id,tenant_id,knowledge_base_id,file_name,folder_path,file_type,title) VALUES (?,1,'kb','办法.md',?,'md','办法')", row.id, row.folder).Error)
	}
	return NewStore(db)
}
func TestPlanBindScopeVersionsAndIdempotence(t *testing.T) {
	s := fixture(t)
	ctx := context.Background()
	b, items, e := s.Plan(ctx, 1, "kb", "政策_MD", "u", []ManifestFile{{"甲/办法.pdf", 10}, {"乙/办法.pdf", 20}, {"未知.pdf", 1}})
	require.NoError(t, e)
	require.Equal(t, "a", items[0].KnowledgeID)
	require.Equal(t, "b", items[1].KnowledgeID)
	require.Equal(t, "unmatched", items[2].Status)
	calls := 0
	save := func(id string) (string, error) { calls++; return "resource://" + id, nil }
	v, e := s.Bind(ctx, 1, "kb", b.ID, items[0].ID, "", "u", "hash", 10, false, save)
	require.NoError(t, e)
	require.Equal(t, 1, calls)
	v2, e := s.Bind(ctx, 1, "kb", b.ID, items[0].ID, "", "u", "hash", 10, false, save)
	require.NoError(t, e)
	require.Equal(t, v.ID, v2.ID)
	require.Equal(t, 1, calls)
	_, _, e = s.Batch(ctx, 2, "kb", b.ID)
	require.Error(t, e)
	_, e = s.Bind(ctx, 1, "wrong", b.ID, items[1].ID, "", "u", "x", 20, false, save)
	require.Error(t, e)
	b2, it2, e := s.Plan(ctx, 1, "kb", "政策_MD", "u", []ManifestFile{{"甲/办法.pdf", 10}})
	require.NoError(t, e)
	require.Equal(t, "conflict", it2[0].Status)
	_, e = s.Bind(ctx, 1, "kb", b2.ID, it2[0].ID, "", "u", "new", 10, false, save)
	require.Error(t, e)
	newer, e := s.Bind(ctx, 1, "kb", b2.ID, it2[0].ID, "", "u", "new", 10, true, save)
	require.NoError(t, e)
	require.NotEqual(t, v.ID, newer.ID)
	require.NoError(t, s.Activate(ctx, 1, "kb", "a", v.ID))
	active, e := s.Active(ctx, 1, "a")
	require.NoError(t, e)
	require.Equal(t, v.ID, active.ID)
	versions, e := s.Versions(ctx, 1, "a")
	require.NoError(t, e)
	require.Len(t, versions, 2)
	require.NoError(t, s.Activate(ctx, 1, "kb", "a", ""))
	active, e = s.Active(ctx, 1, "a")
	require.NoError(t, e)
	require.Nil(t, active)
	var count int64
	require.NoError(t, s.DB.Model(&types.Knowledge{}).Count(&count).Error)
	require.EqualValues(t, 2, count)
}
func TestArchiveQuotaAndCleanup(t *testing.T) {
	s := fixture(t)
	ctx := context.Background()
	b, items, e := s.Plan(ctx, 1, "kb", "政策_MD", "u", []ManifestFile{{"甲/办法.pdf", 1001}})
	require.NoError(t, e)
	called := false
	_, e = s.Bind(ctx, 1, "kb", b.ID, items[0].ID, "", "u", "h", 1001, false, func(string) (string, error) { called = true; return "resource://x", nil })
	require.Error(t, e)
	require.False(t, called)
	var used int64
	require.NoError(t, s.DB.Table("tenants").Select("storage_used").Row().Scan(&used))
	require.Zero(t, used)
	b, items, e = s.Plan(ctx, 1, "kb", "政策_MD", "u", []ManifestFile{{"甲/办法.pdf", 20}})
	require.NoError(t, e)
	v, e := s.Bind(ctx, 1, "kb", b.ID, items[0].ID, "", "u", "h", 20, false, func(string) (string, error) { return "resource://x", nil })
	require.NoError(t, e)
	require.NoError(t, s.RemoveVersion(ctx, 1, "a", v.ID))
	active, e := s.Active(ctx, 1, "a")
	require.NoError(t, e)
	require.Nil(t, active)
	require.NoError(t, s.DB.Table("tenants").Select("storage_used").Row().Scan(&used))
	require.Zero(t, used)
}
func TestIndependentCopiesAndRuntimeMigrationGuard(t *testing.T) {
	s := fixture(t)
	ctx := context.Background()
	_, e := NewRuntimeStore(s.DB)
	require.NoError(t, e)
	v := Version{ID: "copied", KnowledgeID: "b", TenantID: 1, FileSize: 10, StoragePath: "resource://independent"}
	require.NoError(t, s.AddCopies(ctx, []Version{v}, v.ID))
	active, e := s.Active(ctx, 1, "b")
	require.NoError(t, e)
	require.Equal(t, v.StoragePath, active.StoragePath)
	// Removing a source that has no copy ownership cannot delete the copy.
	require.Error(t, s.RemoveVersion(ctx, 1, "a", v.ID))
	active, e = s.Active(ctx, 1, "b")
	require.NoError(t, e)
	require.NotNil(t, active)
	require.NoError(t, s.DB.Migrator().DropTable(&Link{}))
	_, e = NewRuntimeStore(s.DB)
	require.Error(t, e)
}

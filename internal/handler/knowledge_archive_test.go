package handler

import (
	"context"
	"fmt"
	"github.com/Tencent/WeKnora/internal/archivemapping"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type archiveFileStub struct {
	interfaces.KnowledgeService
	opened         string
	db             *gorm.DB
	saved, deleted int
}

func (s *archiveFileStub) GetKnowledgeByID(context.Context, string) (*types.Knowledge, error) {
	return &types.Knowledge{ID: "k", TenantID: 1, KnowledgeBaseID: "kb", FileType: "md"}, nil
}
func (s *archiveFileStub) GetKnowledgeFile(context.Context, string) (io.ReadCloser, string, error) {
	s.opened = "parsed"
	return io.NopCloser(strings.NewReader("# MD")), "法规.md", nil
}
func (s *archiveFileStub) SaveArchiveFile(ctx context.Context, _ string, _ *multipart.FileHeader, _, _ string) (string, error) {
	s.saved++
	if s.db != nil {
		if e := s.db.WithContext(ctx).Exec("INSERT INTO archive_test_objects (id) VALUES (?)", s.saved).Error; e != nil {
			return "", e
		}
	}
	return fmt.Sprintf("local://staged-%d", s.saved), nil
}
func (s *archiveFileStub) OpenArchiveFile(_ context.Context, _ string, p string) (io.ReadCloser, error) {
	s.opened = p
	return io.NopCloser(strings.NewReader("%PDF-")), nil
}
func (s *archiveFileStub) DeleteArchiveFile(context.Context, string, string) error {
	s.deleted++
	return nil
}
func TestSelectedKnowledgeFileKeepsInternalMDAndResolvesBrowserOriginal(t *testing.T) {
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, e)
	sql, _ := db.DB()
	defer sql.Close()
	require.NoError(t, db.AutoMigrate(&archivemapping.Link{}, &archivemapping.Version{}))
	require.NoError(t, db.Create(&archivemapping.Version{ID: "v", TenantID: 1, KnowledgeID: "k", FileName: "法规.pdf", StoragePath: "resource://original"}).Error)
	require.NoError(t, db.Create(&archivemapping.Link{KnowledgeID: "k", TenantID: 1, VersionID: "v"}).Error)
	files := &archiveFileStub{}
	h := &KnowledgeHandler{kgService: files, archiveStore: archivemapping.NewStore(db)}
	for _, test := range []struct {
		source, name, opened string
		fail                 bool
	}{{"auto", "法规.pdf", "resource://original", false}, {"parsed", "法规.md", "parsed", false}, {"archive", "法规.pdf", "resource://original", false}, {"invalid", "", "", true}} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/preview?source="+test.source, nil)
		stream, name, e := h.selectedKnowledgeFile(c, context.Background(), "k")
		if test.fail {
			require.Error(t, e)
			continue
		}
		require.NoError(t, e)
		stream.Close()
		require.Equal(t, test.name, name)
		require.Equal(t, test.opened, files.opened)
	}
	require.NoError(t, db.Where("knowledge_id = ?", "k").Delete(&archivemapping.Link{}).Error)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/preview?source=archive", nil)
	_, _, e = h.selectedKnowledgeFile(c, context.Background(), "k")
	require.Error(t, e)
}
func TestArchiveCatalogWritesOutsideSQLiteTransaction(t *testing.T) {
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, e)
	sql, _ := db.DB()
	sql.SetMaxOpenConns(1)
	defer sql.Close()
	require.NoError(t, db.AutoMigrate(&archivemapping.Batch{}, &archivemapping.Item{}, &archivemapping.Link{}, &archivemapping.Version{}))
	require.NoError(t, db.Exec("CREATE TABLE knowledges(id TEXT PRIMARY KEY,tenant_id INTEGER,knowledge_base_id TEXT,file_type TEXT,deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO knowledges VALUES('k',1,'kb','md',NULL)").Error)
	require.NoError(t, db.Exec("CREATE TABLE tenants(id INTEGER PRIMARY KEY,storage_used INTEGER,storage_quota INTEGER,updated_at DATETIME,deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO tenants(id,storage_used,storage_quota) VALUES(1,0,100)").Error)
	require.NoError(t, db.Exec("CREATE TABLE archive_test_objects(id INTEGER)").Error)
	require.NoError(t, db.Create(&archivemapping.Batch{ID: "batch", TenantID: 1, KnowledgeBaseID: "kb"}).Error)
	item := &archivemapping.Item{ID: "item", BatchID: "batch", RelativePath: "法规.pdf", FileSize: 3, KnowledgeID: "k", Status: "ready"}
	require.NoError(t, db.Create(item).Error)
	files := &archiveFileStub{db: db}
	h := &KnowledgeHandler{archiveStore: archivemapping.NewStore(db), kgService: files}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	original, e := h.bindArchiveUpload(ctx, 1, "kb", "batch", item, "k", "u", "hash", 3, false, &multipart.FileHeader{})
	require.NoError(t, e)
	require.Equal(t, 1, files.saved)
	repeat, e := h.bindArchiveUpload(ctx, 1, "kb", "batch", item, "k", "u", "hash", 3, false, &multipart.FileHeader{})
	require.NoError(t, e)
	require.Equal(t, original.ID, repeat.ID)
	require.Equal(t, 1, files.saved)
	_, e = h.bindArchiveUpload(ctx, 1, "kb", "batch", item, "k", "u", "different", 3, true, &multipart.FileHeader{})
	require.Error(t, e)
	require.Equal(t, 1, files.deleted)
	active, e := h.archiveStore.Active(ctx, 1, "k")
	require.NoError(t, e)
	require.Equal(t, original.ID, active.ID)
}

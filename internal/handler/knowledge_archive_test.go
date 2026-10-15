package handler

import (
	"context"
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
)

type archiveFileStub struct {
	interfaces.KnowledgeService
	opened string
}

func (s *archiveFileStub) GetKnowledgeByID(context.Context, string) (*types.Knowledge, error) {
	return &types.Knowledge{ID: "k", TenantID: 1, KnowledgeBaseID: "kb"}, nil
}
func (s *archiveFileStub) GetKnowledgeFile(context.Context, string) (io.ReadCloser, string, error) {
	s.opened = "parsed"
	return io.NopCloser(strings.NewReader("# MD")), "法规.md", nil
}
func (s *archiveFileStub) SaveArchiveFile(context.Context, string, *multipart.FileHeader, string) (string, error) {
	return "", nil
}
func (s *archiveFileStub) OpenArchiveFile(_ context.Context, _ string, p string) (io.ReadCloser, error) {
	s.opened = p
	return io.NopCloser(strings.NewReader("%PDF-")), nil
}
func (s *archiveFileStub) DeleteArchiveFile(context.Context, string, string) error { return nil }
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

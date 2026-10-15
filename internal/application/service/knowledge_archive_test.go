package service

import (
	"context"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/stretchr/testify/require"
	"mime/multipart"
	"testing"
)

type archiveKBStorageStub struct {
	interfaces.KnowledgeBaseService
}

func (archiveKBStorageStub) GetKnowledgeBaseByID(context.Context, string) (*types.KnowledgeBase, error) {
	return &types.KnowledgeBase{ID: "kb", TenantID: 1}, nil
}

type archiveSaveStorageStub struct {
	interfaces.FileService
	name, owner string
	tenant      uint64
}

func (s *archiveSaveStorageStub) SaveFile(_ context.Context, f *multipart.FileHeader, tenant uint64, owner string) (string, error) {
	s.name = f.Filename
	s.owner = owner
	s.tenant = tenant
	return "local://" + f.Filename, nil
}
func TestArchiveStorageKeepsRealOwnerAndUniqueVersionNames(t *testing.T) {
	files := &archiveSaveStorageStub{}
	s := &knowledgeService{kbService: archiveKBStorageStub{}, fileSvc: files}
	ctx := context.WithValue(context.Background(), types.TenantInfoContextKey, &types.Tenant{ID: 1})
	f := &multipart.FileHeader{Filename: "办法.pdf"}
	stored, e := s.SaveArchiveFile(ctx, "kb", f, "real-knowledge", "version-a")
	require.NoError(t, e)
	require.Equal(t, "real-knowledge", files.owner)
	require.EqualValues(t, 1, files.tenant)
	require.Equal(t, "local://version-a_办法.pdf", stored)
	require.Equal(t, "办法.pdf", f.Filename)
	next, e := s.SaveArchiveFile(ctx, "kb", f, "real-knowledge", "version-b")
	require.NoError(t, e)
	require.NotEqual(t, stored, next)
}

package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path"

	"github.com/Tencent/WeKnora/internal/archivemapping"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/google/uuid"
)

// AttachArchiveStore is a DI decorator; it leaves existing constructors and
// mocks compatible. Normal processing continues to use GetKnowledgeFile (MD).
func AttachArchiveStore(s interfaces.KnowledgeService, store *archivemapping.Store) interfaces.KnowledgeService {
	if concrete, ok := s.(*knowledgeService); ok {
		concrete.archiveStore = store
	}
	return s
}
func (s *knowledgeService) SaveArchiveFile(ctx context.Context, kbID string, f *multipart.FileHeader, knowledgeID, versionID string) (string, error) {
	kb, e := s.kbService.GetKnowledgeBaseByID(ctx, kbID)
	if e != nil {
		return "", e
	}
	ctx, e = s.archiveOwnerContext(ctx, kb.TenantID)
	if e != nil {
		return "", e
	}
	// The catalog owner is a real knowledge ID, while each physical object gets
	// a version-specific name even on providers that preserve uploaded filenames.
	header := *f
	header.Filename = versionID + "_" + path.Base(f.Filename)
	return s.resolveFileService(ctx, kb).SaveFile(ctx, &header, kb.TenantID, knowledgeID)
}
func (s *knowledgeService) OpenArchiveFile(ctx context.Context, kbID, stored string) (io.ReadCloser, error) {
	kb, e := s.kbService.GetKnowledgeBaseByID(ctx, kbID)
	if e != nil {
		return nil, e
	}
	ctx, e = s.archiveOwnerContext(ctx, kb.TenantID)
	if e != nil {
		return nil, e
	}
	return s.resolveFileServiceForPath(ctx, kb, stored).GetFile(ctx, stored)
}
func (s *knowledgeService) DeleteArchiveFile(ctx context.Context, kbID, stored string) error {
	kb, e := s.kbService.GetKnowledgeBaseByID(ctx, kbID)
	if e != nil {
		return e
	}
	ctx, e = s.archiveOwnerContext(ctx, kb.TenantID)
	if e != nil {
		return e
	}
	return s.resolveFileServiceForPath(ctx, kb, stored).DeleteFile(ctx, stored)
}
func (s *knowledgeService) archiveOwnerContext(ctx context.Context, tenantID uint64) (context.Context, error) {
	current, _ := ctx.Value(types.TenantInfoContextKey).(*types.Tenant)
	if current == nil || current.ID != tenantID {
		owner, e := s.tenantRepo.GetTenantByID(ctx, tenantID)
		if e != nil {
			return ctx, e
		}
		ctx = context.WithValue(ctx, types.TenantInfoContextKey, owner)
	}
	return context.WithValue(ctx, types.TenantIDContextKey, tenantID), nil
}
func (s *knowledgeService) cleanupArchiveFiles(ctx context.Context, k *types.Knowledge) {
	if s.archiveStore == nil {
		return
	}
	versions, e := s.archiveStore.Versions(ctx, k.TenantID, k.ID)
	if e != nil {
		logger.Errorf(ctx, "[ArchiveMapping] cleanup list knowledge=%s error=%v", k.ID, e)
		return
	}
	for _, v := range versions {
		if e = s.DeleteArchiveFile(ctx, k.KnowledgeBaseID, v.StoragePath); e != nil {
			logger.Errorf(ctx, "[ArchiveMapping] cleanup knowledge=%s version=%s error=%v", k.ID, v.ID, e)
			continue
		}
		if e = s.archiveStore.RemoveVersion(ctx, k.TenantID, k.ID, v.ID); e != nil {
			logger.Errorf(ctx, "[ArchiveMapping] cleanup metadata knowledge=%s version=%s error=%v", k.ID, v.ID, e)
		}
	}
}
func (s *knowledgeService) cloneArchiveFiles(ctx context.Context, src, dst *types.Knowledge, targetKB *types.KnowledgeBase) error {
	if s.archiveStore == nil {
		return nil
	}
	versions, e := s.archiveStore.Versions(ctx, src.TenantID, src.ID)
	if e != nil {
		return e
	}
	if len(versions) == 0 {
		return nil
	}
	active, e := s.archiveStore.Active(ctx, src.TenantID, src.ID)
	if e != nil {
		return e
	}
	sourceKB, e := s.kbService.GetKnowledgeBaseByID(ctx, src.KnowledgeBaseID)
	if e != nil {
		return e
	}
	var copied []archivemapping.Version
	activeID := ""
	for _, v := range versions {
		originalID := v.ID
		v.ID = uuid.NewString()
		v.TenantID = dst.TenantID
		v.KnowledgeID = dst.ID
		v.StoragePath, e = copyOwnedObject(ctx, s.resolveFileServiceForPath(ctx, sourceKB, v.StoragePath), s.resolveFileService(ctx, targetKB), v.StoragePath, dst.TenantID, dst.ID)
		if e != nil {
			for _, c := range copied {
				_ = s.DeleteArchiveFile(ctx, dst.KnowledgeBaseID, c.StoragePath)
			}
			return fmt.Errorf("clone archive: %w", e)
		}
		copied = append(copied, v)
		if active != nil && active.ID == originalID {
			activeID = v.ID
		}
	}
	if e = s.archiveStore.AddCopies(ctx, copied, activeID); e != nil {
		for _, v := range copied {
			_ = s.DeleteArchiveFile(ctx, dst.KnowledgeBaseID, v.StoragePath)
		}
		return e
	}
	return nil
}

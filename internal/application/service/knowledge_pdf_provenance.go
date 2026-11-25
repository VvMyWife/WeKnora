package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/archivemapping"
	"github.com/Tencent/WeKnora/internal/infrastructure/docparser"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/google/uuid"
)

// Pin only locators emitted by the parser reading this exact PDF.
func (s *knowledgeService) pinParsedPDF(ctx context.Context, knowledge *types.Knowledge, name string, data []byte, result *types.ReadResult) error {
	if s.archiveStore == nil || s.currentAttemptSuperseded(ctx, knowledge.ID) {
		return nil
	}
	pages, err := docparser.CountPDFPages(data)
	if err != nil {
		return nil
	}
	if !strings.EqualFold(path.Ext(name), ".pdf") {
		name = knowledge.ID + ".pdf"
	}
	sum := sha256.Sum256(data)
	staged := archivemapping.Version{
		ID: uuid.NewString(), TenantID: knowledge.TenantID, KnowledgeID: knowledge.ID,
		FileName: path.Base(name), RelativePath: path.Base(name), FileType: "pdf",
		FileSize: int64(len(data)), SHA256: hex.EncodeToString(sum[:]), CreatedBy: "native-pdf", CreatedAt: time.Now().UTC(),
	}
	err = withArchiveFile(bytes.NewReader(data), name, func(file *multipart.FileHeader) error {
		var err error
		staged.StoragePath, err = s.SaveArchiveFile(ctx, knowledge.KnowledgeBaseID, file, knowledge.ID, staged.ID)
		return err
	})
	if err != nil {
		return err
	}
	version, err := s.archiveStore.RegisterParsedPDF(ctx, knowledge.KnowledgeBaseID, knowledge.FileHash, staged)
	if err != nil || version.ID != staged.ID {
		if cleanupErr := s.DeleteArchiveFile(ctx, knowledge.KnowledgeBaseID, staged.StoragePath); err == nil && cleanupErr != nil {
			return cleanupErr
		}
	}
	if err != nil {
		return err
	}
	for i := range result.SourceBlocks {
		locator := &result.SourceBlocks[i].Locator
		if locator.Type == types.SourceLocatorPDF && locator.Page > 0 && locator.Page <= pages {
			locator.ArchiveVersionID = version.ID
		}
	}
	return nil
}

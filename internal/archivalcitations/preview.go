package archivalcitations

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Tencent/WeKnora/internal/types"
)

type PreviewLocation struct {
	Pages []int `json:"pages"`
	Hops  int   `json:"hops"`
}

type previewChunk struct {
	ID, PreChunkID, NextChunkID string
	ChunkIndex, ContentRevision int
	SourceLocators              types.SourceLocators `gorm:"type:json"`
	ExternalRecordID            string
	ExternalAnchors             string
}

func (s *Store) CitationPreview(ctx context.Context, tenant uint64, summary Summary) *PreviewLocation {
	if summary.SourceKind != "chunk" || summary.ArchiveVersionID == "" ||
		summary.ArchiveSHA256 == "" || !strings.EqualFold(filepath.Ext(summary.ArchiveFileName), ".pdf") {
		return nil
	}
	db := s.DB.WithContext(ctx)
	var archive struct{ SHA256 string }
	if db.Table("archive_mapping_versions").Select("sha256").
		Where("tenant_id = ? AND knowledge_id = ? AND id = ?", tenant, summary.KnowledgeID, summary.ArchiveVersionID).
		Take(&archive).Error != nil || archive.SHA256 != summary.ArchiveSHA256 {
		return nil
	}
	if summary.Location.Status == PageKnown && ValidateLocation(summary.Location) == nil {
		return &PreviewLocation{Pages: summary.Location.Pages}
	}
	if summary.Location.Status != PageMissing {
		return nil
	}
	load := func(id string) (previewChunk, bool) {
		var chunk previewChunk
		err := db.Table("chunks").Select(`id, pre_chunk_id, next_chunk_id, chunk_index, content_revision, source_locators,
			metadata -> 'external_import' ->> 'record_id' AS external_record_id,
			metadata -> 'external_import' ->> 'source_anchors' AS external_anchors`).
			Where("tenant_id = ? AND knowledge_id = ? AND knowledge_base_id = ? AND id = ? AND chunk_type = ? AND deleted_at IS NULL",
				tenant, summary.KnowledgeID, summary.KnowledgeBaseID, id, types.ChunkTypeText).Take(&chunk).Error
		return chunk, err == nil
	}
	first, ok := load(summary.ChunkID)
	if !ok || first.ContentRevision != summary.ContentRevision {
		return nil
	}

	var externalHash string
	if first.ExternalRecordID != "" {
		var record struct{ SourceHash string }
		payload := "payload"
		if db.Dialector.Name() == "postgres" {
			payload = "CAST(payload AS JSON)"
		}
		if db.Table("external_chunk_records").Select(payload+" -> 'document' ->> 'source_sha256' AS source_hash").
			Where("tenant_id = ? AND knowledge_id = ? AND id = ?", tenant, summary.KnowledgeID, first.ExternalRecordID).
			Take(&record).Error != nil || record.SourceHash != archive.SHA256 {
			return nil
		}
		externalHash = record.SourceHash
	}
	return nextPreview(first, load, func(chunk previewChunk) ([]int, bool) {
		if externalHash != "" && chunk.ExternalAnchors != "" {
			var anchors []map[string]any
			if json.Unmarshal([]byte(chunk.ExternalAnchors), &anchors) != nil {
				return nil, false
			}
			location := PagesFromAnchors(anchors, externalHash, archive.SHA256)
			if location.Status != PageMissing {
				return location.Pages, location.Status == PageKnown
			}
		}
		numberSet := map[int]struct{}{}
		for _, locator := range chunk.SourceLocators {
			if locator.Type != types.SourceLocatorPDF {
				continue
			}
			if locator.SourceHash != archive.SHA256 || locator.Page < 1 {
				return nil, false
			}
			numberSet[locator.Page] = struct{}{}
		}
		var pages []int
		for page := range numberSet {
			pages = append(pages, page)
		}
		sort.Ints(pages)
		return pages, true
	})
}

// Follow only forward links within the same parser generation.
func nextPreview(chunk previewChunk, load func(string) (previewChunk, bool), locate func(previewChunk) ([]int, bool)) *PreviewLocation {
	for hops := 0; hops <= 3; hops++ {
		pages, valid := locate(chunk)
		if !valid {
			return nil
		}
		if len(pages) > 0 {
			return &PreviewLocation{Pages: pages, Hops: hops}
		}
		if hops == 3 || chunk.NextChunkID == "" {
			return nil
		}
		next, ok := load(chunk.NextChunkID)
		if !ok || next.PreChunkID != chunk.ID || next.ChunkIndex <= chunk.ChunkIndex ||
			next.ExternalRecordID != chunk.ExternalRecordID {
			return nil
		}
		chunk = next
	}
	return nil
}

package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/archivemapping"
	"github.com/Tencent/WeKnora/internal/externalchunks"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AttachExternalChunkStore(s interfaces.KnowledgeService, store *externalchunks.Store) interfaces.KnowledgeService {
	if concrete, ok := s.(*knowledgeService); ok {
		concrete.externalStore = store
	}
	return s
}

// withExternalFile uses disk-backed multipart headers so storage providers keep
// their existing catalog ownership rules. No ZIP path is extracted to disk.
func withExternalFile(filename, member, name string, fn func(*multipart.FileHeader) error) error {
	z, e := zip.OpenReader(filename)
	if e != nil {
		return e
	}
	defer z.Close()
	var file *zip.File
	for _, f := range z.File {
		if f.Name == member {
			file = f
			break
		}
	}
	if file == nil {
		return fmt.Errorf("missing package member %s", member)
	}
	in, e := file.Open()
	if e != nil {
		return e
	}
	defer in.Close()
	temp, e := os.CreateTemp("", "external-member-*.multipart")
	if e != nil {
		return e
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	w := multipart.NewWriter(temp)
	part, e := w.CreateFormFile("file", name)
	if e != nil {
		return e
	}
	if _, e = io.Copy(part, in); e != nil {
		return e
	}
	if e = w.Close(); e != nil {
		return e
	}
	if _, e = temp.Seek(0, 0); e != nil {
		return e
	}
	form, e := multipart.NewReader(temp, w.Boundary()).ReadForm(8 << 20)
	if e != nil {
		return e
	}
	defer form.RemoveAll()
	return fn(form.File["file"][0])
}

func (s *knowledgeService) StartExternalImport(ctx context.Context, kbID, batchID, recordID string) (retErr error) {
	tenant := types.MustTenantIDFromContext(ctx)
	_, rows, e := s.externalStore.Batch(ctx, tenant, kbID, batchID)
	if e != nil {
		return e
	}
	var record *externalchunks.Record
	for i := range rows {
		if rows[i].ID == recordID {
			record = &rows[i]
			break
		}
	}
	if record == nil {
		return fmt.Errorf("record does not belong to this batch")
	}
	kb, e := s.kbService.GetKnowledgeBaseByID(ctx, kbID)
	if e != nil {
		return e
	}
	ctx, e = s.archiveOwnerContext(ctx, tenant)
	if e != nil {
		return e
	}
	// Postprocessing can finish with a summary failure after indexing succeeded.
	// Make that terminal failure explicitly retryable without treating a live
	// finalizing task as failed or enqueueing a second copy of active work.
	if record.Status == "postprocessing" {
		var current types.Knowledge
		q := s.externalStore.DB.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, record.KnowledgeID).First(&current)
		if q.Error == nil && (current.ParseStatus == types.ParseStatusFailed || (current.ParseStatus == types.ParseStatusCompleted && current.SummaryStatus == "failed")) {
			if e = s.externalStore.Update(ctx, record, "failed", "postprocess", "后处理失败，可显式重试"); e != nil {
				return e
			}
		} else {
			return nil
		}
	}
	// CAS prevents two upload windows from constructing the same knowledge.
	claimed := s.externalStore.DB.WithContext(ctx).Model(&externalchunks.Record{}).Where("id = ? AND tenant_id = ? AND (status IN ? OR (status = ? AND updated_at < ?))", record.ID, tenant, []string{"preview", "failed"}, "preparing", time.Now().Add(-30*time.Minute)).Updates(map[string]any{"status": "preparing", "error": "", "updated_at": time.Now().UTC()})
	if claimed.Error != nil {
		return claimed.Error
	}
	if claimed.RowsAffected == 0 {
		return nil
	}
	defer func() {
		if retErr != nil {
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			_ = s.externalStore.Update(cleanupCtx, record, "failed", "import", retErr.Error())
		}
	}()
	var payload externalchunks.DocumentPayload
	if e = json.Unmarshal([]byte(record.Payload), &payload); e != nil {
		return e
	}
	batch, _, e := s.externalStore.Batch(ctx, tenant, kbID, record.BatchID)
	if e != nil {
		return e
	}
	var manifest externalchunks.Manifest
	if e = json.Unmarshal([]byte(batch.ManifestJSON), &manifest); e != nil {
		return e
	}
	knowledge := &types.Knowledge{}
	e = s.externalStore.DB.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenant, record.KnowledgeID).First(knowledge).Error
	if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
		return e
	}
	if errors.Is(e, gorm.ErrRecordNotFound) {
		folder, file := types.SplitKnowledgeRelativePath(payload.Document.RelativePath)
		knowledge = &types.Knowledge{ID: record.KnowledgeID, TenantID: tenant, KnowledgeBaseID: kbID, Type: "file", Source: "external_chunks", Channel: "web", Title: payload.Document.Title, FileName: file, FolderPath: folder, FileType: "md", ParseStatus: "pending", SummaryStatus: "none", EnableStatus: "disabled", EmbeddingModelID: kb.EmbeddingModelID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		knowledge.Metadata, _ = json.Marshal(map[string]any{"external_import": map[string]string{"record_id": record.ID, "source_system": manifest.SourceSystem, "dataset_id": manifest.DatasetID, "generation_id": manifest.GenerationID, "document_id": record.ExternalDocumentID}})
		if e = s.repo.CreateKnowledge(ctx, knowledge); e != nil {
			return e
		}
	}
	if knowledge.Source != "external_chunks" || knowledge.KnowledgeBaseID != kbID {
		return fmt.Errorf("knowledge identity conflict")
	}
	if knowledge.ParseStatus == types.ParseStatusCompleted || knowledge.ParseStatus == types.ParseStatusFailed {
		if knowledge.StorageSize > 0 {
			if e = s.cleanupKnowledgeResources(ctx, knowledge); e != nil {
				return e
			}
		}
		resetKnowledgeForReparse(knowledge, kb)
		if e = s.repo.UpdateKnowledge(ctx, knowledge); e != nil {
			return e
		}
	}
	if knowledge.FilePath == "" {
		e = withExternalFile(batch.PackagePath, payload.Document.ParsedFile, knowledge.FileName, func(f *multipart.FileHeader) error {
			stored, err := s.resolveFileService(ctx, kb).SaveFile(ctx, f, tenant, knowledge.ID)
			if err == nil {
				knowledge.FilePath = stored
				knowledge.FileSize = f.Size
				err = s.repo.UpdateKnowledge(ctx, knowledge)
			}
			return err
		})
		if e != nil {
			return e
		}
	}
	if payload.Document.OriginalFile != "" {
		active, err := s.archiveStore.Active(ctx, tenant, knowledge.ID)
		if err != nil {
			return err
		}
		if active == nil {
			rel := strings.TrimSuffix(payload.Document.RelativePath, path.Ext(payload.Document.RelativePath)) + path.Ext(payload.Document.OriginalFile)
			entry := manifest.Files[payload.Document.OriginalFile]
			ab, items, err := s.archiveStore.Plan(ctx, tenant, kbID, "", "external-import", []archivemapping.ManifestFile{{RelativePath: rel, FileSize: entry.Size}})
			if err != nil {
				return err
			}
			err = withExternalFile(batch.PackagePath, payload.Document.OriginalFile, path.Base(rel), func(f *multipart.FileHeader) error {
				stored, saveErr := s.SaveArchiveFile(ctx, kbID, f, knowledge.ID, record.ID[:16])
				if saveErr != nil {
					return saveErr
				}
				version, bindErr := s.archiveStore.Bind(ctx, tenant, kbID, ab.ID, items[0].ID, knowledge.ID, "external-import", entry.SHA256, entry.Size, false, func(string) (string, error) { return stored, nil })
				if bindErr != nil || version == nil || version.StoragePath != stored {
					_ = s.DeleteArchiveFile(ctx, kbID, stored)
				}
				return bindErr
			})
			if err != nil {
				return err
			}
		} else if active.SHA256 != payload.Document.SourceSHA256 {
			return fmt.Errorf("original association conflict")
		}
	}
	return s.enqueueExternalDocument(ctx, knowledge, record, 0)
}

func (s *knowledgeService) enqueueExternalDocument(ctx context.Context, k *types.Knowledge, r *externalchunks.Record, attempt int) error {
	raw, e := json.Marshal(types.DocumentProcessPayload{TenantID: k.TenantID, KnowledgeID: k.ID, KnowledgeBaseID: k.KnowledgeBaseID, FileType: "md", FilePath: k.FilePath, FileName: k.FileName, Attempt: attempt})
	if e != nil {
		return e
	}
	_, e = s.task.Enqueue(asynq.NewTask(types.TypeDocumentProcess, raw, documentProcessTaskOptions(s.config, asynq.TaskID(fmt.Sprintf("external:%s:%s", r.ID, uuid.NewString())))...))
	if e != nil && !errors.Is(e, asynq.ErrTaskIDConflict) {
		return e
	}
	return s.externalStore.Update(ctx, r, "queued", "embedding", "")
}

func (s *knowledgeService) processExternalDocument(ctx context.Context, kb *types.KnowledgeBase, k *types.Knowledge) error {
	r, e := s.externalStore.ForKnowledge(ctx, k.TenantID, k.ID)
	if e != nil {
		return e
	}
	var payload externalchunks.DocumentPayload
	if e = json.Unmarshal([]byte(r.Payload), &payload); e != nil {
		return e
	}
	if e = s.checkExternalEmbeddingLimits(ctx, kb, k, &payload); e != nil {
		return e
	}
	parsed := make([]types.ParsedChunk, len(payload.Chunks))
	for i, ch := range payload.Chunks {
		parsed[i] = types.ParsedChunk{Content: ch.Content, Seq: ch.Order, ChunkID: externalchunks.MappedID(r.ID, ch.ID)}
	}
	s.skipStage(ctx, k.ID, types.StageDocReader, "external chunks: OCR and conversion bypassed")
	s.processChunks(ctx, kb, k, parsed, ProcessChunksOptions{ExternalPayload: &payload, ExternalRecordID: r.ID})
	if k.ParseStatus == types.ParseStatusFailed {
		return fmt.Errorf("%s", k.ErrorMessage)
	}
	if k.EnableStatus != "enabled" {
		return fmt.Errorf("external chunk indexing did not reach enabled state")
	}
	if e = s.externalStore.DB.WithContext(ctx).Model(&types.Chunk{}).Where("tenant_id = ? AND knowledge_id = ?", k.TenantID, k.ID).Update("index_status", "ready").Error; e != nil {
		return e
	}
	return s.externalStore.Update(ctx, r, "postprocessing", "postprocess", "")
}

// The configured adapter must expose its real tokenizer and prohibit implicit
// truncation. An estimate in the package is not proof that the provider fits.
func (s *knowledgeService) checkExternalEmbeddingLimits(ctx context.Context, kb *types.KnowledgeBase, k *types.Knowledge, p *externalchunks.DocumentPayload) error {
	if !kb.NeedsEmbeddingModel() {
		return nil
	}
	model, e := s.modelService.GetModelByID(ctx, kb.EmbeddingModelID)
	if e != nil {
		return e
	}
	if model.Parameters.EmbeddingParameters.TruncatePromptTokens > 0 {
		return fmt.Errorf("external imports require truncate_prompt_tokens=0; chunk content must not be truncated: %w", asynq.SkipRetry)
	}
	client := &http.Client{Timeout: 60 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	endpoint := strings.TrimRight(model.Parameters.BaseURL, "/") + "/token-counts"
	for start := 0; start < len(p.Chunks); start += 32 {
		end := min(start+32, len(p.Chunks))
		texts := []string{}
		for _, ch := range p.Chunks[start:end] {
			text := ch.RetrievalText
			if text == "" {
				text = ch.Content
			}
			texts = append(texts, buildKnowledgeIndexContent(k, text))
		}
		body, _ := json.Marshal(map[string]any{"model": model.Name, "input": texts})
		req, e := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if e != nil {
			return e
		}
		req.Header.Set("Content-Type", "application/json")
		if model.Parameters.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+model.Parameters.APIKey)
		}
		response, e := client.Do(req)
		if e != nil {
			return fmt.Errorf("embedding tokenizer preflight unavailable: %w", e)
		}
		raw, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		response.Body.Close()
		if readErr != nil {
			return readErr
		}
		if response.StatusCode != 200 {
			return fmt.Errorf("embedding tokenizer preflight HTTP %d; adapter must expose /v1/token-counts and disable silent truncation: %w", response.StatusCode, asynq.SkipRetry)
		}
		var limits struct {
			Maximum    int   `json:"max_input_tokens"`
			Counts     []int `json:"input_tokens"`
			Truncation bool  `json:"truncation"`
		}
		if e = json.Unmarshal(raw, &limits); e != nil || limits.Maximum < 1 || len(limits.Counts) != len(texts) || limits.Truncation {
			return fmt.Errorf("embedding tokenizer returned invalid limits: %w", asynq.SkipRetry)
		}
		for i, count := range limits.Counts {
			if count > limits.Maximum {
				return fmt.Errorf("document %s chunk %s requires %d tokens, embedding limit %d; original chunk retained, no truncation: %w", p.Document.ID, p.Chunks[start+i].ID, count, limits.Maximum, asynq.SkipRetry)
			}
		}
	}
	return nil
}

// Stable IDs must resurrect a soft-deleted imported chunk on retry; ordinary
// ingestion keeps its existing create behavior. Scope is checked before upsert.
func (s *knowledgeService) persistExternalChunks(ctx context.Context, chunks []*types.Chunk) error {
	return s.externalStore.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, c := range chunks {
			var old types.Chunk
			e := tx.Unscoped().Where("id = ?", c.ID).First(&old).Error
			if e == nil && (old.TenantID != c.TenantID || old.KnowledgeID != c.KnowledgeID) {
				return fmt.Errorf("external chunk identity belongs to another knowledge")
			}
			if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
				return e
			}
			if e == nil {
				c.SeqID = old.SeqID
			}
			c.SourceContent = c.Content
			c.IndexStatus = "processing"
		}
		if e := types.AssignChunkSeqIDs(tx, chunks); e != nil {
			return e
		}
		return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"content", "source_content", "context_header", "chunk_index", "is_enabled", "start_at", "end_at", "chunk_type", "metadata", "pre_chunk_id", "next_chunk_id", "updated_at", "deleted_at", "index_status"})}).CreateInBatches(chunks, 100).Error
	})
}

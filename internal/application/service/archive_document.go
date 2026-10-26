package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ArchiveDocumentContext exposes document-scoped provenance without returning file paths.
func (s *knowledgeService) ArchiveDocumentContext(ctx context.Context, id string) (map[string]any, error) {
	tenant := types.MustTenantIDFromContext(ctx)
	k, err := s.repo.GetKnowledgeByID(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	if s.externalStore == nil {
		return nil, fmt.Errorf("archive storage unavailable")
	}
	db := s.externalStore.DB.WithContext(ctx)
	var chunks []*types.Chunk
	if err := db.Where("tenant_id=? AND knowledge_id=? AND chunk_type=?", tenant, id, types.ChunkTypeText).Order("chunk_index").Find(&chunks).Error; err != nil {
		return nil, err
	}
	archive := map[string]any{"title": k.Title, "page_count": 0, "merged": false}
	if len(chunks) > 0 {
		var meta struct {
			External struct {
				Metadata struct {
					Archive map[string]any `json:"archive"`
				} `json:"metadata"`
			} `json:"external_import"`
		}
		if json.Unmarshal(chunks[0].Metadata, &meta) == nil && meta.External.Metadata.Archive != nil {
			archive = meta.External.Metadata.Archive
		}
	}
	var all []*types.WikiPage
	if err := db.Where("tenant_id=? AND knowledge_base_id=? AND status<>?", tenant, k.KnowledgeBaseID, types.WikiPageStatusArchived).
		Where("EXISTS (SELECT 1 FROM jsonb_array_elements_text(COALESCE(source_refs,'[]')::jsonb) r WHERE split_part(r,'|',1)=?)", id).Limit(500).Find(&all).Error; err != nil {
		return nil, err
	}
	pages := []map[string]any{}
	for _, p := range all {
		for _, source := range p.SourceKnowledgeIDs() {
			if source == id {
				pages = append(pages, map[string]any{"slug": p.Slug, "title": p.Title, "page_type": p.PageType, "summary": p.Summary, "content": p.Content, "chunk_refs": p.ChunkRefs})
				break
			}
		}
	}
	// Query latest delivery per logical stage, not cumulative historical failures.
	var spans []struct {
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	err = db.Raw(`SELECT DISTINCT ON(name) name,status FROM knowledge_processing_spans
      WHERE knowledge_id=? AND attempt=(SELECT COALESCE(MAX(attempt),0) FROM knowledge_processing_spans WHERE knowledge_id=?) ORDER BY name,updated_at DESC,id DESC`, id, id).Scan(&spans).Error
	if err != nil {
		return nil, err
	}
	graph := map[string]int{"done": 0, "failed": 0, "pending": 0}
	for _, span := range spans {
		if strings.HasPrefix(span.Name, "postprocess.graph.chunk[") {
			switch span.Status {
			case "done":
				graph["done"]++
			case "failed":
				graph["failed"]++
			default:
				graph["pending"]++
			}
		}
	}
	var graphData *types.GraphData
	var graphError string
	if source, ok := s.graphEngine.(interface {
		SearchArchiveGraph(context.Context, types.NameSpace) (*types.GraphSearchCandidates, error)
	}); ok {
		candidates, e := source.SearchArchiveGraph(ctx, types.NameSpace{KnowledgeBase: k.KnowledgeBaseID, Knowledge: id})
		if e == nil && candidates != nil {
			graphData = candidates.Graph
		} else if e != nil {
			graphError = "graph temporarily unavailable"
		}
	}
	state := "complete"
	if k.ParseStatus == types.ParseStatusPending || k.ParseStatus == types.ParseStatusProcessing || k.ParseStatus == types.ParseStatusFinalizing {
		state = "processing"
	}
	for _, span := range spans {
		if span.Status == "failed" {
			state = "partial"
			break
		}
	}
	if k.SummaryStatus == types.SummaryStatusFailed || k.ParseStatus == types.ParseStatusFailed {
		state = "partial"
	}
	return map[string]any{"knowledge_id": id, "archive": archive, "chunks": len(chunks), "wiki_pages": pages, "graph": graph, "graph_data": graphData, "graph_error": graphError, "status": state, "stages": spans}, nil
}

func archiveRetryStage(stage string) (string, error) {
	switch {
	case stage == "" || stage == "all":
		return "all", nil
	case strings.Contains(stage, "summary"):
		return "summary", nil
	case strings.Contains(stage, "graph"):
		return "graph", nil
	case strings.Contains(stage, "wiki"):
		return "wiki", nil
	case strings.Contains(stage, "embedding") || strings.Contains(stage, "docreader") || strings.Contains(stage, "chunk") || stage == "parse":
		return "parse", nil
	default:
		return "", fmt.Errorf("retry stage must be summary, graph, wiki or all")
	}
}

// RetryArchiveFailures admits a new manual round into the same durable outbox used by recovery.
func (s *knowledgeService) RetryArchiveFailures(ctx context.Context, id, requested string) (map[string]any, error) {
	stage, err := archiveRetryStage(requested)
	if err != nil {
		return nil, err
	}
	tenant := types.MustTenantIDFromContext(ctx)
	k, err := s.repo.GetKnowledgeByID(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	if k.ParseStatus == types.ParseStatusCancelled || k.ParseStatus == types.ParseStatusDeleting {
		return nil, fmt.Errorf("cancelled or deleting document cannot be retried")
	}
	recovery := enrichmentRecovery.Load()
	if recovery == nil {
		return nil, fmt.Errorf("durable recovery service unavailable")
	}
	attempt, err := latestEnrichmentAttempt(recovery.db.WithContext(ctx), id)
	if err != nil {
		return nil, err
	}
	var spans []struct {
		Name   string
		Status string
	}
	err = recovery.db.WithContext(ctx).Raw(`SELECT DISTINCT ON(name) name,status FROM knowledge_processing_spans WHERE knowledge_id=? AND attempt=? ORDER BY name,updated_at DESC,id DESC`, id, attempt).Scan(&spans).Error
	if err != nil {
		return nil, err
	}
	failed := map[string]bool{}
	for _, span := range spans {
		if span.Status == "failed" {
			failed[span.Name] = true
		}
	}
	candidates := map[string]EnrichmentRecoveryRow{}
	if (stage == "all" || stage == "summary") && k.SummaryStatus == types.SummaryStatusFailed {
		payload, _ := json.Marshal(types.SummaryGenerationPayload{TenantID: tenant, KnowledgeBaseID: k.KnowledgeBaseID, KnowledgeID: id, Attempt: attempt, Refresh: true})
		key := fmt.Sprintf("summary:%d:%s:%d", tenant, id, attempt)
		candidates[key] = EnrichmentRecoveryRow{LogicalKey: key, TaskType: types.TypeSummaryGeneration, Payload: payload}
	}
	var letters []types.TaskDeadLetter
	err = recovery.db.WithContext(ctx).Where("tenant_id=? AND (related_id=? OR (scope=? AND scope_id=?)) AND task_type IN ?", tenant, id, types.TaskScopeKnowledge, id, []string{types.TypeChunkExtract, types.TypeWikiIngest}).Order("id DESC").Find(&letters).Error
	if err != nil {
		return nil, err
	}
	for _, dl := range letters {
		var key string
		if dl.TaskType == types.TypeChunkExtract && (stage == "all" || stage == "graph") {
			var p types.ExtractChunkPayload
			if json.Unmarshal(dl.Payload, &p) != nil || p.Attempt != attempt || p.KnowledgeID != id || !failed[fmt.Sprintf("postprocess.graph.chunk[%d]", p.ChunkIndex)] {
				continue
			}
			key = fmt.Sprintf("graph:%d:%s:%d:%s", tenant, id, attempt, p.ChunkID)
		} else if dl.TaskType == types.TypeWikiIngest && (stage == "all" || stage == "wiki") && failed["postprocess.wiki"] {
			var p WikiPendingOp
			if json.Unmarshal(dl.Payload, &p) != nil || p.KnowledgeID != id {
				continue
			}
			key = fmt.Sprintf("wiki:%d:%s:%d", tenant, id, attempt)
		}
		if key != "" {
			if _, ok := candidates[key]; !ok {
				candidates[key] = EnrichmentRecoveryRow{LogicalKey: key, TaskType: dl.TaskType, Payload: dl.Payload, SourceDeadLetterID: dl.ID}
			}
		}
	}
	if len(candidates) == 0 && (stage == "all" || stage == "parse") && k.ParseStatus == types.ParseStatusFailed {
		_, err = s.ReparseKnowledge(ctx, id, nil)
		if err != nil {
			return nil, err
		}
		return map[string]any{"knowledge_id": id, "accepted": 1, "stage": "parse", "status": "queued_or_waiting"}, nil
	}
	accepted := 0
	for key, candidate := range candidates {
		candidate.TenantID = tenant
		candidate.KnowledgeBaseID = k.KnowledgeBaseID
		candidate.KnowledgeID = id
		candidate.Attempt = attempt
		candidate.State = "ready"
		candidate.AvailableAt = time.Now()
		candidate.UpdatedAt = time.Now()
		err = recovery.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if e := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&candidate).Error; e != nil {
				return e
			}
			var row EnrichmentRecoveryRow
			if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("logical_key=?", key).First(&row).Error; e != nil {
				return e
			}
			if row.State == "queued" || row.State == "skipped" {
				return nil
			}
			// A second click cannot replenish an already-admitted round's budget.
			if row.ID != candidate.ID && row.State == "ready" && !row.AvailableAt.After(time.Now()) {
				return nil
			}
			updates := map[string]any{"state": "ready", "available_at": time.Now(), "updated_at": time.Now()}
			if row.State == "exhausted" || row.State == "recovered" {
				updates["attempts"] = 0
			}
			if e := tx.Model(&row).Updates(updates).Error; e != nil {
				return e
			}
			accepted++
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	// Dispatch uses pause admission and persistent ownership; errors leave durable ready rows.
	if accepted > 0 {
		var ready []EnrichmentRecoveryRow
		if e := recovery.db.WithContext(ctx).Where("tenant_id=? AND knowledge_id=? AND state=?", tenant, id, "ready").Find(&ready).Error; e != nil {
			return nil, e
		}
		for _, row := range ready {
			_ = recovery.dispatch(ctx, row.ID)
		}
	}
	return map[string]any{"knowledge_id": id, "accepted": accepted, "status": "queued_or_waiting", "stage": stage}, nil
}

func (s *knowledgeService) ArchiveFailureCandidates(ctx context.Context, kb string) ([]string, error) {
	if s.externalStore == nil {
		return nil, fmt.Errorf("archive storage unavailable")
	}
	var ids []string
	err := s.externalStore.DB.WithContext(ctx).Raw(`SELECT k.id FROM knowledges k
      WHERE k.tenant_id=? AND k.knowledge_base_id=? AND k.deleted_at IS NULL AND
      (k.summary_status='failed' OR k.parse_status='failed' OR EXISTS(
        SELECT 1 FROM (SELECT DISTINCT ON(name) status FROM knowledge_processing_spans
          WHERE knowledge_id=k.id AND attempt=(SELECT COALESCE(MAX(attempt),0) FROM knowledge_processing_spans WHERE knowledge_id=k.id)
          ORDER BY name,updated_at DESC,id DESC) x WHERE x.status='failed')) ORDER BY k.id LIMIT 10000`, types.MustTenantIDFromContext(ctx), kb).Scan(&ids).Error
	return ids, err
}

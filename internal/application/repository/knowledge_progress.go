package repository

import (
	"context"
	"github.com/Tencent/WeKnora/internal/types"
	"sort"
	"strings"
	"time"
)

// Project only current attempts; never mutate old processing records.
func (r *knowledgeRepository) loadKnowledgeProgress(ctx context.Context, tenantID uint64, kbID string) ([]types.Knowledge, []types.KnowledgeProcessingSpan, error) {
	var docs []types.Knowledge
	err := r.db.WithContext(ctx).Select("id, file_name, folder_path, parse_status, summary_status, error_message, metadata, updated_at").Where("tenant_id = ? AND knowledge_base_id = ? AND parse_status <> ?", tenantID, kbID, types.ParseStatusDeleting).Find(&docs).Error
	if err != nil || len(docs) == 0 {
		return docs, nil, err
	}
	ids := make([]string, 0, len(docs))
	for _, d := range docs {
		ids = append(ids, d.ID)
	}
	var spans []types.KnowledgeProcessingSpan
	err = r.db.WithContext(ctx).Select("knowledge_id, attempt, span_id, name, kind, status, error_code, error_message, started_at, updated_at").Where("knowledge_id IN ? AND kind IN ? AND attempt = (SELECT MAX(s.attempt) FROM knowledge_processing_spans s WHERE s.knowledge_id = knowledge_processing_spans.knowledge_id)", ids, []string{types.SpanKindStage, types.SpanKindSubSpan}).Order("updated_at DESC, id DESC").Find(&spans).Error
	return docs, spans, err
}

// Public reasons contain no provider response, request body, URL or credentials.
func safeProgressReason(reason, code string) string {
	if strings.Contains(reason, "PDF 共 ") {
		i := strings.Index(reason, "PDF 共 ")
		return reason[i:]
	}
	if strings.Contains(reason, "未获得可索引正文") {
		return "解析完成但未获得可索引正文"
	}
	if strings.Contains(reason, "PDF 无法打开") || strings.Contains(reason, "PDF 没有") {
		return "PDF 无法打开或没有可读取页面，请检查文件是否损坏或加密"
	}
	if strings.Contains(reason, "未配置摘要模型") {
		return "未配置摘要模型，无法生成摘要"
	}
	lower := strings.ToLower(reason)
	switch {
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline"):
		return "服务请求超时，请查看文档处理诊断"
	case strings.Contains(lower, "eof") || strings.Contains(lower, "connection") || strings.Contains(lower, "broken pipe"):
		return "解析服务连接中断，系统将按重试状态恢复"
	case strings.Contains(lower, "cancel"):
		return "处理已取消或工作进程已退出"
	default:
		return "处理失败，请进入文档详情查看阶段诊断（" + code + "）"
	}
}

func projectKnowledgeProgress(docs []types.Knowledge, spans []types.KnowledgeProcessingSpan) (*types.KnowledgeProcessingOverview, []types.KnowledgeProcessingFailure) {
	out := &types.KnowledgeProcessingOverview{Stages: map[string]map[string]int64{}, FailureCounts: map[string]int64{}, ActiveDocuments: []types.KnowledgeProcessingFailure{}}
	failures := []types.KnowledgeProcessingFailure{}
	byDoc := map[string][]types.KnowledgeProcessingSpan{}
	seen := map[string]bool{}
	for _, s := range spans {
		key := s.KnowledgeID + "/" + s.Name
		if seen[key] {
			continue
		}
		seen[key] = true
		byDoc[s.KnowledgeID] = append(byDoc[s.KnowledgeID], s)
	}
	increment := func(stage, status string) {
		if out.Stages[stage] == nil {
			out.Stages[stage] = map[string]int64{}
		}
		out.Stages[stage][status]++
	}
	for _, d := range docs {
		meta := d.GetMetadata()
		item := types.KnowledgeProcessingFailure{ID: d.ID, FileName: d.FileName, FolderPath: d.FolderPath, ParseStatus: d.ParseStatus, SummaryStatus: d.SummaryStatus, PDFIntegrity: meta["pdf_page_integrity"], ExpectedPages: meta["pdf_expected_pages"], ProcessedPages: meta["pdf_processed_pages"], UpdatedAt: d.UpdatedAt, RetryStatus: "retry_available"}
		active := d.ParseStatus == "pending" || d.ParseStatus == "processing" || d.ParseStatus == "finalizing"
		if active {
			item.RetryStatus = "retrying_or_processing"
		}
		var failed *types.KnowledgeProcessingSpan
		for _, s := range byDoc[d.ID] {
			if s.Kind == types.SpanKindStage {
				status := s.Status
				// Historical open spans on terminal successful documents are not live work.
				if !active && status == types.SpanStatusRunning {
					continue
				}
				increment(s.Name, status)
				if active && status == types.SpanStatusRunning {
					item.Stage = s.Name
					item.Attempt = s.Attempt
					if s.StartedAt != nil {
						item.ElapsedSeconds = int64(time.Since(*s.StartedAt).Seconds())
					}
				}
			}
			if s.Status == types.SpanStatusFailed && (d.ParseStatus == "failed" || d.SummaryStatus == "failed" || active) && failed == nil {
				copy := s
				failed = &copy
			}
		}
		if d.SummaryStatus != "" && d.SummaryStatus != "none" {
			increment("summary", d.SummaryStatus)
		}
		if started, err := time.Parse(time.RFC3339, meta["pdf_started_at"]); err == nil && d.ParseStatus == "processing" {
			item.ElapsedSeconds = int64(time.Since(started).Seconds())
		}
		if active {
			out.ActiveDocuments = append(out.ActiveDocuments, item)
		}
		if failed != nil || d.ParseStatus == "failed" || d.SummaryStatus == "failed" {
			item.Stage = "docreader"
			item.ErrorCode = "processing_failed"
			reason := d.ErrorMessage
			if failed != nil {
				item.Stage = failed.Name
				item.Attempt = failed.Attempt
				item.ErrorCode = failed.ErrorCode
				item.UpdatedAt = failed.UpdatedAt
				if reason == "" {
					reason = failed.ErrorMessage
				}
			}
			if d.SummaryStatus == "failed" && d.ParseStatus != "failed" {
				item.Stage = "summary"
			}
			item.Reason = safeProgressReason(reason, item.ErrorCode)
			failures = append(failures, item)
			out.FailureCounts[item.Stage]++
			out.FailuresTotal++
		}
	}
	sort.Slice(failures, func(i, j int) bool {
		if failures[i].UpdatedAt.Equal(failures[j].UpdatedAt) {
			return failures[i].ID > failures[j].ID
		}
		return failures[i].UpdatedAt.After(failures[j].UpdatedAt)
	})
	return out, failures
}

func (r *knowledgeRepository) KnowledgeProgressDetails(ctx context.Context, tenantID uint64, kbID string) (*types.KnowledgeProcessingOverview, error) {
	docs, spans, err := r.loadKnowledgeProgress(ctx, tenantID, kbID)
	if err != nil {
		return nil, err
	}
	overview, _ := projectKnowledgeProgress(docs, spans)
	return overview, nil
}

func (r *knowledgeRepository) ListKnowledgeProcessingFailures(ctx context.Context, tenantID uint64, kbID string, page *types.Pagination, stage ...string) ([]types.KnowledgeProcessingFailure, int64, error) {
	docs, spans, err := r.loadKnowledgeProgress(ctx, tenantID, kbID)
	if err != nil {
		return nil, 0, err
	}
	_, items := projectKnowledgeProgress(docs, spans)
	if len(stage) > 0 && stage[0] != "" {
		filtered := []types.KnowledgeProcessingFailure{}
		for _, item := range items {
			if item.Stage == stage[0] {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	total := int64(len(items))
	start := page.Offset()
	if start >= len(items) {
		return []types.KnowledgeProcessingFailure{}, total, nil
	}
	end := start + page.Limit()
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

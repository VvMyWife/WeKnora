package archivalcitations

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/WeKnora/internal/logger"
)

type scopeKey struct{}
type offeredEvidence struct {
	summary Summary
	record  Record
}
type Scope struct {
	mu               sync.Mutex
	store            *Store
	tenant           uint64
	session, message string
	evidence         map[string]offeredEvidence
	captureFailures  map[string]bool
	active           bool
	finalAnswer      string
	finalManifest    *Manifest
}

func WithScope(ctx context.Context, store *Store, tenant uint64, session, message string) context.Context {
	if !Enabled() || store == nil {
		return ctx
	}
	return context.WithValue(ctx, scopeKey{}, &Scope{store: store, tenant: tenant, session: session, message: message, evidence: map[string]offeredEvidence{}})
}
func FromContext(ctx context.Context) *Scope { s, _ := ctx.Value(scopeKey{}).(*Scope); return s }

// DiagnosticMessageID is immutable request metadata, not evidence content.
func (s *Scope) DiagnosticMessageID() string {
	if s == nil {
		return ""
	}
	return s.message
}

// NoteCaptureFailure records registration failures without storing source text.
func (s *Scope) NoteCaptureFailure(chunk string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	if s.captureFailures == nil {
		s.captureFailures = map[string]bool{}
	}
	s.captureFailures[chunk] = true
}

// Offer is called before an LLM request, only for content-bearing source rows.
// History-only IDs are not eligible. The snapshot is pinned before generation.
func (s *Scope) Offer(ctx context.Context, chunk, knowledge, kb, provided string) error {
	if s == nil || strings.TrimSpace(provided) == "" || chunk == "" || knowledge == "" || kb == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	if existing, ok := s.evidence[chunk]; ok {
		if existing.summary.KnowledgeBaseID != kb || existing.summary.KnowledgeID != knowledge {
			return fmt.Errorf("chunk source scope changed")
		}
		return nil
	}
	var row struct {
		ID, KnowledgeID, KnowledgeBaseID, Content, ChunkType, Metadata, RelationChunks string
		TenantID                                                                       uint64
		ContentRevision                                                                int
	}
	e := s.store.DB.WithContext(ctx).Table("chunks").Where("id=? AND knowledge_id=? AND knowledge_base_id=? AND deleted_at IS NULL", chunk, knowledge, kb).Take(&row).Error
	if e != nil {
		return e
	}
	var original struct {
		TenantID               uint64
		KnowledgeBaseID, Title string
	}
	if e = s.store.DB.WithContext(ctx).Table("knowledges").Where("id=? AND tenant_id=? AND deleted_at IS NULL", knowledge, row.TenantID).Take(&original).Error; e != nil {
		return e
	}
	if original.KnowledgeBaseID != kb {
		return fmt.Errorf("citation source mismatch")
	}
	snapshot := row.Content
	kind := "chunk"
	revision := row.ContentRevision
	if !strings.Contains(provided, row.Content) || row.Content == "" {
		snapshot = provided
		kind = "excerpt"
		revision = -1
	}
	location := Location{Status: PageMissing, Pages: []int{}, Reason: "未提供可信物理页码"}
	version := ""
	origins := []SourceOrigin{}
	var archive struct{ ID, SHA256, FileName string }
	// Absence of an original is valid, but a database error is not a version choice.
	result := s.store.DB.WithContext(ctx).Table("archive_mapping_links AS l").Select("v.id,v.sha256,v.file_name").Joins("JOIN archive_mapping_versions v ON v.id=l.version_id AND v.tenant_id=l.tenant_id AND v.knowledge_id=l.knowledge_id").Where("l.tenant_id=? AND l.knowledge_id=?", row.TenantID, knowledge).Scan(&archive)
	if result.Error != nil {
		return result.Error
	}
	version = archive.ID
	if kind == "chunk" && row.ChunkType != "wiki_page" && row.ChunkType != "entity" && row.ChunkType != "relationship" && row.ChunkType != "summary" {
		var metadata struct {
			External struct {
				RecordID string           `json:"record_id"`
				Anchors  []map[string]any `json:"source_anchors"`
			} `json:"external_import"`
		}
		if json.Unmarshal([]byte(row.Metadata), &metadata) == nil && metadata.External.RecordID != "" {
			var external struct{ Payload string }
			if e = s.store.DB.WithContext(ctx).Table("external_chunk_records").Where("id=? AND tenant_id=? AND knowledge_id=?", metadata.External.RecordID, row.TenantID, knowledge).Take(&external).Error; e != nil {
				return e
			}
			var payload struct {
				Document struct {
					SHA256 string `json:"source_sha256"`
				} `json:"document"`
			}
			if json.Unmarshal([]byte(external.Payload), &payload) == nil {
				location = PagesFromAnchors(metadata.External.Anchors, payload.Document.SHA256, archive.SHA256)
			}
		}
	} else if row.ChunkType == "wiki_page" || row.ChunkType == "entity" || row.ChunkType == "relationship" || row.ChunkType == "summary" {
		kind = "derived"
		location.Reason = "派生内容，尚无可验证的原始页来源"
		version = ""
		archive.FileName = ""
		archive.SHA256 = ""
		var related []string
		_ = json.Unmarshal([]byte(row.RelationChunks), &related)
		origins, e = s.sourceOrigins(ctx, row.TenantID, kb, []string{knowledge}, related)
		if e != nil {
			return e
		}
	}
	if strings.TrimSpace(original.Title) == "" {
		original.Title = knowledge
	}
	sum := Summary{ID: EvidenceID(s.tenant, s.session, s.message, chunk, snapshot, version, revision), ChunkID: chunk, KnowledgeID: knowledge, KnowledgeBaseID: kb, Title: original.Title, ContentRevision: revision, ContentSHA256: Hash(snapshot), ArchiveVersionID: version, ArchiveFileName: archive.FileName, ArchiveSHA256: archive.SHA256, SourceKind: kind, Location: location, Origins: origins}
	s.evidence[chunk] = offeredEvidence{sum, Record{ID: sum.ID, TenantID: s.tenant, KnowledgeTenantID: row.TenantID, SessionID: s.session, MessageID: s.message, KnowledgeID: knowledge, ChunkID: chunk, Snapshot: snapshot}}
	return nil
}

var kbTag = regexp.MustCompile(`(?is)<kb\b[^>]*>`)
var chunkAttribute = regexp.MustCompile(`(?i)\bchunk_id\s*=\s*"([^"]+)"`)
var webTag = regexp.MustCompile(`(?is)<web\b[^>]*>`)
var tableDivider = regexp.MustCompile(`^\s*\|?[\s:|-]+\|[\s:|-]*$`)

func citationChunk(tag string) string {
	m := chunkAttribute.FindStringSubmatch(tag)
	if len(m) == 2 {
		return html.UnescapeString(m[1])
	}
	return ""
}

// Finalize validates generation-time bindings locally, with no model or retrieval calls, and
// atomically stores only sources actually cited by the final answer.
func (s *Scope) Finalize(ctx context.Context, answer string) (string, *Manifest, error) {
	if s == nil {
		return answer, nil, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	started := time.Now()
	if s.finalManifest != nil {
		if answer == s.finalAnswer {
			return s.finalAnswer, s.finalManifest, nil
		}
		return answer, nil, ErrImmutable
	}
	if !s.active && len(s.evidence) == 0 {
		manifest := Manifest{Version: SchemaVersion, State: "not_applicable", Citations: []Summary{}, Unbound: []UnboundBlock{}}
		if e := s.store.Save(ctx, s.tenant, s.session, s.message, answer, manifest, nil); e != nil {
			return answer, nil, e
		}
		s.finalAnswer = answer
		s.finalManifest = &manifest
		return answer, &manifest, nil
	}
	invalidLines := map[int]bool{}
	invalidCount := 0
	lineIndex := -1
	for _, line := range strings.Split(answer, "\n") {
		lineIndex++
		_ = transformCitationTags(line, func(tag string) string {
			if _, ok := s.evidence[citationChunk(tag)]; !ok {
				invalidLines[lineIndex] = true
				invalidCount++
			}
			return tag
		})
	}
	answer = transformCitationTags(answer, func(tag string) string {
		if _, ok := s.evidence[citationChunk(tag)]; ok {
			return tag
		}
		return ""
	})
	manifest := Manifest{Version: SchemaVersion, State: "verified", Citations: []Summary{}, Unbound: []UnboundBlock{}, RepairAttempted: false}
	records := []Record{}
	numbers := map[string]int{}
	answer = transformCitationTags(answer, func(tag string) string {
		ev, ok := s.evidence[citationChunk(tag)]
		if !ok {
			return ""
		}
		n, ok := numbers[ev.summary.ID]
		if !ok {
			n = len(manifest.Citations) + 1
			numbers[ev.summary.ID] = n
			ev.summary.Number = n
			manifest.Citations = append(manifest.Citations, ev.summary)
			records = append(records, ev.record)
		}
		return fmt.Sprintf(`<kb doc="%s" chunk_id="%s" kb_id="%s" citation_id="%s" footnote="%d" />`, html.EscapeString(ev.summary.Title), html.EscapeString(ev.summary.ChunkID), html.EscapeString(ev.summary.KnowledgeBaseID), ev.summary.ID, n)
	})
	lines := strings.Split(answer, "\n")
	for _, block := range answerBlocks(answer) {
		binding := BlockBinding{Index: block.start, EndIndex: block.end, SHA256: Hash(block.text), Kind: block.kind, CitationIDs: []string{}, State: "bound"}
		seen := map[string]bool{}
		for _, tag := range citationTagsOutsideCode(block.text) {
			if ev, ok := s.evidence[citationChunk(tag)]; ok && !seen[ev.summary.ID] {
				binding.CitationIDs = append(binding.CitationIDs, ev.summary.ID)
				seen[ev.summary.ID] = true
			}
		}
		reason := ""
		if len(binding.CitationIDs) == 0 && !webTag.MatchString(inlineCode.ReplaceAllString(block.text, "")) {
			reason = "missing_reference"
		}
		if hasUncitedTail(block.text) {
			reason = "missing_reference"
		}
		for i := block.start; i <= block.end; i++ {
			if invalidLines[i] {
				reason = "invalid_reference"
			}
		}
		if reason == "missing_reference" && len(s.captureFailures) > 0 && len(s.evidence) == 0 {
			reason = "evidence_capture_failed"
		}
		if reason != "" {
			binding.State = "needs_review"
			binding.ReasonCode = reason
			manifest.Unbound = append(manifest.Unbound, UnboundBlock{Index: block.start, EndIndex: block.end, SHA256: Hash(block.text), Reason: reviewReason(reason), ReasonCode: reason})
		}
		manifest.Bindings = append(manifest.Bindings, binding)
	}
	if len(manifest.Unbound) > 0 {
		manifest.State = "needs_review"
		for _, block := range manifest.Unbound {
			line := lines[block.EndIndex]
			marker := fmt.Sprintf(" <span class=\"citation-unbound\" title=\"%s\">〔引用待核验〕</span>", html.EscapeString(block.Reason))
			if strings.HasSuffix(strings.TrimSpace(line), "|") {
				line = strings.TrimSuffix(strings.TrimRight(line, " \t"), "|") + marker + " |"
			} else {
				line += marker
			}
			lines[block.EndIndex] = line
		}
		answer = strings.Join(lines, "\n")
	}
	validated := time.Now()
	if e := s.store.Save(ctx, s.tenant, s.session, s.message, answer, manifest, records); e != nil {
		return answer, &manifest, e
	}
	logger.Infof(ctx, "[Citations] stage=finalize session=%s message=%s evidence=%d valid=%d invalid=%d capture_failed=%d review=%d repair_calls=0 validation_ms=%d save_ms=%d", s.session, s.message, len(s.evidence), len(manifest.Citations), invalidCount, len(s.captureFailures), len(manifest.Unbound), validated.Sub(started).Milliseconds(), time.Since(validated).Milliseconds())
	s.finalAnswer = answer
	s.finalManifest = &manifest
	return answer, &manifest, nil
}

var inlineCode = regexp.MustCompile("`+[^`\\n]*`+")

func transformCitationTags(answer string, replace func(string) string) string {
	lines := strings.Split(answer, "\n")
	fence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fence = !fence
			continue
		}
		if fence {
			continue
		}
		var out strings.Builder
		offset := 0
		for _, span := range inlineCode.FindAllStringIndex(line, -1) {
			out.WriteString(kbTag.ReplaceAllStringFunc(line[offset:span[0]], replace))
			out.WriteString(line[span[0]:span[1]])
			offset = span[1]
		}
		out.WriteString(kbTag.ReplaceAllStringFunc(line[offset:], replace))
		lines[i] = out.String()
	}
	return strings.Join(lines, "\n")
}

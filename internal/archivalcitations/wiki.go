package archivalcitations

import (
	"context"
	"fmt"
	"strings"
)

// Wiki evidence has its own entity ID; it never claims to be a source PDF chunk.
// The generic chunk_id wire field holds that page ID only with wiki_page_id set.
func (s *Scope) OfferWiki(ctx context.Context, page, kb, provided string, version int, sourceRefs, chunkRefs []string) error {
	if s == nil || page == "" || kb == "" || strings.TrimSpace(provided) == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = true
	if existing, ok := s.evidence[page]; ok {
		if existing.summary.KnowledgeBaseID != kb {
			return fmt.Errorf("wiki source scope changed")
		}
		return nil
	}
	var row struct {
		ID, KnowledgeBaseID, Title string
		TenantID                   uint64
	}
	if e := s.store.DB.WithContext(ctx).Table("wiki_pages").Where("id=? AND knowledge_base_id=? AND deleted_at IS NULL", page, kb).Take(&row).Error; e != nil {
		return e
	}
	if row.TenantID == 0 {
		return fmt.Errorf("wiki source tenant missing")
	}
	origins, e := s.sourceOrigins(ctx, row.TenantID, kb, sourceRefs, chunkRefs)
	if e != nil {
		return e
	}
	if version < 1 {
		version = -1
	}
	summary := Summary{ID: EvidenceID(s.tenant, s.session, s.message, page, provided, "", version), WikiPageID: page, ChunkID: page, KnowledgeBaseID: kb, Title: row.Title, ContentRevision: version, ContentSHA256: Hash(provided), SourceKind: "derived", Origins: origins, Location: Location{Status: PageMissing, Pages: []int{}, Reason: "Wiki 派生页面，来源关系不等同于原始 PDF 页码"}}
	s.evidence[page] = offeredEvidence{summary, Record{ID: summary.ID, TenantID: s.tenant, KnowledgeTenantID: row.TenantID, SessionID: s.session, MessageID: s.message, ChunkID: page, Snapshot: provided}}
	return nil
}
func (s *Scope) sourceOrigins(ctx context.Context, tenant uint64, kb string, sourceRefs, chunkRefs []string) ([]SourceOrigin, error) {
	ids := []string{}
	seen := map[string]bool{}
	for _, ref := range sourceRefs {
		if id := strings.TrimSpace(strings.SplitN(ref, "|", 2)[0]); id != "" && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	type sourceChunk struct{ ID, KnowledgeID string }
	chunks := []sourceChunk{}
	if len(chunkRefs) > 0 {
		if e := s.store.DB.WithContext(ctx).Table("chunks").Select("id,knowledge_id").Where("id IN ? AND tenant_id=? AND knowledge_base_id=? AND deleted_at IS NULL", chunkRefs, tenant, kb).Find(&chunks).Error; e != nil {
			return nil, e
		}
		for _, chunk := range chunks {
			if !seen[chunk.KnowledgeID] {
				seen[chunk.KnowledgeID] = true
				ids = append(ids, chunk.KnowledgeID)
			}
		}
	}
	origins := []SourceOrigin{}
	if len(ids) == 0 {
		return origins, nil
	}
	var docs []struct{ ID, Title string }
	if e := s.store.DB.WithContext(ctx).Table("knowledges").Select("id,title").Where("id IN ? AND tenant_id=? AND knowledge_base_id=? AND deleted_at IS NULL", ids, tenant, kb).Order("id").Find(&docs).Error; e != nil {
		return nil, e
	}
	for _, doc := range docs {
		origin := SourceOrigin{KnowledgeID: doc.ID, Title: doc.Title, ChunkIDs: []string{}}
		for _, chunk := range chunks {
			if chunk.KnowledgeID == doc.ID {
				origin.ChunkIDs = append(origin.ChunkIDs, chunk.ID)
			}
		}
		origins = append(origins, origin)
	}
	return origins, nil
}

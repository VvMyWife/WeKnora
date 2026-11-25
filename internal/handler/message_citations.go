package handler

import (
	"errors"
	"github.com/Tencent/WeKnora/internal/archivalcitations"
	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func AttachMessageCitationStore(h *MessageHandler, store *archivalcitations.Store, k *KnowledgeHandler) *MessageHandler {
	h.citationStore = store
	h.citationKnowledge = k
	return h
}
func (h *MessageHandler) citationMessage(c *gin.Context) (*types.Message, bool) {
	if !archivalcitations.Enabled() || h.citationStore == nil || h.citationKnowledge == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "档案级引用尚未启用"})
		return nil, false
	}
	message, e := h.MessageService.GetMessage(c.Request.Context(), c.Param("session_id"), c.Param("id"))
	if e != nil || message == nil || message.Role != "assistant" {
		c.Error(apperrors.NewNotFoundError("引用消息不存在或不可访问"))
		return nil, false
	}
	return message, true
}
func (h *MessageHandler) canReadCitation(c *gin.Context, summary archivalcitations.Summary) bool {
	if summary.WikiPageID != "" {
		kb, _, tenant, permission, err := h.citationKnowledge.validateKnowledgeBaseAccessWithKBID(c, summary.KnowledgeBaseID)
		if err != nil || kb == nil || !permission.HasPermission(types.OrgRoleViewer) {
			return false
		}
		var count int64
		err = h.citationStore.DB.WithContext(c.Request.Context()).Table("wiki_pages").Where("id=? AND tenant_id=? AND knowledge_base_id=? AND deleted_at IS NULL", summary.WikiPageID, tenant, summary.KnowledgeBaseID).Count(&count).Error
		return err == nil && count == 1
	}
	k, _, e := h.citationKnowledge.resolveKnowledgeAndValidateKBAccess(c, summary.KnowledgeID, types.OrgRoleViewer)
	return e == nil && k != nil && k.KnowledgeBaseID == summary.KnowledgeBaseID
}
func (h *MessageHandler) GetMessageCitations(c *gin.Context) {
	if !archivalcitations.Enabled() {
		c.JSON(200, gin.H{"enabled": false, "success": true, "data": nil})
		return
	}
	message, ok := h.citationMessage(c)
	if !ok {
		return
	}
	manifest, e := h.citationStore.Manifest(c.Request.Context(), types.MustTenantIDFromContext(c.Request.Context()), message.SessionID, message.ID, message.Content)
	if errors.Is(e, gorm.ErrRecordNotFound) {
		c.JSON(200, gin.H{"success": true, "data": archivalcitations.Manifest{Version: 1, State: "legacy", Citations: []archivalcitations.Summary{}, Unbound: []archivalcitations.UnboundBlock{}}})
		return
	}
	if e != nil {
		c.Error(apperrors.NewInternalServerError("引用记录读取失败"))
		return
	}
	allowed := make([]archivalcitations.Summary, 0, len(manifest.Citations))
	unavailable := 0
	for _, summary := range manifest.Citations {
		if h.canReadCitation(c, summary) {
			summary.Origins = h.authorizedCitationOrigins(c, summary)
			allowed = append(allowed, summary)
		} else {
			unavailable++
		}
	}
	manifest.Citations = allowed
	if unavailable > 0 {
		manifest.State = "needs_review"
	}
	c.JSON(200, gin.H{"success": true, "data": manifest, "unavailable_sources": unavailable})
}
func (h *MessageHandler) GetMessageCitation(c *gin.Context) {
	message, ok := h.citationMessage(c)
	if !ok {
		return
	}
	row, summary, e := h.citationStore.Evidence(c.Request.Context(), types.MustTenantIDFromContext(c.Request.Context()), message.SessionID, message.ID, c.Param("citation_id"))
	if e != nil || !h.canReadCitation(c, summary) {
		c.Error(apperrors.NewNotFoundError("引用来源不存在或不可访问"))
		return
	}
	var current types.Chunk
	if summary.WikiPageID != "" {
		var page struct {
			Content string
			Version int
		}
		e = h.citationStore.DB.WithContext(c.Request.Context()).Table("wiki_pages").Where("id=? AND tenant_id=? AND knowledge_base_id=? AND deleted_at IS NULL", summary.WikiPageID, row.KnowledgeTenantID, summary.KnowledgeBaseID).Take(&page).Error
		current.Content = page.Content
		current.ContentRevision = page.Version
	} else {
		e = h.citationStore.DB.WithContext(c.Request.Context()).Where("tenant_id=? AND knowledge_id=? AND id=?", row.KnowledgeTenantID, row.KnowledgeID, row.ChunkID).First(&current).Error
	}
	if e != nil {
		c.Error(apperrors.NewNotFoundError("来源分块已删除或不可访问"))
		return
	}
	summary.Origins = h.authorizedCitationOrigins(c, summary)
	c.JSON(200, gin.H{"success": true, "data": archivalcitations.Detail{Summary: summary, Snapshot: row.Snapshot, CurrentContent: current.Content, CurrentRevision: current.ContentRevision, CurrentChanged: archivalcitations.Hash(current.Content) != summary.ContentSHA256, SnapshotAt: row.CreatedAt.Format(time.RFC3339Nano)}})
}

func (h *MessageHandler) authorizedCitationOrigins(c *gin.Context, summary archivalcitations.Summary) []archivalcitations.SourceOrigin {
	allowed := []archivalcitations.SourceOrigin{}
	for _, origin := range summary.Origins {
		knowledge, _, err := h.citationKnowledge.resolveKnowledgeAndValidateKBAccess(c, origin.KnowledgeID, types.OrgRoleViewer)
		if err == nil && knowledge != nil && knowledge.KnowledgeBaseID == summary.KnowledgeBaseID {
			allowed = append(allowed, origin)
		}
	}
	return allowed
}

// Current origin content is never substituted for the immutable quote.
func (h *MessageHandler) GetCitationOriginChunk(c *gin.Context) {
	message, ok := h.citationMessage(c)
	if !ok {
		return
	}
	_, summary, e := h.citationStore.Evidence(c.Request.Context(), types.MustTenantIDFromContext(c.Request.Context()), message.SessionID, message.ID, c.Param("citation_id"))
	if e != nil || !h.canReadCitation(c, summary) {
		c.Error(apperrors.NewNotFoundError("引用来源不存在或不可访问"))
		return
	}
	originID, chunkID := c.Param("origin_id"), c.Param("chunk_id")
	found := false
	for _, origin := range summary.Origins {
		if origin.KnowledgeID == originID {
			for _, id := range origin.ChunkIDs {
				if id == chunkID {
					found = true
				}
			}
		}
	}
	if !found {
		c.Error(apperrors.NewNotFoundError("该分块不在引用时记录的来源关系中"))
		return
	}
	knowledge, _, e := h.citationKnowledge.resolveKnowledgeAndValidateKBAccess(c, originID, types.OrgRoleViewer)
	if e != nil || knowledge == nil || knowledge.KnowledgeBaseID != summary.KnowledgeBaseID {
		c.Error(apperrors.NewNotFoundError("来源已删除或不可访问"))
		return
	}
	var chunk types.Chunk
	if e = h.citationStore.DB.WithContext(c.Request.Context()).Where("tenant_id=? AND knowledge_id=? AND id=?", knowledge.TenantID, originID, chunkID).First(&chunk).Error; e != nil {
		c.Error(apperrors.NewNotFoundError("来源分块已删除"))
		return
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"content": chunk.Content, "content_revision": chunk.ContentRevision, "chunk_id": chunk.ID, "knowledge_id": knowledge.ID, "state": "current_source_not_snapshot"}})
}

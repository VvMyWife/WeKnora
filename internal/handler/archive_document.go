package handler

import (
	"context"
	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

type archiveDocumentService interface {
	ArchiveDocumentContext(context.Context, string) (map[string]any, error)
	RetryArchiveFailures(context.Context, string, string) (map[string]any, error)
	ArchiveFailureCandidates(context.Context, string) ([]string, error)
}

func (h *KnowledgeHandler) GetArchiveDocumentContext(c *gin.Context) {
	k, ctx, err := h.resolveKnowledgeAndValidateKBAccess(c, c.Param("id"), types.OrgRoleViewer)
	if err != nil {
		c.Error(err)
		return
	}
	s, ok := h.kgService.(archiveDocumentService)
	if !ok {
		c.JSON(503, gin.H{"success": false})
		return
	}
	data, err := s.ArchiveDocumentContext(ctx, k.ID)
	if err != nil {
		c.Error(apperrors.NewInternalServerError(err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func (h *KnowledgeHandler) RetryArchiveDocumentFailures(c *gin.Context) {
	k, ctx, err := h.resolveKnowledgeAndValidateKBAccess(c, c.Param("id"), types.OrgRoleEditor)
	if err != nil {
		c.Error(err)
		return
	}
	var body struct {
		Stage string `json:"stage"`
	}
	if c.Request.ContentLength != 0 && c.ShouldBindJSON(&body) != nil {
		c.Error(apperrors.NewBadRequestError("invalid retry request"))
		return
	}
	s, ok := h.kgService.(archiveDocumentService)
	if !ok {
		c.JSON(503, gin.H{"success": false})
		return
	}
	data, err := s.RetryArchiveFailures(ctx, k.ID, body.Stage)
	if err != nil {
		c.Error(apperrors.NewBadRequestError(err.Error()))
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"success": true, "data": data})
}

func (h *KnowledgeHandler) RetryArchiveBatchFailures(c *gin.Context) {
	_, kb, tenant, _, err := h.validateKnowledgeBaseAccess(c)
	if err != nil {
		c.Error(err)
		return
	}
	if err = h.requireKBOwnershipOrAdmin(c, kb); err != nil {
		c.Error(err)
		return
	}
	var body struct {
		Stage string `json:"stage"`
	}
	if c.Request.ContentLength != 0 && c.ShouldBindJSON(&body) != nil {
		c.Error(apperrors.NewBadRequestError("invalid retry request"))
		return
	}
	s, ok := h.kgService.(archiveDocumentService)
	if !ok {
		c.JSON(503, gin.H{"success": false})
		return
	}
	ctx := context.WithValue(c.Request.Context(), types.TenantIDContextKey, tenant)
	ids, err := s.ArchiveFailureCandidates(ctx, kb)
	if err != nil {
		c.Error(apperrors.NewInternalServerError(err.Error()))
		return
	}
	results := []map[string]any{}
	for _, id := range ids {
		result, e := s.RetryArchiveFailures(ctx, id, body.Stage)
		if e != nil {
			result = map[string]any{"knowledge_id": id, "accepted": 0, "error": e.Error()}
		}
		results = append(results, result)
	}
	c.JSON(http.StatusAccepted, gin.H{"success": true, "data": results, "documents": len(results)})
}

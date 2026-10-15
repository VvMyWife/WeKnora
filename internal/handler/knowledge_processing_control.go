package handler

import (
	"net/http"

	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/processingcontrol"
	"github.com/gin-gonic/gin"
)

func (h *KnowledgeHandler) GetKnowledgeProcessingControl(c *gin.Context) {
	_, kbID, tenantID, _, err := h.validateKnowledgeBaseAccess(c)
	if err != nil {
		c.Error(err)
		return
	}
	s, err := processingcontrol.Current().State(c.Request.Context(), tenantID, kbID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "无法读取后台处理暂停状态，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": s})
}

func (h *KnowledgeHandler) SetKnowledgeProcessingControl(c *gin.Context) {
	_, kbID, tenantID, _, err := h.validateKnowledgeBaseAccess(c)
	if err != nil {
		c.Error(err)
		return
	}
	if err := h.requireKBOwnershipOrAdmin(c, kbID); err != nil {
		c.Error(err)
		return
	}
	var body struct {
		Paused *bool `json:"paused"`
	}
	if c.ShouldBindJSON(&body) != nil || body.Paused == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "paused 必须为布尔值"})
		return
	}
	control := processingcontrol.Current()
	if !control.Supported() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "当前任务后端不支持暂停"})
		return
	}
	s, err := control.SetPaused(c.Request.Context(), tenantID, kbID, *body.Paused)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "暂停状态保存失败，请重试"})
		return
	}
	logger.Infof(c.Request.Context(), "[KnowledgeProcessingControl] kb_id=%s paused=%t active_tasks=%d state=%s", kbID, *body.Paused, s.ActiveTasks, s.State)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": s})
}

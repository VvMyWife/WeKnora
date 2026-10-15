package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/WeKnora/internal/archivemapping"
	apperrors "github.com/Tencent/WeKnora/internal/errors"
	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AttachArchiveStore(h *KnowledgeHandler, s *archivemapping.Store) *KnowledgeHandler {
	h.archiveStore = s
	return h
}
func (h *KnowledgeHandler) enrichArchiveSource(ctx context.Context, k *types.Knowledge) {
	if k == nil || h.archiveStore == nil {
		return
	}
	v, e := h.archiveStore.Active(ctx, k.TenantID, k.ID)
	if e != nil {
		logger.Warnf(ctx, "[ArchiveMapping] metadata knowledge=%s error=%v", k.ID, e)
		return
	}
	k.ArchiveSource = v
}

type archiveFiles interface {
	SaveArchiveFile(context.Context, string, *multipart.FileHeader, string) (string, error)
	OpenArchiveFile(context.Context, string, string) (io.ReadCloser, error)
	DeleteArchiveFile(context.Context, string, string) error
}

func (h *KnowledgeHandler) archiveScope(c *gin.Context, write bool) (context.Context, uint64, string, bool) {
	_, kb, tenant, _, err := h.validateKnowledgeBaseAccess(c)
	if err == nil && write {
		err = h.requireKBOwnershipOrAdmin(c, kb)
	}
	if err != nil {
		c.Error(err)
		return nil, 0, "", false
	}
	if h.archiveStore == nil {
		c.JSON(503, gin.H{"success": false, "message": "档案映射服务未启用"})
		return nil, 0, "", false
	}
	return context.WithValue(c.Request.Context(), types.TenantIDContextKey, tenant), tenant, kb, true
}
func archiveError(c *gin.Context, err error) {
	logger.Warnf(c.Request.Context(), "[ArchiveMapping] operation failed: %v", err)
	c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
}
func (h *KnowledgeHandler) CreateArchiveBatch(c *gin.Context) {
	ctx, tenant, kb, ok := h.archiveScope(c, true)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 8<<20)
	var req struct {
		ParsedRoot string                        `json:"parsed_root"`
		Files      []archivemapping.ManifestFile `json:"files"`
	}
	if e := c.ShouldBindJSON(&req); e != nil {
		archiveError(c, fmt.Errorf("无效的原件清单: %w", e))
		return
	}
	b, items, e := h.archiveStore.Plan(ctx, tenant, kb, req.ParsedRoot, c.GetString(types.UserIDContextKey.String()), req.Files)
	if e != nil {
		archiveError(c, e)
		return
	}
	h.archiveLog(ctx, tenant, kb, b.ID, "batch_created", "", nil)
	c.JSON(201, gin.H{"success": true, "data": gin.H{"batch": b, "items": items}})
}
func (h *KnowledgeHandler) ListArchiveBatches(c *gin.Context) {
	ctx, tenant, kb, ok := h.archiveScope(c, false)
	if !ok {
		return
	}
	batches, e := h.archiveStore.Batches(ctx, tenant, kb)
	if e != nil {
		archiveError(c, e)
		return
	}
	targets, e := h.archiveStore.Targets(ctx, tenant, kb)
	if e != nil {
		archiveError(c, e)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"batches": batches, "targets": targets}})
}
func (h *KnowledgeHandler) GetArchiveBatch(c *gin.Context) {
	ctx, tenant, kb, ok := h.archiveScope(c, false)
	if !ok {
		return
	}
	b, items, e := h.archiveStore.Batch(ctx, tenant, kb, c.Param("batch_id"))
	if e != nil {
		archiveError(c, e)
		return
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"batch": b, "items": items}})
}
func (h *KnowledgeHandler) SkipArchiveItem(c *gin.Context) {
	ctx, tenant, kb, ok := h.archiveScope(c, true)
	if !ok {
		return
	}
	e := h.archiveStore.ItemState(ctx, tenant, kb, c.Param("batch_id"), c.Param("item_id"), "skipped", "用户跳过")
	if e != nil {
		archiveError(c, e)
		return
	}
	h.archiveLog(ctx, tenant, kb, c.Param("batch_id"), "item_skipped", c.Param("item_id"), nil)
	c.JSON(200, gin.H{"success": true})
}
func (h *KnowledgeHandler) UploadArchiveItem(c *gin.Context) {
	ctx, tenant, kb, ok := h.archiveScope(c, true)
	if !ok {
		return
	}
	batch, item := c.Param("batch_id"), c.Param("item_id")
	// Validate batch and item before reading a potentially large request body.
	_, items, e := h.archiveStore.Batch(ctx, tenant, kb, batch)
	if e != nil {
		archiveError(c, e)
		return
	}
	var planned *archivemapping.Item
	for i := range items {
		if items[i].ID == item {
			planned = &items[i]
			break
		}
	}
	if planned == nil {
		archiveError(c, fmt.Errorf("批次中不存在该文件"))
		return
	}
	// No inherited parsing upload size cap: original storage is streamed to disk
	// by multipart with a bounded memory threshold, then through FileService.
	if e = c.Request.ParseMultipartForm(8 << 20); e != nil {
		h.archiveUploadFailure(c, ctx, tenant, kb, batch, item, e)
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	f, e := c.FormFile("file")
	if e != nil {
		h.archiveUploadFailure(c, ctx, tenant, kb, batch, item, e)
		return
	}
	if f.Filename != path.Base(planned.RelativePath) {
		h.archiveUploadFailure(c, ctx, tenant, kb, batch, item, fmt.Errorf("文件名与批次清单不一致"))
		return
	}
	stream, e := f.Open()
	if e != nil {
		h.archiveUploadFailure(c, ctx, tenant, kb, batch, item, e)
		return
	}
	digest := sha256.New()
	n, e := io.Copy(digest, stream)
	stream.Close()
	if e != nil {
		h.archiveUploadFailure(c, ctx, tenant, kb, batch, item, e)
		return
	}
	files, ok := h.kgService.(archiveFiles)
	if !ok {
		archiveError(c, fmt.Errorf("原件存储不可用"))
		return
	}
	saved := ""
	v, e := h.archiveStore.Bind(ctx, tenant, kb, batch, item, c.PostForm("knowledge_id"), c.GetString(types.UserIDContextKey.String()), hex.EncodeToString(digest.Sum(nil)), n, c.PostForm("replace") == "true", func(version string) (string, error) {
		stored, err := files.SaveArchiveFile(ctx, kb, f, version)
		saved = stored
		return stored, err
	})
	if e != nil {
		if saved != "" {
			if cleanupErr := files.DeleteArchiveFile(context.WithoutCancel(ctx), kb, saved); cleanupErr != nil {
				logger.Errorf(ctx, "[ArchiveMapping] rollback cleanup batch=%s item=%s error=%v", batch, item, cleanupErr)
			}
		}
		h.archiveUploadFailure(c, ctx, tenant, kb, batch, item, e)
		return
	}
	h.archiveLog(ctx, tenant, kb, batch, "item_bound", item, v)
	c.JSON(200, gin.H{"success": true, "data": v})
}
func (h *KnowledgeHandler) archiveUploadFailure(c *gin.Context, ctx context.Context, tenant uint64, kb, batch, item string, e error) {
	persist := context.WithoutCancel(ctx)
	if err := h.archiveStore.ItemState(persist, tenant, kb, batch, item, "failed", e.Error()); err != nil {
		logger.Errorf(ctx, "[ArchiveMapping] failed to persist upload error batch=%s item=%s error=%v", batch, item, err)
	}
	h.archiveLog(persist, tenant, kb, batch, "item_failed", item, map[string]string{"reason": e.Error()})
	archiveError(c, e)
}
func (h *KnowledgeHandler) GetArchiveSource(c *gin.Context) {
	k, ctx, e := h.resolveKnowledgeAndValidateKBAccess(c, c.Param("id"), types.OrgRoleViewer)
	if e != nil {
		c.Error(e)
		return
	}
	if h.archiveStore == nil {
		c.JSON(200, gin.H{"success": true, "data": gin.H{"active": nil, "versions": []any{}}})
		return
	}
	active, e := h.archiveStore.Active(ctx, k.TenantID, k.ID)
	if e != nil {
		archiveError(c, e)
		return
	}
	versions, e := h.archiveStore.Versions(ctx, k.TenantID, k.ID)
	if e != nil {
		archiveError(c, e)
		return
	}
	var total int64
	for _, v := range versions {
		total += v.FileSize
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"active": active, "versions": versions, "storage_bytes": total}})
}
func (h *KnowledgeHandler) SetArchiveSource(c *gin.Context) {
	k, ctx, e := h.resolveKnowledgeAndValidateKBAccess(c, c.Param("id"), types.OrgRoleEditor)
	if e != nil {
		c.Error(e)
		return
	}
	if e = h.requireKBOwnershipOrAdmin(c, k.KnowledgeBaseID); e != nil {
		c.Error(e)
		return
	}
	if h.archiveStore == nil {
		archiveError(c, fmt.Errorf("档案映射服务未启用"))
		return
	}
	var req struct {
		VersionID *string `json:"version_id"`
	}
	if e = c.ShouldBindJSON(&req); e != nil || req.VersionID == nil {
		archiveError(c, fmt.Errorf("version_id 必填，空字符串表示解除关联但保留历史版本"))
		return
	}
	if e = h.archiveStore.Activate(ctx, k.TenantID, k.KnowledgeBaseID, k.ID, *req.VersionID); e != nil {
		archiveError(c, e)
		return
	}
	logger.Infof(ctx, "[ArchiveMapping] source_changed kb_id=%s knowledge_id=%s version_id=%s", k.KnowledgeBaseID, k.ID, *req.VersionID)
	c.JSON(200, gin.H{"success": true})
}

var archiveLogMu sync.Mutex

func (h *KnowledgeHandler) archiveLog(ctx context.Context, tenant uint64, kb, batch, event, item string, detail any) {
	logger.Infof(ctx, "[ArchiveMapping] event=%s kb_id=%s batch_id=%s item_id=%s", event, kb, batch, item)
	// IDs are DB-validated as well; UUID validation prevents filesystem traversal.
	if _, e := uuid.Parse(kb); e != nil {
		return
	}
	if _, e := uuid.Parse(batch); e != nil {
		return
	}
	archiveLogMu.Lock()
	defer archiveLogMu.Unlock()
	dir := filepath.Join("data", "processing-logs", kb, "archive-mapping", batch)
	err := os.MkdirAll(dir, 0750)
	if err == nil {
		var f *os.File
		f, err = os.OpenFile(filepath.Join(dir, "events.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
		if err == nil {
			err = json.NewEncoder(f).Encode(map[string]any{"time": time.Now().UTC(), "event": event, "item_id": item, "detail": detail})
			f.Close()
		}
	}
	if err == nil {
		b, items, e := h.archiveStore.Batch(ctx, tenant, kb, batch)
		err = e
		if e == nil {
			var text strings.Builder
			fmt.Fprintf(&text, "档案映射批次 %s\n知识库 %s\nMD 根目录 %s\n文件数 %d\n", b.ID, kb, b.ParsedRoot, len(items))
			for _, it := range items {
				fmt.Fprintf(&text, "[%s] %s -> %s %s\n", it.Status, it.RelativePath, it.KnowledgeID, it.Error)
			}
			tmp := filepath.Join(dir, "status.txt.tmp")
			err = os.WriteFile(tmp, []byte(text.String()), 0640)
			if err == nil {
				err = os.Rename(tmp, filepath.Join(dir, "status.txt"))
			}
		}
	}
	if err != nil {
		logger.Errorf(ctx, "[ArchiveMapping] log_export_failed kb=%s batch=%s error=%v", kb, batch, err)
	}
}

// Browser-only resolution. Internal OCR/reparse reads always keep using MD.
func (h *KnowledgeHandler) selectedKnowledgeFile(c *gin.Context, ctx context.Context, id string) (io.ReadCloser, string, error) {
	source := c.DefaultQuery("source", "auto")
	if source != "auto" && source != "archive" && source != "parsed" {
		return nil, "", apperrors.NewBadRequestError("source 必须是 auto、archive 或 parsed")
	}
	if source != "parsed" && h.archiveStore != nil {
		k, e := h.kgService.GetKnowledgeByID(ctx, id)
		if e != nil {
			return nil, "", e
		}
		v, e := h.archiveStore.Active(ctx, k.TenantID, id)
		if e != nil {
			return nil, "", e
		}
		if v != nil {
			files, ok := h.kgService.(archiveFiles)
			if !ok {
				return nil, "", fmt.Errorf("原件存储不可用")
			}
			f, e := files.OpenArchiveFile(ctx, k.KnowledgeBaseID, v.StoragePath)
			return f, v.FileName, e
		}
	}
	if source == "archive" {
		return nil, "", apperrors.NewNotFoundError("文档尚未关联原件")
	}
	return h.kgService.GetKnowledgeFile(ctx, id)
}

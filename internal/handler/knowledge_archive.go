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
	SaveArchiveFile(context.Context, string, *multipart.FileHeader, string, string) (string, error)
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
	// Bound the body by the declared original size plus multipart overhead,
	// rather than applying the unrelated parser's 50 MB limit.
	if planned.FileSize > (1 << 60) {
		archiveError(c, fmt.Errorf("清单文件大小无效"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, planned.FileSize+(2<<20))
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
	target := c.PostForm("knowledge_id")
	if target == "" {
		target = planned.KnowledgeID
	}
	v, e := h.bindArchiveUpload(ctx, tenant, kb, batch, planned, target, c.GetString(types.UserIDContextKey.String()), hex.EncodeToString(digest.Sum(nil)), n, c.PostForm("replace") == "true", f)
	if e != nil {
		h.archiveUploadFailure(c, ctx, tenant, kb, batch, item, e)
		return
	}
	h.archiveLog(ctx, tenant, kb, batch, "item_bound", item, v)
	c.JSON(200, gin.H{"success": true, "data": v})
}
func (h *KnowledgeHandler) bindArchiveUpload(ctx context.Context, tenant uint64, kb, batch string, item *archivemapping.Item, target, user, hash string, size int64, replace bool, f *multipart.FileHeader) (*archivemapping.Version, error) {
	files, ok := h.kgService.(archiveFiles)
	if !ok {
		return nil, fmt.Errorf("原件存储不可用")
	}
	k, e := h.kgService.GetKnowledgeByID(ctx, target)
	if e != nil {
		return nil, e
	}
	if k == nil || k.ID != target || k.TenantID != tenant || k.KnowledgeBaseID != kb || (k.FileType != "md" && k.FileType != "markdown") {
		return nil, fmt.Errorf("目标必须是当前知识库的 MD 文档")
	}
	if size != item.FileSize || !archivemapping.AllowedOriginal(item.RelativePath) {
		return nil, fmt.Errorf("原件格式或大小与清单不一致")
	}
	current, e := h.archiveStore.Active(ctx, tenant, target)
	if e != nil {
		return nil, e
	}
	saved := ""
	// Catalog registration uses the database too. Store the object BEFORE the
	// metadata transaction: doing I/O inside Bind deadlocks SQLite's one-connection
	// pool and holds PostgreSQL row/quota locks for the entire large-file copy.
	if item.Status != "bound" && (current == nil || current.SHA256 != hash) {
		if current != nil && !replace {
			return nil, fmt.Errorf("目标已有不同原件，请明确选择替换")
		}
		saved, e = files.SaveArchiveFile(ctx, kb, f, target, uuid.NewString())
		if e != nil {
			return nil, e
		}
		h.archiveLog(ctx, tenant, kb, batch, "item_staged", item.ID, map[string]string{"resource": saved, "sha256": hash})
	}
	v, e := h.archiveStore.Bind(ctx, tenant, kb, batch, item.ID, target, user, hash, size, replace, func(string) (string, error) {
		if saved == "" {
			return "", fmt.Errorf("关联状态已变化，请重试")
		}
		return saved, nil
	})
	// A concurrent identical upload may have won; remove our unused staging copy
	// on both a rejected publication and an idempotent winner.
	if saved != "" && (e != nil || v == nil || v.StoragePath != saved) {
		if cleanupErr := files.DeleteArchiveFile(context.WithoutCancel(ctx), kb, saved); cleanupErr != nil {
			logger.Errorf(ctx, "[ArchiveMapping] staged cleanup batch=%s item=%s resource=%s error=%v", batch, item.ID, saved, cleanupErr)
		}
	}
	return v, e
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
	h.archiveLog(ctx, k.TenantID, k.KnowledgeBaseID, "", "source_changed", k.ID, map[string]string{"version_id": *req.VersionID})
	c.JSON(200, gin.H{"success": true})
}

var archiveLogMu sync.Mutex

func (h *KnowledgeHandler) archiveLog(ctx context.Context, tenant uint64, kb, batch, event, item string, detail any) {
	logger.Infof(ctx, "[ArchiveMapping] event=%s kb_id=%s batch_id=%s item_id=%s", event, kb, batch, item)
	// IDs are DB-validated as well; UUID validation prevents filesystem traversal.
	if _, e := uuid.Parse(kb); e != nil {
		return
	}
	root := os.Getenv("ARCHIVE_MAPPING_LOG_DIR")
	if root == "" {
		root = filepath.Join("data", "processing-logs")
	}
	if batch == "" {
		archiveLogMu.Lock()
		defer archiveLogMu.Unlock()
		dir := filepath.Join(root, kb, "archive-mapping")
		err := os.MkdirAll(dir, 0750)
		if err == nil {
			var f *os.File
			f, err = os.OpenFile(filepath.Join(dir, "source-events.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
			if err == nil {
				err = json.NewEncoder(f).Encode(map[string]any{"time": time.Now().UTC(), "event": event, "knowledge_id": item, "detail": detail})
				f.Close()
			}
		}
		if err != nil {
			logger.Errorf(ctx, "[ArchiveMapping] source_log_export_failed kb=%s error=%v", kb, err)
		}
		return
	}
	if _, e := uuid.Parse(batch); e != nil {
		return
	}
	archiveLogMu.Lock()
	defer archiveLogMu.Unlock()
	dir := filepath.Join(root, kb, "archive-mapping", batch)
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
	if versionID := c.Query("archive_version_id"); versionID != "" {
		if source == "parsed" {
			return nil, "", apperrors.NewBadRequestError("解析文件不能指定原件版本")
		}
		if _, e := uuid.Parse(versionID); e != nil {
			return nil, "", apperrors.NewBadRequestError("原件版本 ID 无效")
		}
		if h.archiveStore == nil {
			return nil, "", apperrors.NewNotFoundError("原件存储不可用")
		}
		k, e := h.kgService.GetKnowledgeByID(ctx, id)
		if e != nil {
			return nil, "", e
		}
		v, e := h.archiveStore.Version(ctx, k.TenantID, id, versionID)
		if e != nil {
			return nil, "", apperrors.NewNotFoundError("指定原件版本不存在")
		}
		files, ok := h.kgService.(archiveFiles)
		if !ok {
			return nil, "", fmt.Errorf("原件存储不可用")
		}
		f, e := files.OpenArchiveFile(ctx, k.KnowledgeBaseID, v.StoragePath)
		return f, v.FileName, e
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

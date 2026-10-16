package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Tencent/WeKnora/internal/externalchunks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AttachExternalChunkStore(h *KnowledgeHandler, s *externalchunks.Store) *KnowledgeHandler {
	h.externalStore = s
	return h
}

type externalImporter interface {
	StartExternalImport(context.Context, string, string, string) error
}

var externalPreflightSlots = make(chan struct{}, 2)
var externalLogMu sync.Mutex

func (h *KnowledgeHandler) externalScope(c *gin.Context, write bool) (context.Context, uint64, string, bool) {
	ctx, tenant, kb, ok := h.archiveScope(c, write)
	if !ok {
		return nil, 0, "", false
	}
	if h.externalStore == nil {
		c.JSON(503, gin.H{"message": "外部分块导入未启用"})
		return nil, 0, "", false
	}
	return ctx, tenant, kb, true
}
func externalLog(kb, batch, event string, detail any) {
	if _, e := uuid.Parse(kb); e != nil {
		return
	}
	if _, e := uuid.Parse(batch); e != nil {
		return
	}
	root := os.Getenv("ARCHIVE_MAPPING_LOG_DIR")
	if root == "" {
		root = "data/processing-logs"
	}
	dir := filepath.Join(root, kb, "external-chunks", batch)
	externalLogMu.Lock()
	defer externalLogMu.Unlock()
	if os.MkdirAll(dir, 0750) != nil {
		return
	}
	f, e := os.OpenFile(filepath.Join(dir, "events.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if e != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(map[string]any{"time": time.Now().Format(time.RFC3339), "event": event, "detail": detail})
}
func (h *KnowledgeHandler) PreviewExternalChunks(c *gin.Context) {
	ctx, tenant, kb, ok := h.externalScope(c, true)
	if !ok {
		return
	}
	select {
	case externalPreflightSlots <- struct{}{}:
		defer func() { <-externalPreflightSlots }()
	default:
		c.JSON(429, gin.H{"message": "当前有两个导入包正在预检，请稍后重试"})
		return
	}
	upload := uuid.NewString()
	root := filepath.Join("/data/files/external-imports", fmt.Sprint(tenant), kb)
	if e := os.MkdirAll(root, 0750); e != nil {
		archiveError(c, e)
		return
	}
	filename := filepath.Join(root, upload+".zip")
	keep := false
	defer func() {
		if !keep {
			_ = os.Remove(filename)
		}
	}()
	fail := func(e error) {
		externalLog(kb, upload, "upload_rejected", map[string]string{"reason": e.Error()})
		archiveError(c, e)
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4<<30)
	reader, e := c.Request.MultipartReader()
	if e != nil {
		fail(e)
		return
	}
	f, e := os.OpenFile(filename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if e != nil {
		fail(e)
		return
	}
	got := false
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			e = err
			break
		}
		if got || part.FormName() != "file" || part.FileName() == "" {
			e = fmt.Errorf("只接受一个 ZIP file 字段")
			part.Close()
			break
		}
		got = true
		_, e = io.Copy(f, part)
		part.Close()
		if e != nil {
			break
		}
	}
	if err := f.Close(); e == nil {
		e = err
	}
	if e != nil {
		fail(e)
		return
	}
	if !got {
		fail(fmt.Errorf("缺少 ZIP 文件"))
		return
	}
	p, e := externalchunks.Read(filename)
	if e != nil {
		fail(e)
		return
	}
	batch, rows, e := h.externalStore.Plan(ctx, tenant, kb, filename, p)
	if e != nil {
		fail(e)
		return
	}
	keep = batch.PackagePath == filename
	externalLog(kb, batch.ID, "preflight_passed", map[string]any{"documents": len(rows), "chunks": len(p.Chunks), "manifest_hash": p.ManifestHash})
	c.JSON(201, gin.H{"success": true, "data": gin.H{"batch": batch, "documents": rows, "counts": p.Manifest.Counts}})
}
func (h *KnowledgeHandler) ListExternalChunks(c *gin.Context) {
	ctx, t, kb, ok := h.externalScope(c, false)
	if !ok {
		return
	}
	rows, e := h.externalStore.List(ctx, t, kb)
	if e != nil {
		archiveError(c, e)
		return
	}
	c.JSON(200, gin.H{"data": rows, "success": true})
}
func (h *KnowledgeHandler) GetExternalChunks(c *gin.Context) {
	ctx, t, kb, ok := h.externalScope(c, false)
	if !ok {
		return
	}
	b, rows, e := h.externalStore.Batch(ctx, t, kb, c.Param("batch_id"))
	if e != nil {
		archiveError(c, e)
		return
	}
	for i := range rows {
		if rows[i].Status != "queued" && rows[i].Status != "postprocessing" {
			continue
		}
		if k, err := h.kgService.GetKnowledgeByID(ctx, rows[i].KnowledgeID); err == nil && k != nil {
			rows[i].Status = k.ParseStatus
			if k.ErrorMessage != "" {
				rows[i].Error = k.ErrorMessage
			}
			if k.SummaryStatus == "failed" {
				rows[i].Status = "failed"
				rows[i].Stage = "summary"
				rows[i].Error = "摘要生成失败，请查看全库进度面板或文档 Trace"
			}
		}
	}
	c.JSON(200, gin.H{"data": gin.H{"batch": b, "documents": rows}, "success": true})
}
func (h *KnowledgeHandler) StartExternalChunks(c *gin.Context) {
	ctx, _, kb, ok := h.externalScope(c, true)
	if !ok {
		return
	}
	s, ok := h.kgService.(externalImporter)
	if !ok {
		archiveError(c, fmt.Errorf("外部导入服务不可用"))
		return
	}
	e := s.StartExternalImport(ctx, kb, c.Param("batch_id"), c.Param("record_id"))
	if e != nil {
		externalLog(kb, c.Param("batch_id"), "document_failed", map[string]string{"record_id": c.Param("record_id"), "reason": e.Error()})
		archiveError(c, e)
		return
	}
	externalLog(kb, c.Param("batch_id"), "document_submitted", map[string]string{"record_id": c.Param("record_id")})
	c.JSON(202, gin.H{"success": true})
}
func (h *KnowledgeHandler) ExternalChunkIDMap(c *gin.Context) {
	ctx, t, kb, ok := h.externalScope(c, false)
	if !ok {
		return
	}
	_, rows, e := h.externalStore.Batch(ctx, t, kb, c.Param("batch_id"))
	if e != nil {
		archiveError(c, e)
		return
	}
	out := []map[string]any{}
	for _, r := range rows {
		var p externalchunks.DocumentPayload
		if e = json.Unmarshal([]byte(r.Payload), &p); e != nil {
			archiveError(c, e)
			return
		}
		chunks := []map[string]string{}
		for _, ch := range p.Chunks {
			chunks = append(chunks, map[string]string{"external_chunk_id": ch.ID, "chunk_id": externalchunks.MappedID(r.ID, ch.ID)})
		}
		out = append(out, map[string]any{"external_document_id": r.ExternalDocumentID, "knowledge_id": r.KnowledgeID, "relative_path": r.RelativePath, "chunks": chunks})
	}
	c.Header("Content-Disposition", "attachment; filename=external-chunks-id-map.json")
	c.JSON(200, gin.H{"format": externalchunks.Format, "documents": out})
}

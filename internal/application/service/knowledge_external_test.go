package service

import (
	"context"
	"github.com/Tencent/WeKnora/internal/externalchunks"
	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestExternalPersistRetriesResurrectStableChunkIDs(t *testing.T) {
	db, e := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&types.Chunk{}); e != nil {
		t.Fatal(e)
	}
	s := &knowledgeService{externalStore: &externalchunks.Store{DB: db}}
	chunk := &types.Chunk{ID: "stable", TenantID: 1, KnowledgeID: "knowledge", KnowledgeBaseID: "kb", Content: "原文", ChunkType: types.ChunkTypeText, IsEnabled: true}
	if e = s.persistExternalChunks(context.Background(), []*types.Chunk{chunk}); e != nil {
		t.Fatal(e)
	}
	seq := chunk.SeqID
	if e = db.Where("id = ?", chunk.ID).Delete(&types.Chunk{}).Error; e != nil {
		t.Fatal(e)
	}
	chunk = &types.Chunk{ID: "stable", TenantID: 1, KnowledgeID: "knowledge", KnowledgeBaseID: "kb", Content: "原文", ChunkType: types.ChunkTypeText, IsEnabled: true}
	if e = s.persistExternalChunks(context.Background(), []*types.Chunk{chunk}); e != nil {
		t.Fatal(e)
	}
	var actual types.Chunk
	if e = db.First(&actual, "id = ?", "stable").Error; e != nil || actual.SeqID != seq || actual.SourceContent != "原文" {
		t.Fatalf("not restored: %v", e)
	}
	var count int64
	db.Unscoped().Model(&types.Chunk{}).Count(&count)
	if count != 1 {
		t.Fatal("duplicate IDs")
	}
	chunk.TenantID = 2
	if e = s.persistExternalChunks(context.Background(), []*types.Chunk{chunk}); e == nil {
		t.Fatal("cross-tenant overwrite")
	}
}

package service

import (
	"context"
	"github.com/Tencent/WeKnora/internal/application/repository"
	"github.com/Tencent/WeKnora/internal/externalchunks"
	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"testing"
)

// Dedicated acceptance DB only. All rows and the sequence live in temporary
// objects on one transaction and disappear on rollback, including on failure.
func TestExternalPersistPostgresSequence(t *testing.T) {
	dsn := os.Getenv("LJC_EXTERNAL_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("set dedicated acceptance PostgreSQL DSN")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	var database string
	if err = db.Raw("SELECT current_database()").Scan(&database).Error; err != nil {
		t.Fatal(err)
	}
	if database != "ljc_acceptance" {
		t.Fatal("requires dedicated ljc_acceptance database")
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	for _, stmt := range []string{
		"CREATE TEMP SEQUENCE ljc_chunk_test_seq START 100000000",
		"CREATE TEMP TABLE chunks (LIKE public.chunks INCLUDING ALL) ON COMMIT DROP",
		"ALTER TABLE pg_temp.chunks ALTER COLUMN seq_id SET DEFAULT nextval('pg_temp.ljc_chunk_test_seq')",
	} {
		if err = tx.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := &knowledgeService{externalStore: &externalchunks.Store{DB: tx}}
	repo := repository.NewChunkRepository(tx)
	newChunk := func(id string) *types.Chunk {
		return &types.Chunk{ID: id, TenantID: 1, KnowledgeID: "acceptance-sequence", KnowledgeBaseID: "acceptance-kb", Content: "sequence regression", ChunkType: types.ChunkTypeText, IsEnabled: true}
	}
	first, external, summary := newChunk("normal"), newChunk("external"), newChunk("summary")
	if err = repo.CreateChunks(context.Background(), []*types.Chunk{first}); err != nil {
		t.Fatal(err)
	}
	if err = s.persistExternalChunks(context.Background(), []*types.Chunk{external}); err != nil {
		t.Fatal(err)
	}
	if err = repo.CreateChunks(context.Background(), []*types.Chunk{summary}); err != nil {
		t.Fatal(err)
	}
	if first.SeqID == external.SeqID || external.SeqID == summary.SeqID || first.SeqID == summary.SeqID {
		t.Fatal("duplicate sequence")
	}
	oldSeq := external.SeqID
	if err = tx.Delete(external).Error; err != nil {
		t.Fatal(err)
	}
	external = newChunk("external")
	if err = s.persistExternalChunks(context.Background(), []*types.Chunk{external}); err != nil {
		t.Fatal(err)
	}
	if external.SeqID != oldSeq {
		t.Fatal("retry changed imported chunk sequence")
	}
}

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// This real pg_search test uses an isolated, rolled-back schema in a disposable
// acceptance database. It checks filter semantics and LIMIT placement together.
func TestKeywordGeneratedQuestionBM25(t *testing.T) {
	dsn := os.Getenv("WEKNORA_RETRIEVER_BM25_TEST_DSN")
	if dsn == "" {
		t.Skip("set disposable pg_search acceptance DSN")
	}
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()
	var name string
	require.NoError(t, db.Raw("SELECT current_database()").Scan(&name).Error)
	require.Contains(t, name, "ljc_sync_", "refuse a non-acceptance database")
	tx := db.Begin()
	require.NoError(t, tx.Error)
	defer tx.Rollback()
	schema := fmt.Sprintf("ljc_bm25_%d", time.Now().UnixNano())
	require.NoError(t, tx.Exec("CREATE SCHEMA "+schema).Error)
	require.NoError(t, tx.Exec("SET LOCAL search_path TO "+schema+", public").Error)
	require.NoError(t, tx.Exec(`CREATE TABLE embeddings (
 id integer PRIMARY KEY, content text, source_id text, source_type integer,
 chunk_id text, knowledge_id text, knowledge_base_id text, tag_id text, is_enabled boolean)`).Error)
	require.NoError(t, tx.Exec(`INSERT INTO embeddings VALUES
 (1,'synchronization','c1',0,'c1','d1','a','tag',true),
 (2,'synchronization synchronization','c2-question',0,'c2','d1','a','tag',true),
 (3,'synchronization','c3-question',0,'c3','d2','b','tag',true),
 (4,'synchronization','c4',0,'c4','d1','a','tag',false),
 (5,'synchronization','c5',0,'c5','d1','a','tag',NULL)`).Error)
	require.NoError(t, tx.Exec(`CREATE INDEX keyword_fixture_idx ON embeddings
 USING bm25 (id, content, knowledge_base_id, knowledge_id, chunk_id)
 WITH (key_field='id')`).Error)
	repo := &pgRepository{db: tx}
	params := types.RetrieveParams{Query: "synchronization", TopK: 10,
		KnowledgeBaseIDs: []string{"a", "b"}, ExcludeGeneratedQuestionKBIDs: []string{"a"}}
	results, err := repo.KeywordsRetrieve(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, results, 1)
	ids := map[string]bool{}
	for _, row := range results[0].Results {
		ids[row.SourceID] = true
		require.Positive(t, row.Score)
	}
	require.Equal(t, map[string]bool{"c1": true, "c3-question": true, "c5": true}, ids)
	params.KnowledgeBaseIDs = []string{"a"}
	params.KnowledgeIDs = []string{"d1"}
	params.TagIDs = []string{"tag"}
	params.TopK = 1
	results, err = repo.KeywordsRetrieve(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, results[0].Results, 1)
	require.Contains(t, []string{"c1", "c5"}, results[0].Results[0].SourceID)
	params.ExcludeGeneratedQuestionKBIDs = nil
	params.TopK = 10
	results, err = repo.KeywordsRetrieve(context.Background(), params)
	require.NoError(t, err)
	require.Len(t, results[0].Results, 3)
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Tencent/WeKnora/internal/application/repository"
	"github.com/Tencent/WeKnora/internal/processingcontrol"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestEnrichmentRecoveryRealPostgresRedis(t *testing.T) {
	dsn := os.Getenv("WEKNORA_ENRICHMENT_TEST_DSN")
	addr := os.Getenv("WEKNORA_ENRICHMENT_TEST_REDIS")
	if dsn == "" || addr == "" {
		t.Skip("isolated PostgreSQL/Redis integration environment required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	var name string
	require.NoError(t, db.Raw("SELECT current_database()").Scan(&name).Error)
	require.Equal(t, "enrichment_test", name, "refusing to reset a non-test database")
	require.NoError(t, db.Exec(`DROP TABLE IF EXISTS enrichment_recovery_sources,enrichment_recoveries,enrichment_retry_settings,task_pending_ops,task_dead_letters,knowledge_processing_spans,chunks,knowledges,knowledge_bases CASCADE;
 CREATE TABLE knowledge_bases(id TEXT PRIMARY KEY,tenant_id BIGINT,deleted_at TIMESTAMPTZ);
 CREATE TABLE knowledges(id TEXT PRIMARY KEY,tenant_id BIGINT,knowledge_base_id TEXT,parse_status TEXT,pending_subtasks_count INT DEFAULT 0,updated_at TIMESTAMPTZ DEFAULT NOW(),deleted_at TIMESTAMPTZ);
 CREATE TABLE chunks(id TEXT PRIMARY KEY,knowledge_id TEXT,deleted_at TIMESTAMPTZ);
 CREATE TABLE knowledge_processing_spans(id BIGSERIAL PRIMARY KEY,knowledge_id TEXT,attempt INT,name TEXT,status TEXT,updated_at TIMESTAMPTZ DEFAULT NOW());
 CREATE TABLE task_pending_ops(id BIGSERIAL PRIMARY KEY,tenant_id BIGINT,task_type TEXT,scope TEXT,scope_id TEXT,op TEXT,dedup_key TEXT,payload JSONB,fail_count INT DEFAULT 0,enqueued_at TIMESTAMPTZ DEFAULT NOW(),claimed_at TIMESTAMPTZ);
 CREATE TABLE task_dead_letters(id BIGSERIAL PRIMARY KEY,tenant_id BIGINT,task_type TEXT,scope TEXT,scope_id TEXT,related_id TEXT,payload JSONB,last_error TEXT,fail_count INT,failed_at TIMESTAMPTZ DEFAULT NOW());`).Error)
	migration, err := os.ReadFile("../../../migrations/versioned/000092_enrichment_retry.up.sql")
	require.NoError(t, err)
	require.NoError(t, db.Exec(string(migration)).Error)
	require.True(t, strings.Contains(addr, "retry-redis"), "refusing to reset a non-test Redis")
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()
	require.NoError(t, rdb.FlushDB(context.Background()).Err())
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: addr})
	defer client.Close()
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: addr})
	defer inspector.Close()
	s := NewEnrichmentRecovery(db, client, inspector)
	enrichmentRecovery.Store(s)
	t.Cleanup(func() { enrichmentRecovery.Store(nil) })
	ctx := context.Background()
	require.NoError(t, db.Exec("INSERT INTO knowledge_bases(id,tenant_id) VALUES (?,1)", recoveryBackfillKB).Error)
	seed := func(kind string) (string, EnrichmentRecoveryRow) {
		kid := uuid.NewString()
		chunk := uuid.NewString()
		require.NoError(t, db.Exec("INSERT INTO knowledges(id,tenant_id,knowledge_base_id,parse_status) VALUES (?,1,?,'completed')", kid, recoveryBackfillKB).Error)
		require.NoError(t, db.Exec("INSERT INTO chunks(id,knowledge_id) VALUES (?,?)", chunk, kid).Error)
		require.NoError(t, db.Exec("INSERT INTO knowledge_processing_spans(knowledge_id,attempt,name,status) VALUES (?,1,'postprocess.wiki','failed')", kid).Error)
		var payload []byte
		scope, scopeID := types.TaskScopeKnowledge, kid
		if kind == types.TypeChunkExtract {
			payload, _ = json.Marshal(types.ExtractChunkPayload{TenantID: 1, ChunkID: chunk, ModelID: "m", KnowledgeID: kid, Attempt: 1})
		} else {
			payload, _ = json.Marshal(WikiPendingOp{Op: WikiOpIngest, KnowledgeID: kid})
			scope, scopeID = wikiTaskScope, recoveryBackfillKB
		}
		dl := types.TaskDeadLetter{TenantID: 1, TaskType: kind, Scope: scope, ScopeID: scopeID, RelatedID: kid, Payload: payload, LastError: "upstream refused", FailCount: 6, FailedAt: time.Now().Add(-24 * time.Hour)}
		require.NoError(t, db.Create(&dl).Error)
		require.NoError(t, s.discover(ctx))
		var row EnrichmentRecoveryRow
		require.NoError(t, db.Where("source_dead_letter_id=?", dl.ID).First(&row).Error)
		require.True(t, row.AvailableAt.After(time.Now().Add(119*time.Minute)))
		return kid, row
	}
	count := func(kid string) int {
		var n int
		require.NoError(t, db.Raw("SELECT pending_subtasks_count FROM knowledges WHERE id=?", kid).Scan(&n).Error)
		return n
	}
	ready := func(id int64) {
		require.NoError(t, db.Model(&EnrichmentRecoveryRow{}).Where("id=?", id).Update("available_at", time.Now().Add(-time.Second)).Error)
	}

	t.Run("graph durable outbox fencing and success", func(t *testing.T) {
		kid, row := seed(types.TypeChunkExtract)
		ready(row.ID)
		require.NoError(t, s.dispatch(ctx, row.ID))
		require.Equal(t, 1, count(kid))
		// A new service instance redispatches the same outbox without adding a counter.
		restart := NewEnrichmentRecovery(db, client, inspector)
		require.NoError(t, restart.dispatch(ctx, row.ID))
		require.Equal(t, 1, count(kid))
		require.NoError(t, s.Complete(ctx, row.ID, 1, errors.New("temporary failure")))
		require.NoError(t, db.First(&row, row.ID).Error)
		require.Equal(t, "ready", row.State)
		require.Equal(t, 1, row.Attempts)
		require.Equal(t, 1, count(kid))
		require.True(t, row.AvailableAt.After(time.Now().Add(119*time.Minute)))
		ready(row.ID)
		require.NoError(t, s.dispatch(ctx, row.ID))
		require.NoError(t, s.Complete(ctx, row.ID, 1, errors.New("late callback")))
		require.NoError(t, db.First(&row, row.ID).Error)
		require.Equal(t, "queued", row.State)
		require.Equal(t, 2, row.Attempts)
		require.NoError(t, s.Complete(ctx, row.ID, 2, nil))
		require.NoError(t, s.Complete(ctx, row.ID, 2, nil))
		require.Equal(t, 0, count(kid))
		require.NoError(t, db.First(&row, row.ID).Error)
		require.Equal(t, "recovered", row.State)
	})
	t.Run("wiki exactly two extra attempts without recursive dead letters", func(t *testing.T) {
		kid, row := seed(types.TypeWikiIngest)
		repo := repository.NewTaskPendingOpsRepository(db)
		wiki := &wikiIngestService{pendingRepo: repo}
		for n := 1; n <= 2; n++ {
			ready(row.ID)
			require.NoError(t, s.dispatch(ctx, row.ID))
			require.Equal(t, 1, count(kid))
			var records []types.TaskPendingOp
			require.NoError(t, db.Where("dedup_key=?", kid).Find(&records).Error)
			require.Len(t, records, 1)
			var op WikiPendingOp
			require.NoError(t, json.Unmarshal(records[0].Payload, &op))
			op.dbID = records[0].ID
			op.LastError = "bad JSON"
			require.Equal(t, int64(n), op.RecoveryDispatch)
			require.NoError(t, wiki.requeueFailedOps(ctx, WikiIngestPayload{TenantID: 1, KnowledgeBaseID: recoveryBackfillKB}, []WikiPendingOp{op}))
		}
		require.NoError(t, db.First(&row, row.ID).Error)
		require.Equal(t, "exhausted", row.State)
		require.Equal(t, 2, row.Attempts)
		require.Equal(t, 0, count(kid))
		require.NoError(t, s.dispatch(ctx, row.ID))
		require.Equal(t, 0, count(kid))
		var n int64
		require.NoError(t, db.Model(&types.TaskDeadLetter{}).Where("related_id=?", kid).Count(&n).Error)
		require.Equal(t, int64(1), n)
	})
	t.Run("postgres FIFO cooldown blocks late siblings", func(t *testing.T) {
		repo := repository.NewTaskPendingOpsRepository(db)
		ops := []*types.TaskPendingOp{}
		for _, key := range []string{"A", "B", "C"} {
			op := &types.TaskPendingOp{TenantID: 1, TaskType: wikiTaskType, Scope: wikiTaskScope, ScopeID: "fifo", Op: WikiOpIngest, DedupKey: key}
			require.NoError(t, repo.Enqueue(ctx, op))
			ops = append(ops, op)
		}
		_, err := repo.IncrFailCount(ctx, ops[0].ID)
		require.NoError(t, err)
		require.NoError(t, repo.Enqueue(ctx, &types.TaskPendingOp{TenantID: 1, TaskType: wikiTaskType, Scope: wikiTaskScope, ScopeID: "fifo", Op: WikiOpIngest, DedupKey: "A"}))
		rows, err := repo.ClaimBatch(ctx, wikiTaskType, wikiTaskScope, "fifo", 10, time.Now().Add(-time.Hour))
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.Equal(t, "B", rows[0].DedupKey)
		require.Equal(t, "C", rows[1].DedupKey)
		for _, r := range rows {
			require.False(t, strings.HasPrefix(r.DedupKey, "A"))
		}
	})
	t.Run("superseded recovery never decrements new attempt", func(t *testing.T) {
		kid, row := seed(types.TypeChunkExtract)
		ready(row.ID)
		require.NoError(t, s.dispatch(ctx, row.ID))
		require.NoError(t, db.Exec("INSERT INTO knowledge_processing_spans(knowledge_id,attempt,name,status) VALUES (?,2,'docreader','running')", kid).Error)
		require.NoError(t, db.Exec("UPDATE knowledges SET pending_subtasks_count=7 WHERE id=?", kid).Error)
		require.NoError(t, s.Complete(ctx, row.ID, 1, nil))
		require.Equal(t, 7, count(kid))
		require.NoError(t, db.First(&row, row.ID).Error)
		require.Equal(t, "skipped", row.State)
	})
	t.Run("pause consumes neither recovery budget nor counter", func(t *testing.T) {
		kid, row := seed(types.TypeChunkExtract)
		ready(row.ID)
		control := processingcontrol.New(rdb)
		processingcontrol.Install(control)
		defer processingcontrol.Install(nil)
		_, err := control.SetPaused(ctx, 1, recoveryBackfillKB, true)
		require.NoError(t, err)
		defer control.SetPaused(ctx, 1, recoveryBackfillKB, false)
		require.NoError(t, s.dispatch(ctx, row.ID))
		require.NoError(t, db.First(&row, row.ID).Error)
		require.Equal(t, "ready", row.State)
		require.Zero(t, row.Attempts)
		require.Zero(t, count(kid))
		require.Empty(t, filterPausedHousekeepingCandidates(ctx, []types.Knowledge{{ID: kid, TenantID: 1, KnowledgeBaseID: recoveryBackfillKB}}))
	})

}

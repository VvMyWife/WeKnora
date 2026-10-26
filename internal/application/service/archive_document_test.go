package service

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"
)

type archiveRetryRepo struct {
	interfaces.KnowledgeRepository
	document *types.Knowledge
}

func (r archiveRetryRepo) GetKnowledgeByID(context.Context, uint64, string) (*types.Knowledge, error) {
	return r.document, nil
}

func TestArchiveRetryStage(t *testing.T) {
	for input, want := range map[string]string{"": "all", "postprocess.graph.chunk[3]": "graph", "postprocess.wiki": "wiki", "summary": "summary", "embedding": "parse"} {
		got, err := archiveRetryStage(input)
		require.NoError(t, err)
		require.Equal(t, want, got)
	}
	_, err := archiveRetryStage("arbitrary")
	require.Error(t, err)
}

func TestArchiveManualSummaryRetryDurableAndDuplicateSafe(t *testing.T) {
	conn, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer conn.Close()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: conn}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)
	recovery := NewEnrichmentRecovery(db, nil, nil)
	enrichmentRecovery.Store(recovery)
	defer enrichmentRecovery.Store(nil)
	doc := &types.Knowledge{ID: "doc", TenantID: 1, KnowledgeBaseID: "kb", ParseStatus: types.ParseStatusCompleted, SummaryStatus: types.SummaryStatusFailed}
	svc := &knowledgeService{repo: archiveRetryRepo{document: doc}}
	ctx := context.WithValue(context.Background(), types.TenantIDContextKey, uint64(1))
	for round := 0; round < 2; round++ {
		mock.ExpectQuery("SELECT COALESCE").WillReturnRows(sqlmock.NewRows([]string{"attempt"}).AddRow(1))
		mock.ExpectQuery("SELECT DISTINCT ON").WillReturnRows(sqlmock.NewRows([]string{"name", "status"}).AddRow("postprocess.summary", "failed"))
		mock.ExpectQuery("SELECT .*task_dead_letters").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		mock.ExpectBegin()
		inserted := sqlmock.NewRows([]string{"id"})
		if round == 0 {
			inserted.AddRow(7)
		}
		mock.ExpectQuery("INSERT INTO .*enrichment_recoveries").WillReturnRows(inserted)
		mock.ExpectQuery("SELECT .*enrichment_recoveries.*FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id", "logical_key", "state", "available_at"}).AddRow(7, "summary:1:doc:1", "ready", time.Now().Add(-time.Second)))
		if round == 0 {
			mock.ExpectExec("UPDATE .*enrichment_recoveries").WillReturnResult(sqlmock.NewResult(0, 1))
		}
		mock.ExpectCommit()
		if round == 0 {
			mock.ExpectQuery("SELECT .*enrichment_recoveries").WillReturnRows(sqlmock.NewRows([]string{"id"}))
		}
		result, err := svc.RetryArchiveFailures(ctx, "doc", "summary")
		require.NoError(t, err)
		require.Equal(t, 1-round, result["accepted"])
	}
	require.NoError(t, mock.ExpectationsWereMet())
}

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tencent/WeKnora/internal/archivalcitations"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http/httptest"
	"testing"
)

type citationMessageStub struct {
	interfaces.MessageService
	denied bool
}

func (s *citationMessageStub) GetMessage(context.Context, string, string) (*types.Message, error) {
	if s.denied {
		return nil, errors.New("denied")
	}
	return &types.Message{ID: "m", SessionID: "s", Role: "assistant"}, nil
}

type citationKnowledgeStub struct {
	interfaces.KnowledgeService
	denied bool
}

func (s *citationKnowledgeStub) GetKnowledgeByIDOnly(context.Context, string) (*types.Knowledge, error) {
	if s.denied {
		return nil, errors.New("deleted")
	}
	return &types.Knowledge{ID: "k", TenantID: 1, KnowledgeBaseID: "kb"}, nil
}
func TestCitationReadsRespectMessageAndCurrentSourceAccess(t *testing.T) {
	t.Setenv("ARCHIVAL_CITATIONS_ENABLED", "true")
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, e)
	require.NoError(t, db.AutoMigrate(&archivalcitations.Record{}, &archivalcitations.ManifestRecord{}))
	// Keep the fixture independent from production vector/JSON database columns.
	require.NoError(t, db.Exec("CREATE TABLE chunks(id TEXT PRIMARY KEY, tenant_id INTEGER, knowledge_id TEXT, content TEXT, content_revision INTEGER, deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO chunks VALUES('c',1,'k','new text',2,NULL)").Error)
	store := archivalcitations.NewStore(db)
	sum := archivalcitations.Summary{ID: archivalcitations.EvidenceID(1, "s", "m", "c", "old text", "", 1), Number: 1, ChunkID: "c", KnowledgeID: "k", KnowledgeBaseID: "kb", ContentRevision: 1, ContentSHA256: archivalcitations.Hash("old text"), Location: archivalcitations.Location{Status: archivalcitations.PageMissing, Pages: []int{}}}
	require.NoError(t, store.Save(context.Background(), 1, "s", "m", "answer", archivalcitations.Manifest{Version: 1, State: "verified", Citations: []archivalcitations.Summary{sum}}, []archivalcitations.Record{{ID: sum.ID, TenantID: 1, KnowledgeTenantID: 1, SessionID: "s", MessageID: "m", KnowledgeID: "k", ChunkID: "c", Snapshot: "old text"}}))
	messages := &citationMessageStub{}
	sources := &citationKnowledgeStub{}
	h := AttachMessageCitationStore(&MessageHandler{MessageService: messages}, store, &KnowledgeHandler{kgService: sources})
	run := func(detail bool) (*httptest.ResponseRecorder, *gin.Context) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil).WithContext(context.WithValue(context.Background(), types.TenantIDContextKey, uint64(1)))
		c.Set(types.TenantIDContextKey.String(), uint64(1))
		c.Params = gin.Params{{Key: "session_id", Value: "s"}, {Key: "id", Value: "m"}, {Key: "citation_id", Value: sum.ID}}
		if detail {
			h.GetMessageCitation(c)
		} else {
			h.GetMessageCitations(c)
		}
		return w, c
	}
	w, c := run(true)
	require.Empty(t, c.Errors)
	require.Contains(t, w.Body.String(), "old text")
	require.Contains(t, w.Body.String(), "\"current_changed\":true")
	sources.denied = true
	w, c = run(true)
	require.NotEmpty(t, c.Errors)
	require.NotContains(t, w.Body.String(), "old text")
	w, c = run(false)
	require.Empty(t, c.Errors)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Equal(t, float64(1), body["unavailable_sources"])
	messages.denied = true
	w, c = run(false)
	require.NotEmpty(t, c.Errors)
	require.NotContains(t, w.Body.String(), sum.ID)
	t.Setenv("ARCHIVAL_CITATIONS_ENABLED", "false")
	w, _ = run(true)
	require.Equal(t, 503, w.Code)
}

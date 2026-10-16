package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/externalchunks"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/hibiken/asynq"
)

type externalModelFixture struct {
	interfaces.ModelService
	value *types.Model
}

func (m externalModelFixture) GetModelByID(context.Context, string) (*types.Model, error) {
	return m.value, nil
}
func TestExternalEmbeddingChecksActualProviderAndRejectsTruncation(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		wantError  bool
	}{
		{"fit", `{"max_input_tokens":8192,"input_tokens":[50],"truncation":false}`, 200, false},
		{"oversize", `{"max_input_tokens":8192,"input_tokens":[10000],"truncation":false}`, 200, true},
		{"no truncation proof", `{"max_input_tokens":8192,"input_tokens":[50],"truncation":true}`, 200, true},
		{"wrong length", `{"max_input_tokens":8192,"input_tokens":[],"truncation":false}`, 200, true},
		{"unsupported", `{}`, 404, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/token-counts" {
					t.Errorf("unexpected path %s", r.URL.Path)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			model := &types.Model{Name: "bge-m3", Parameters: types.ModelParameters{BaseURL: server.URL + "/v1"}}
			s := &knowledgeService{modelService: externalModelFixture{value: model}}
			kb := &types.KnowledgeBase{EmbeddingModelID: "embed", IndexingStrategy: types.DefaultIndexingStrategy()}
			k := &types.Knowledge{Title: "样例"}
			p := &externalchunks.DocumentPayload{Document: externalchunks.Document{ID: "doc"}, Chunks: []externalchunks.Chunk{{ID: "chunk", Content: "原文"}}}
			e := s.checkExternalEmbeddingLimits(context.Background(), kb, k, p)
			if (e != nil) != tc.wantError {
				t.Fatalf("error=%v expected=%v", e, tc.wantError)
			}
			if tc.wantError && !errors.Is(e, asynq.SkipRetry) {
				t.Fatalf("permanent incompatibility must not retry forever: %v", e)
			}
			model.Parameters.EmbeddingParameters.TruncatePromptTokens = 512
			if e = s.checkExternalEmbeddingLimits(context.Background(), kb, k, p); e == nil {
				t.Fatal("accepted configured truncation")
			}
		})
	}
}

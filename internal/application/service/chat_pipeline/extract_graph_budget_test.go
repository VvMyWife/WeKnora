package chatpipeline

import (
	"context"
	"testing"

	"github.com/Tencent/WeKnora/internal/models/chat"
	"github.com/Tencent/WeKnora/internal/types"
)

type graphBudgetChat struct {
	options chat.ChatOptions
	content string
}

func (m *graphBudgetChat) ChatStream(context.Context, []chat.Message, *chat.ChatOptions) (<-chan types.StreamResponse, error) {
	panic("streaming is not used by graph extraction")
}

func (m *graphBudgetChat) GetModelName() string { return "graph-budget-test" }
func (m *graphBudgetChat) GetModelID() string   { return "graph-budget-test" }

func (m *graphBudgetChat) Chat(_ context.Context, _ []chat.Message, opts *chat.ChatOptions) (*types.ChatResponse, error) {
	m.options = *opts
	return &types.ChatResponse{Content: m.content}, nil
}

func TestExtractionCompletionBudgets(t *testing.T) {
	for _, tc := range []struct {
		name  string
		build func(chat.Chat, *types.PromptTemplateStructured) Extractor
		want  int
	}{
		{"background_graph", NewGraphExtractor, 12288},
		{"query_entities", NewExtractor, 8192},
	} {
		t.Run(tc.name, func(t *testing.T) {
			model := &graphBudgetChat{content: `[{"entity":"档案管理办法"}]`}
			extractor := tc.build(model, &types.PromptTemplateStructured{Description: "Extract entities as JSON."})
			graph, err := extractor.Extract(context.Background(), "档案管理办法")
			if err != nil {
				t.Fatal(err)
			}
			if model.options.MaxTokens != tc.want {
				t.Fatalf("completion budget = %d, want %d", model.options.MaxTokens, tc.want)
			}
			if model.options.Thinking == nil || *model.options.Thinking {
				t.Fatal("extraction must keep thinking disabled")
			}
			if len(graph.Node) != 1 || graph.Node[0].Name != "档案管理办法" {
				t.Fatalf("graph parsing changed: %+v", graph)
			}
		})
	}
}

func TestGraphExtractorRejectsInvalidJSON(t *testing.T) {
	model := &graphBudgetChat{content: `[{"entity":"档案管理办法"}}]`}
	extractor := NewGraphExtractor(model, &types.PromptTemplateStructured{Description: "Extract entities as JSON."})
	if _, err := extractor.Extract(context.Background(), "档案管理办法"); err == nil {
		t.Fatal("increasing output budget must not turn malformed JSON into a successful graph")
	}
}

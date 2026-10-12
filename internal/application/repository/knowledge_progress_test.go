package repository

import (
	"github.com/Tencent/WeKnora/internal/types"
	"strings"
	"testing"
	"time"
)

func TestCurrentProgressFailureAndRecovery(t *testing.T) {
	docs := []types.Knowledge{{ID: "retry", ParseStatus: "processing", FolderPath: "a"}, {ID: "ok", ParseStatus: "completed", SummaryStatus: "completed", FolderPath: "b"}, {ID: "bad", ParseStatus: "failed", ErrorMessage: "PDF 共 16 页，实际处理 1 页；缺少第 2–16 页"}}
	spans := []types.KnowledgeProcessingSpan{{KnowledgeID: "retry", Name: "docreader", Kind: "stage", Status: "failed", ErrorCode: "ocr_error", UpdatedAt: time.Now()}, {KnowledgeID: "ok", Name: "docreader", Kind: "stage", Status: "running"}}
	overview, failures := projectKnowledgeProgress(docs, spans)
	if overview.FailuresTotal != 2 || len(failures) != 2 || len(overview.ActiveDocuments) != 1 {
		t.Fatalf("wrong current counts: %+v", overview)
	}
	if overview.Stages["docreader"]["running"] != 0 {
		t.Fatal("historical running state leaked")
	}
	docs[0].ParseStatus = "completed"
	spans[0].Status = "done"
	overview, _ = projectKnowledgeProgress(docs, spans)
	if overview.FailuresTotal != 1 {
		t.Fatal("recovered error still counted")
	}
	if strings.Contains(safeProgressReason("Authorization: Bearer secret; request body private", "err"), "secret") {
		t.Fatal("secret leaked")
	}
}

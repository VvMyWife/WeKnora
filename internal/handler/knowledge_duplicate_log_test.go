package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
)

func TestDuplicateFileResponseIdentifiesExistingDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/upload", nil)
	existing := &types.Knowledge{
		ID: "existing-id", KnowledgeBaseID: "kb-id",
		FolderPath: "other/folder", FileName: "copy.md",
	}
	h := &KnowledgeHandler{}
	if !h.handleDuplicateKnowledgeError(c, types.NewDuplicateFileError(existing), existing,
		"file", "new/folder/new.md") {
		t.Fatal("duplicate was not handled")
	}
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
	var body struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != "duplicate_file" || !strings.Contains(body.Message, "other/folder/copy.md") {
		t.Fatalf("response does not identify the duplicate: %+v", body)
	}
}

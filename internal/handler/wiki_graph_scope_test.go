package handler

import (
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
)

func TestParseDocumentGraphScope(t *testing.T) {
	for _, tc := range []struct {
		query string
		valid bool
	}{
		{"", true},
		{"scope=documents", false},
		{"scope=documents&knowledge_ids=bad", false},
		{"scope=other", false},
		{"scope=kb&knowledge_ids=0d0b4eb2-e860-405f-b0a7-af020c17ecf9", false},
		{"scope=documents&view=connections&knowledge_ids=0d0b4eb2-e860-405f-b0a7-af020c17ecf9", true},
	} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/?"+tc.query, nil)
		req := &types.WikiGraphRequest{Limit: 900}
		err := parseWikiGraphScope(c, req)
		if (err == nil) != tc.valid {
			t.Errorf("%s: %v", tc.query, err)
		}
		if err == nil && req.Scope == "documents" && req.Limit != 200 {
			t.Errorf("default limit: %d", req.Limit)
		}
	}
}

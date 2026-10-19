package handler

import (
	"fmt"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func parseWikiGraphScope(c *gin.Context, req *types.WikiGraphRequest) error {
	req.Scope, req.View = c.DefaultQuery("scope", "kb"), c.DefaultQuery("view", "union")
	if req.Scope != "kb" && req.Scope != "documents" {
		return fmt.Errorf("invalid graph scope")
	}
	if req.View != "union" && req.View != "connections" {
		return fmt.Errorf("invalid graph view")
	}
	ids := c.QueryArray("knowledge_ids")
	if req.Scope == "kb" {
		if len(ids) > 0 || req.View != "union" {
			return fmt.Errorf("document filters require documents scope")
		}
		return nil
	}
	if len(ids) == 0 || len(ids) > 50 {
		return fmt.Errorf("select between 1 and 50 documents")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil || parsed.String() != id {
			return fmt.Errorf("invalid document ID")
		}
		if !seen[id] {
			req.KnowledgeIDs = append(req.KnowledgeIDs, id)
			seen[id] = true
		}
	}
	if c.Query("limit") == "" {
		req.Limit = 200
	}
	if req.Limit > 500 {
		req.Limit = 500
	}
	return nil
}

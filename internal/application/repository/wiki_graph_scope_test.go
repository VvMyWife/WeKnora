package repository

import (
	"context"
	"errors"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListDocumentGraphExactProjectionAndOwnership(t *testing.T) {
	db := setupWikiPagesTestDB(t)
	require.NoError(t, db.Exec(`CREATE TABLE knowledges (id TEXT PRIMARY KEY, knowledge_base_id TEXT, deleted_at DATETIME)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO knowledges VALUES ('a','kb',NULL),('b','other',NULL),('deleted','kb',CURRENT_TIMESTAMP)`).Error)
	for _, p := range []*types.WikiPage{
		{ID: "1", KnowledgeBaseID: "kb", Slug: "yes", SourceRefs: types.StringArray{"a|Title"}, Content: "must not load", TenantID: 1},
		{ID: "2", KnowledgeBaseID: "kb", Slug: "prefix", SourceRefs: types.StringArray{"ab|Title"}, TenantID: 1},
		{ID: "3", KnowledgeBaseID: "other", Slug: "foreign", SourceRefs: types.StringArray{"a"}, TenantID: 1},
	} {
		require.NoError(t, db.Create(p).Error)
	}
	r := &wikiPageRepository{db: db}
	// Historical index/derived pages may contain JSON null or scalar refs.
	// None is a source array; queries must skip them rather than fail the KB.
	for _, raw := range []string{"null", `"a"`, `42`, `{}`, `[]`} {
		require.NoError(t, db.Exec(`INSERT INTO wiki_pages(id,tenant_id,knowledge_base_id,slug,source_refs) VALUES (?,1,'kb',?,?)`, "legacy-"+raw, "legacy-"+raw, raw).Error)
	}
	pages, err := r.ListDocumentGraph(context.Background(), "kb", []string{"a", "a"})
	require.NoError(t, err)
	require.Len(t, pages, 1)
	require.Equal(t, "yes", pages[0].Slug)
	require.Empty(t, pages[0].Content)
	for _, ids := range [][]string{nil, {"b"}, {"deleted"}, {"a", "missing"}} {
		_, err = r.ListDocumentGraph(context.Background(), "kb", ids)
		require.True(t, errors.Is(err, ErrInvalidWikiGraphDocuments))
	}
}

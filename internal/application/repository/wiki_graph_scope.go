package repository

import (
	"context"
	"errors"

	"github.com/Tencent/WeKnora/internal/types"
)

var ErrInvalidWikiGraphDocuments = errors.New("selected documents are unavailable in this knowledge base")

// ListDocumentGraph projects only graph columns. KB read authorization is checked
// by the route; this query additionally rejects deleted and foreign-KB documents.
func (r *wikiPageRepository) ListDocumentGraph(ctx context.Context, kbID string, ids []string) ([]*types.WikiPage, error) {
	unique := map[string]bool{}
	for _, id := range ids {
		unique[id] = true
	}
	if len(unique) == 0 || len(unique) > 50 {
		return nil, ErrInvalidWikiGraphDocuments
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&types.Knowledge{}).
		Where("knowledge_base_id = ? AND id IN ?", kbID, ids).Count(&count).Error; err != nil {
		return nil, err
	}
	if count != int64(len(unique)) {
		return nil, ErrInvalidWikiGraphDocuments
	}
	query := r.db.WithContext(ctx).Model(&types.WikiPage{}).
		Select("slug, title, page_type, source_refs, in_links, out_links").
		Where("knowledge_base_id = ? AND status <> ?", kbID, types.WikiPageStatusArchived)
	if r.db.Dialector.Name() == "sqlite" {
		query = query.Where(`EXISTS (SELECT 1 FROM json_each(source_refs) AS source WHERE trim(CASE WHEN instr(source.value, '|') > 0 THEN substr(source.value, 1, instr(source.value, '|') - 1) ELSE source.value END) IN ?)`, ids)
	} else {
		query = query.Where(`EXISTS (SELECT 1 FROM jsonb_array_elements_text(COALESCE(source_refs::jsonb, '[]'::jsonb)) AS source(value) WHERE btrim(split_part(source.value, '|', 1)) IN ?)`, ids)
	}
	var pages []*types.WikiPage
	if err := query.Find(&pages).Error; err != nil {
		return nil, err
	}
	return pages, nil
}

package service

import (
	"context"
	"encoding/json"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/stretchr/testify/require"
	"testing"
)

type pdfIntegrityMetadataRepo struct {
	interfaces.KnowledgeRepository
	metadata types.JSON
	writes   int
}

func (r *pdfIntegrityMetadataRepo) UpdateKnowledgeColumn(_ context.Context, _ string, column string, value interface{}) error {
	if column == "metadata" {
		r.metadata = append(types.JSON(nil), value.(types.JSON)...)
		r.writes++
	}
	return nil
}
func TestPDFIntegrityPreservesStructuredProcessOverrides(t *testing.T) {
	repo := &pdfIntegrityMetadataRepo{}
	s := &knowledgeService{repo: repo}
	k := &types.Knowledge{ID: "pdf", Metadata: types.JSON(`{"process_overrides":{"summary_enabled":false,"parser_engine_rules":[{"file_types":["pdf"],"engine":"paddleocr_vl"}],"parser_engine_overrides":{"paddleocr_vl_endpoint":"http://ocr.test"}},"external":{"anchors":[1,true,null],"counter":9007199254740993},"source":"archive"}`)}
	for _, state := range []string{"pending", "failed", "verified"} {
		s.setPDFPageIntegrity(context.Background(), k, state, map[string]string{"pdf_expected_pages": "2", "pdf_processed_pages": "2"})
		overrides, err := k.ProcessOverrides()
		require.NoError(t, err)
		require.NotNil(t, overrides)
		require.NotNil(t, overrides.SummaryEnabled)
		require.False(t, *overrides.SummaryEnabled)
		require.Equal(t, "paddleocr_vl", overrides.ParserEngineRules[0].Engine)
		require.Equal(t, "http://ocr.test", overrides.ParserEngineOverrides["paddleocr_vl_endpoint"])
		var values map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(k.Metadata, &values))
		require.JSONEq(t, `{"anchors":[1,true,null],"counter":9007199254740993}`, string(values["external"]))
		var ext map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(values["external"], &ext))
		require.Equal(t, "9007199254740993", string(ext["counter"]))
		require.JSONEq(t, `"archive"`, string(values["source"]))
		require.JSONEq(t, `"`+state+`"`, string(values["pdf_page_integrity"]))
		require.JSONEq(t, string(k.Metadata), string(repo.metadata))
	}
	require.Equal(t, 3, repo.writes)
}
func TestPDFIntegrityRejectsInvalidMetadataWithoutOverwriting(t *testing.T) {
	for _, raw := range []string{`{"process_overrides":`, `[]`, `"legacy scalar"`} {
		repo := &pdfIntegrityMetadataRepo{}
		s := &knowledgeService{repo: repo}
		k := &types.Knowledge{ID: "invalid", Metadata: types.JSON(raw)}
		s.setPDFPageIntegrity(context.Background(), k, "verified", nil)
		require.Equal(t, raw, string(k.Metadata))
		require.Zero(t, repo.writes)
	}
}

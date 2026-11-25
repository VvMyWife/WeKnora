package archivalcitations

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestCitationPreviewUsesAtMostThreeForwardChunks(t *testing.T) {
	for _, tc := range []struct {
		name      string
		match     int
		breakLink bool
		external  bool
		want      *PreviewLocation
	}{
		{"cited chunk", 0, false, false, &PreviewLocation{Pages: []int{17}}},
		{"first next chunk", 1, false, false, &PreviewLocation{Pages: []int{17}, Hops: 1}},
		{"second next chunk", 2, false, false, &PreviewLocation{Pages: []int{17}, Hops: 2}},
		{"third next chunk", 3, false, true, &PreviewLocation{Pages: []int{17}, Hops: 3}},
		{"fourth next chunk", 4, false, false, nil},
		{"broken link", 1, true, false, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := requestFixture(t).store
			db := store.DB
			for _, ddl := range []string{
				"CREATE TABLE archive_mapping_versions (id TEXT, tenant_id INTEGER, knowledge_id TEXT, sha256 TEXT)",
				"CREATE TABLE external_chunk_records (id TEXT, tenant_id INTEGER, knowledge_id TEXT, payload TEXT)",
			} {
				if err := db.Exec(ddl).Error; err != nil {
					t.Fatal(err)
				}
			}
			if err := db.AutoMigrate(&types.Chunk{}); err != nil {
				t.Fatal(err)
			}
			if err := db.Exec("INSERT INTO archive_mapping_versions VALUES ('v', 1, 'k', 'pdf-hash')").Error; err != nil {
				t.Fatal(err)
			}
			if err := db.Exec("INSERT INTO external_chunk_records VALUES ('r', 1, 'k', ?)", `{"document":{"source_sha256":"pdf-hash"}}`).Error; err != nil {
				t.Fatal(err)
			}

			for i := 0; i < 5; i++ {
				chunk := types.Chunk{
					ID: fmt.Sprint(i), TenantID: 1, KnowledgeID: "k", KnowledgeBaseID: "kb", ChunkIndex: i,
					PreChunkID: fmt.Sprint(i - 1), NextChunkID: fmt.Sprint(i + 1), ChunkType: types.ChunkTypeText,
				}
				if tc.breakLink && i == 1 {
					chunk.PreChunkID = "different"
				}
				switch {
				case tc.external && i == tc.match:
					chunk.Metadata = types.JSON(`{"external_import":{"record_id":"r","source_anchors":[{"page_index":16}]}}`)
				case tc.external:
					chunk.Metadata = types.JSON(`{"external_import":{"record_id":"r"}}`)
				case i == tc.match:
					chunk.SourceLocators = types.SourceLocators{
						{Type: types.SourceLocatorPDF, Page: 17, SourceHash: "pdf-hash"},
						{Type: types.SourceLocatorPDF, Page: 17, SourceHash: "pdf-hash"},
					}
				}
				if err := db.Create(&chunk).Error; err != nil {
					t.Fatal(err)
				}
			}
			summary := Summary{ChunkID: "0", KnowledgeID: "k", KnowledgeBaseID: "kb",
				ArchiveVersionID: "v", ArchiveFileName: "archive.pdf", ArchiveSHA256: "pdf-hash",
				SourceKind: "chunk", Location: Location{Status: PageMissing}}
			if got := store.CitationPreview(context.Background(), 1, summary); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %#v, want %#v", got, tc.want)
			}
			summary.ArchiveSHA256 = "different-hash"
			if got := store.CitationPreview(context.Background(), 1, summary); got != nil {
				t.Fatalf("mismatched archive: %#v", got)
			}
		})
	}
}

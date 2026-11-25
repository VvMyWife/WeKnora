package archivalcitations

import (
	"reflect"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
)

func TestPinnedPDFPages(t *testing.T) {
	first := types.SourceLocator{Type: types.SourceLocatorPDF, Page: 17, ArchiveVersionID: "version-a"}
	second := first
	second.Page = 18
	id, pages := pinnedPDFPages(types.SourceLocators{second, first, first})
	if id != "version-a" || !reflect.DeepEqual(pages, []int{17, 18}) {
		t.Fatalf("id=%q pages=%v", id, pages)
	}
	second.ArchiveVersionID = "version-b"
	id, pages = pinnedPDFPages(types.SourceLocators{first, second})
	if id != "" || len(pages) != 0 {
		t.Fatalf("mixed versions: %q %v", id, pages)
	}
	first.Partial = true
	id, pages = pinnedPDFPages(types.SourceLocators{first})
	if id != "" || len(pages) != 0 {
		t.Fatalf("partial mapping: %q %v", id, pages)
	}
}

package archivalcitations

import (
	"encoding/json"
	"testing"
)

func TestPagesNeverGuessed(t *testing.T) {
	cases := []struct {
		anchors                 []map[string]any
		source, archive, status string
		pages                   int
	}{
		{nil, "a", "a", PageMissing, 0},
		{[]map[string]any{{"page_index": float64(3)}, {"page_index": 0}, {"page_index": 3}}, "a", "a", PageKnown, 2},
		{[]map[string]any{{"page_index": 2}}, "a", "b", PageMismatch, 0},
		{[]map[string]any{{"page_index": -1}}, "a", "a", PageInvalid, 0},
		{[]map[string]any{{"page_index": 1.2}}, "a", "a", PageInvalid, 0},
		{[]map[string]any{{"page_index": 2}}, "", "a", PageMissing, 0},
		{[]map[string]any{{"chunk_index": 4, "file_name": "page_9.jpg"}}, "a", "a", PageMissing, 0},
	}
	for _, c := range cases {
		got := PagesFromAnchors(c.anchors, c.source, c.archive)
		if got.Status != c.status || len(got.Pages) != c.pages {
			t.Fatalf("%+v", got)
		}
		if e := ValidateLocation(got); e != nil {
			t.Fatal(e)
		}
	}
}
func TestVersionedIdentityAndWireCompatibility(t *testing.T) {
	a := EvidenceID(1, "s", "m", "c", "original", "v", 0)
	if len(a) != 64 || a != EvidenceID(1, "s", "m", "c", "original", "v", 0) {
		t.Fatal("unstable identity")
	}
	for _, b := range []string{EvidenceID(2, "s", "m", "c", "original", "v", 0), EvidenceID(1, "s", "m", "c", "edited", "v", 0), EvidenceID(1, "s", "m", "c", "original", "v2", 0)} {
		if a == b {
			t.Fatal("identity collision")
		}
	}
	var legacy struct {
		Content  string    `json:"content"`
		Manifest *Manifest `json:"citation_manifest,omitempty"`
	}
	if e := json.Unmarshal([]byte(`{"content":"old answer"}`), &legacy); e != nil || legacy.Manifest != nil {
		t.Fatal("legacy broken")
	}
	if e := ValidateLocation(Location{Pages: []int{2, 1}, Status: PageKnown}); e == nil {
		t.Fatal("unsorted pages")
	}
}
func TestFeatureOffByDefault(t *testing.T) {
	t.Setenv("ARCHIVAL_CITATIONS_ENABLED", "")
	if Enabled() {
		t.Fatal("unsafe rollout")
	}
	t.Setenv("ARCHIVAL_CITATIONS_ENABLED", "true")
	if !Enabled() {
		t.Fatal("flag ignored")
	}
}

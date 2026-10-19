// Package archivalcitations owns versioned answer evidence, not retrieval rankings.
package archivalcitations

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

const SchemaVersion = 1
const (
	PageKnown    = "known"
	PageMissing  = "missing"
	PageMismatch = "version_mismatch"
	PageInvalid  = "invalid"
)

// Enabled is deliberately off until migrations and both application surfaces ship.
func Enabled() bool { return strings.EqualFold(os.Getenv("ARCHIVAL_CITATIONS_ENABLED"), "true") }

type Location struct {
	Pages  []int  `json:"pages"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}
type Summary struct {
	ID               string   `json:"id"`
	Number           int      `json:"number"`
	ChunkID          string   `json:"chunk_id"`
	KnowledgeID      string   `json:"knowledge_id"`
	KnowledgeBaseID  string   `json:"knowledge_base_id"`
	Title            string   `json:"title"`
	ContentRevision  int      `json:"content_revision"`
	ContentSHA256    string   `json:"content_sha256"`
	ArchiveVersionID string   `json:"archive_version_id,omitempty"`
	ArchiveFileName  string   `json:"archive_file_name,omitempty"`
	ArchiveSHA256    string   `json:"archive_sha256,omitempty"`
	SourceKind       string   `json:"source_kind"`
	Location         Location `json:"location"`
}
type UnboundBlock struct {
	Index  int    `json:"index"`
	SHA256 string `json:"sha256"`
	Reason string `json:"reason"`
}
type Manifest struct {
	Version         int            `json:"version"`
	State           string         `json:"state"`
	Citations       []Summary      `json:"citations"`
	Unbound         []UnboundBlock `json:"unbound_blocks"`
	RepairAttempted bool           `json:"repair_attempted"`
}
type Detail struct {
	Summary
	Snapshot        string `json:"snapshot"`
	CurrentContent  string `json:"current_content,omitempty"`
	CurrentRevision int    `json:"current_revision"`
	CurrentChanged  bool   `json:"current_changed"`
	SnapshotAt      string `json:"snapshot_at"`
}

func Hash(text string) string { sum := sha256.Sum256([]byte(text)); return hex.EncodeToString(sum[:]) }
func EvidenceID(tenant uint64, session, message, chunk, content, archive string, revision int) string {
	raw, _ := json.Marshal([]any{tenant, session, message, chunk, revision, Hash(content), archive})
	return Hash(string(raw))
}

// PagesFromAnchors consumes only parser-provided, zero-based physical page indexes.
// File names, chunk indexes and printed page labels are never accepted as page evidence.
func PagesFromAnchors(anchors []map[string]any, sourceHash, archiveHash string) Location {
	result := Location{Pages: []int{}, Status: PageMissing, Reason: "未提供可信页码"}
	if len(anchors) == 0 {
		return result
	}
	if sourceHash == "" || archiveHash == "" {
		return result
	}
	if sourceHash != archiveHash {
		result.Status = PageMismatch
		result.Reason = "页码对应的原件版本不匹配"
		return result
	}
	seen := map[int]bool{}
	for _, anchor := range anchors {
		value, ok := anchor["page_index"]
		if !ok {
			continue
		}
		number, e := json.Number(fmt.Sprint(value)).Float64()
		if e != nil || math.IsNaN(number) || math.IsInf(number, 0) || number < 0 || number != math.Trunc(number) || number > 1000000 {
			return Location{Pages: []int{}, Status: PageInvalid, Reason: "来源页码格式无效"}
		}
		seen[int(number)+1] = true
	}
	for page := range seen {
		result.Pages = append(result.Pages, page)
	}
	sort.Ints(result.Pages)
	if len(result.Pages) > 0 {
		result.Status = PageKnown
		result.Reason = ""
	}
	return result
}
func ValidateLocation(value Location) error {
	switch value.Status {
	case PageKnown, PageMissing, PageMismatch, PageInvalid:
	default:
		return fmt.Errorf("unknown page status")
	}
	previous := 0
	for _, page := range value.Pages {
		if page <= previous {
			return fmt.Errorf("pages must be sorted, distinct and positive")
		}
		previous = page
	}
	if value.Status == PageKnown && len(value.Pages) == 0 {
		return fmt.Errorf("known location has no pages")
	}
	if value.Status != PageKnown && len(value.Pages) > 0 {
		return fmt.Errorf("untrusted location cannot expose pages")
	}
	return nil
}

// PhysicalPage is a boundary helper for APIs accepting a user-selected PDF page.
func PhysicalPage(raw string) (int, error) {
	page, e := strconv.Atoi(raw)
	if e != nil || page < 1 {
		return 0, fmt.Errorf("physical PDF page must be positive")
	}
	return page, nil
}

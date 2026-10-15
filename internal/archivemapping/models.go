// Package archivemapping stores original-file associations independently of
// knowledge/chunk processing. It never creates parsing or model tasks.
package archivemapping

import (
	"fmt"
	"path"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type Batch struct {
	ID              string    `json:"id" gorm:"primaryKey;size:36"`
	TenantID        uint64    `json:"-" gorm:"index"`
	KnowledgeBaseID string    `json:"knowledge_base_id" gorm:"index;size:36"`
	ParsedRoot      string    `json:"parsed_root"`
	CreatedBy       string    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
}

func (Batch) TableName() string { return "archive_mapping_batches" }

type Candidate struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Path   string `json:"path"`
	Mapped bool   `json:"mapped"`
}
type Item struct {
	ID           string      `json:"id" gorm:"primaryKey;size:36"`
	BatchID      string      `json:"batch_id" gorm:"index;size:36"`
	RelativePath string      `json:"relative_path"`
	FileSize     int64       `json:"file_size"`
	Candidates   []Candidate `json:"candidates" gorm:"serializer:json;type:text"`
	KnowledgeID  string      `json:"knowledge_id" gorm:"size:36"`
	Status       string      `json:"status"` // ready, unmatched, ambiguous, conflict, bound, skipped, failed
	Error        string      `json:"error"`
	VersionID    string      `json:"version_id" gorm:"size:36"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func (Item) TableName() string { return "archive_mapping_items" }

type Version struct {
	ID           string    `json:"id" gorm:"primaryKey;size:36"`
	TenantID     uint64    `json:"-" gorm:"index"`
	KnowledgeID  string    `json:"knowledge_id" gorm:"index;size:36"`
	FileName     string    `json:"file_name"`
	RelativePath string    `json:"relative_path"`
	FileType     string    `json:"file_type"`
	FileSize     int64     `json:"file_size"`
	SHA256       string    `json:"sha256"`
	StoragePath  string    `json:"-"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Version) TableName() string { return "archive_mapping_versions" }

type Link struct {
	KnowledgeID string    `json:"knowledge_id" gorm:"primaryKey;size:36"`
	TenantID    uint64    `json:"-"`
	VersionID   string    `json:"version_id" gorm:"size:36"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Link) TableName() string { return "archive_mapping_links" }

type ManifestFile struct {
	RelativePath string `json:"relative_path"`
	FileSize     int64  `json:"file_size"`
}

// NormalizePath rejects traversal, absolute paths and invisible control bytes.
// Root removal is explicit, never a fuzzy suffix or filename-only fallback.
func NormalizePath(raw string, allowEmpty bool) (string, error) {
	s := norm.NFC.String(strings.ReplaceAll(raw, "\\", "/"))
	if s == "" && allowEmpty {
		return "", nil
	}
	if s == "" || len(s) > 4096 || strings.HasPrefix(s, "/") || strings.Contains(s, ":") {
		return "", fmt.Errorf("必须使用有效的相对路径")
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("路径包含控制字符")
		}
	}
	for _, p := range strings.Split(s, "/") {
		if p == "" || p == "." || p == ".." {
			return "", fmt.Errorf("路径不允许空段或 . / ..")
		}
	}
	return s, nil
}
func MatchKey(relative string) string { return strings.TrimSuffix(relative, path.Ext(relative)) }
func UnderRoot(relative, root string) (string, bool) {
	if root == "" {
		return relative, true
	}
	prefix := root + "/"
	if !strings.HasPrefix(relative, prefix) {
		return "", false
	}
	return strings.TrimPrefix(relative, prefix), true
}
func AllowedOriginal(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".pdf", ".png", ".jpg", ".jpeg", ".webp", ".gif", ".bmp", ".tif", ".tiff", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".odt", ".ods", ".odp", ".rtf":
		return true
	}
	return false
}

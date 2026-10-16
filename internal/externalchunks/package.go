// Package externalchunks defines the versioned, parser-independent import boundary.
package externalchunks

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"unicode/utf8"
)

const Format = "weknora-chunks-v1"

type Entry struct {
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type Manifest struct {
	Format           string           `json:"format"`
	PackageID        string           `json:"package_id"`
	SourceSystem     string           `json:"source_system"`
	DatasetID        string           `json:"dataset_id"`
	GenerationID     string           `json:"generation_id"`
	ProducerCommit   string           `json:"producer_commit"`
	GenerationConfig json.RawMessage  `json:"generation_config"`
	Counts           map[string]int   `json:"counts"`
	Files            map[string]Entry `json:"files"`
}
type Document struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	RelativePath string `json:"relative_path"`
	Version      int    `json:"version"`
	SourceSHA256 string `json:"source_sha256"`
	ParsedFile   string `json:"parsed_file"`
	OriginalFile string `json:"original_file"`
}
type Chunk struct {
	ID            string           `json:"id"`
	DocumentID    string           `json:"document_id"`
	Order         int              `json:"order"`
	Content       string           `json:"content"`
	RetrievalText string           `json:"retrieval_text"`
	ContextText   string           `json:"context_text"`
	TokenCount    int              `json:"token_count"`
	Metadata      map[string]any   `json:"metadata"`
	SourceAnchors []map[string]any `json:"source_anchors"`
}
type SourceUnit struct {
	ID          string          `json:"id"`
	DocumentID  string          `json:"document_id"`
	UnitID      string          `json:"unit_id"`
	PageIndex   int             `json:"page_index"`
	Order       int             `json:"order"`
	BlockType   string          `json:"block_type"`
	Text        string          `json:"text"`
	BoundingBox map[string]any  `json:"bounding_box"`
	TitlePath   []string        `json:"title_path"`
	Table       json.RawMessage `json:"table"`
}
type Package struct {
	Manifest     Manifest     `json:"manifest"`
	ManifestHash string       `json:"manifest_hash"`
	Documents    []Document   `json:"documents"`
	Chunks       []Chunk      `json:"chunks"`
	SourceUnits  []SourceUnit `json:"source_units"`
}

func SafePath(name string) bool {
	if name == "" || !utf8.ValidString(name) || strings.ContainsAny(name, "\\:") || strings.HasPrefix(name, "/") {
		return false
	}
	for _, c := range name {
		if c < 32 || c == 127 {
			return false
		}
	}
	for _, part := range strings.Split(name, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
func hashOK(s string) bool {
	b, e := hex.DecodeString(s)
	return e == nil && len(b) == 32 && strings.ToLower(s) == s
}
func readJSONLines[T any](f *zip.File) ([]T, error) {
	r, e := f.Open()
	if e != nil {
		return nil, e
	}
	defer r.Close()
	scan := bufio.NewScanner(r)
	scan.Buffer(make([]byte, 64*1024), 64*1024*1024)
	values := []T{}
	for scan.Scan() {
		var value T
		if !utf8.Valid(scan.Bytes()) {
			return nil, fmt.Errorf("%s: invalid UTF-8", f.Name)
		}
		if e = json.Unmarshal(scan.Bytes(), &value); e != nil {
			return nil, fmt.Errorf("%s line %d: %w", f.Name, len(values)+1, e)
		}
		values = append(values, value)
		if len(values) > 200000 {
			return nil, fmt.Errorf("too many rows")
		}
	}
	return values, scan.Err()
}

// Read validates every raw member before making any database or storage mutation.
// Limits are explicit safety limits, not the ordinary document parser's file cap.
func Read(filename string) (*Package, error) {
	z, e := zip.OpenReader(filename)
	if e != nil {
		return nil, e
	}
	defer z.Close()
	if len(z.File) > 100000 {
		return nil, fmt.Errorf("ZIP exceeds 100000 members")
	}
	files := map[string]*zip.File{}
	var total uint64
	for _, f := range z.File {
		if !SafePath(f.Name) || files[f.Name] != nil || f.Mode()&os.ModeSymlink != 0 || !f.Mode().IsRegular() || f.Flags&1 != 0 {
			return nil, fmt.Errorf("unsafe/duplicate ZIP member: %s", f.Name)
		}
		if f.FileInfo().IsDir() || f.Mode()&0111 != 0 {
			return nil, fmt.Errorf("directories/executable members not accepted: %s", f.Name)
		}
		total += f.UncompressedSize64
		if total > 32<<30 || f.UncompressedSize64 > 32<<30 {
			return nil, fmt.Errorf("expanded ZIP exceeds 32 GiB")
		}
		if f.UncompressedSize64 > 1<<20 && f.UncompressedSize64/max(f.CompressedSize64, 1) > 1000 {
			return nil, fmt.Errorf("unsafe expansion ratio: %s", f.Name)
		}
		files[f.Name] = f
	}
	mf := files["manifest.json"]
	if mf == nil || mf.UncompressedSize64 > 8<<20 {
		return nil, fmt.Errorf("missing/oversized manifest")
	}
	mr, e := mf.Open()
	if e != nil {
		return nil, e
	}
	raw, e := io.ReadAll(io.LimitReader(mr, 8<<20))
	mr.Close()
	if e != nil {
		return nil, e
	}
	p := &Package{}
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("manifest is not UTF-8")
	}
	if e = json.Unmarshal(raw, &p.Manifest); e != nil {
		return nil, e
	}
	m := p.Manifest
	if m.Format != Format || m.PackageID == "" || m.SourceSystem == "" || m.DatasetID == "" || m.GenerationID == "" {
		return nil, fmt.Errorf("invalid format or source identity")
	}
	if len(files) != len(m.Files)+1 {
		return nil, fmt.Errorf("undeclared/missing ZIP members")
	}
	sum := sha256.Sum256(raw)
	p.ManifestHash = hex.EncodeToString(sum[:])
	for name, entry := range m.Files {
		f := files[name]
		if f == nil || name == "manifest.json" || !hashOK(entry.SHA256) || entry.Size < 0 || uint64(entry.Size) != f.UncompressedSize64 {
			return nil, fmt.Errorf("invalid member declaration: %s", name)
		}
		if name != "documents.jsonl" && name != "chunks.jsonl" && name != "source_units.jsonl" && !strings.HasPrefix(name, "files/") {
			return nil, fmt.Errorf("invalid attachment location: %s", name)
		}
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		h := sha256.New()
		_, err = io.Copy(h, r)
		r.Close()
		if err != nil || hex.EncodeToString(h.Sum(nil)) != entry.SHA256 {
			return nil, fmt.Errorf("checksum mismatch: %s", name)
		}
	}
	for _, name := range []string{"documents.jsonl", "chunks.jsonl", "source_units.jsonl"} {
		if files[name] == nil || files[name].UncompressedSize64 > 512<<20 {
			return nil, fmt.Errorf("missing/oversized %s", name)
		}
	}
	if p.Documents, e = readJSONLines[Document](files["documents.jsonl"]); e != nil {
		return nil, e
	}
	if p.Chunks, e = readJSONLines[Chunk](files["chunks.jsonl"]); e != nil {
		return nil, e
	}
	if p.SourceUnits, e = readJSONLines[SourceUnit](files["source_units.jsonl"]); e != nil {
		return nil, e
	}
	if len(p.Documents) == 0 || len(p.Documents) != m.Counts["documents"] || len(p.Chunks) != m.Counts["chunks"] || len(p.SourceUnits) != m.Counts["source_units"] {
		return nil, fmt.Errorf("invalid record counts")
	}
	docs := map[string]Document{}
	paths := map[string]bool{}
	orders := map[string]int{}
	for _, d := range p.Documents {
		if d.ID == "" || docs[d.ID].ID != "" || !SafePath(d.RelativePath) || paths[d.RelativePath] || strings.ToLower(path.Ext(d.RelativePath)) != ".md" || d.Version < 1 || !hashOK(d.SourceSHA256) {
			return nil, fmt.Errorf("invalid document: %s", d.ID)
		}
		if files[d.ParsedFile] == nil || !strings.HasPrefix(d.ParsedFile, "files/") {
			return nil, fmt.Errorf("missing parsed file: %s", d.ID)
		}
		if d.OriginalFile != "" {
			entry, ok := m.Files[d.OriginalFile]
			if !ok || !strings.HasPrefix(d.OriginalFile, "files/") || entry.SHA256 != d.SourceSHA256 {
				return nil, fmt.Errorf("original hash mismatch: %s", d.ID)
			}
		}
		docs[d.ID] = d
		paths[d.RelativePath] = true
	}
	unitIDs := map[string]bool{}
	unitKeys := map[string]bool{}
	for _, u := range p.SourceUnits {
		key := u.DocumentID + "\x00" + u.UnitID
		if u.ID == "" || unitIDs[u.ID] || u.UnitID == "" || unitKeys[key] || docs[u.DocumentID].ID == "" || u.PageIndex < 0 || u.Order < 0 {
			return nil, fmt.Errorf("invalid source unit: %s", u.ID)
		}
		unitIDs[u.ID] = true
		unitKeys[key] = true
	}
	chunks := map[string]bool{}
	for _, c := range p.Chunks {
		if c.ID == "" || chunks[c.ID] || docs[c.DocumentID].ID == "" || c.Order != orders[c.DocumentID] || strings.TrimSpace(c.Content) == "" || c.TokenCount < 1 {
			return nil, fmt.Errorf("invalid chunk: %s", c.ID)
		}
		chunks[c.ID] = true
		orders[c.DocumentID]++
		for _, a := range c.SourceAnchors {
			id, ok := a["unit_id"].(string)
			if !ok || !unitKeys[c.DocumentID+"\x00"+id] {
				return nil, fmt.Errorf("unknown source anchor: %s", c.ID)
			}
		}
	}
	for id := range docs {
		if orders[id] == 0 {
			return nil, fmt.Errorf("document has no chunks: %s", id)
		}
	}
	return p, nil
}

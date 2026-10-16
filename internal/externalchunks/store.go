package externalchunks

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Batch struct {
	ID              string    `json:"id" gorm:"primaryKey;size:36"`
	TenantID        uint64    `json:"-"`
	KnowledgeBaseID string    `json:"knowledge_base_id"`
	ManifestHash    string    `json:"manifest_hash"`
	ManifestJSON    string    `json:"-" gorm:"type:text"`
	PackagePath     string    `json:"-" gorm:"type:text"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

func (Batch) TableName() string { return "external_chunk_batches" }

type Record struct {
	ID                 string    `json:"id" gorm:"primaryKey;size:64"`
	TenantID           uint64    `json:"-"`
	KnowledgeBaseID    string    `json:"knowledge_base_id"`
	BatchID            string    `json:"batch_id"`
	ExternalDocumentID string    `json:"external_document_id"`
	KnowledgeID        string    `json:"knowledge_id" gorm:"size:36"`
	RelativePath       string    `json:"relative_path"`
	ContentHash        string    `json:"content_hash"`
	Payload            string    `json:"-" gorm:"type:text"`
	Status             string    `json:"status"`
	Stage              string    `json:"stage"`
	Error              string    `json:"error" gorm:"type:text"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (Record) TableName() string { return "external_chunk_records" }

type Membership struct {
	BatchID  string `gorm:"primaryKey;size:36"`
	RecordID string `gorm:"primaryKey;size:64"`
}

func (Membership) TableName() string { return "external_chunk_memberships" }

type DocumentPayload struct {
	Document    Document     `json:"document"`
	Chunks      []Chunk      `json:"chunks"`
	SourceUnits []SourceUnit `json:"source_units"`
}
type Store struct{ DB *gorm.DB }

func NewRuntimeStore(db *gorm.DB) (*Store, error) {
	for _, table := range []any{&Batch{}, &Record{}, &Membership{}} {
		if !db.Migrator().HasTable(table) {
			return nil, fmt.Errorf("external chunk migration missing")
		}
	}
	return &Store{DB: db}, nil
}
func Identity(tenant uint64, kb string, m Manifest, old string) string {
	raw, _ := json.Marshal([]any{tenant, kb, m.SourceSystem, m.DatasetID, m.GenerationID, old})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func MappedID(identity, old string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(identity+"\x00"+old)).String()
}

// Plan atomically reserves source identities. Same identity/different content is
// a conflict, never an update. A repeated package returns the original batch.
func (s *Store) Plan(ctx context.Context, tenant uint64, kb, filename string, p *Package) (*Batch, []Record, error) {
	b := &Batch{ID: MappedID(fmt.Sprint(tenant)+":"+kb, p.ManifestHash), TenantID: tenant, KnowledgeBaseID: kb, ManifestHash: p.ManifestHash, PackagePath: filename, Status: "preview", CreatedAt: time.Now().UTC()}
	raw, _ := json.Marshal(p.Manifest)
	b.ManifestJSON = string(raw)
	var result []Record
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(b).Error; e != nil {
			return e
		}
		if e := tx.Where("id = ? AND tenant_id = ? AND knowledge_base_id = ?", b.ID, tenant, kb).First(b).Error; e != nil {
			return e
		}
		for _, d := range p.Documents {
			payload := DocumentPayload{Document: d, Chunks: []Chunk{}, SourceUnits: []SourceUnit{}}
			for _, c := range p.Chunks {
				if c.DocumentID == d.ID {
					payload.Chunks = append(payload.Chunks, c)
				}
			}
			for _, u := range p.SourceUnits {
				if u.DocumentID == d.ID {
					payload.SourceUnits = append(payload.SourceUnits, u)
				}
			}
			raw, _ := json.Marshal(payload)
			digest := sha256.Sum256(raw)
			id := Identity(tenant, kb, p.Manifest, d.ID)
			record := Record{ID: id, TenantID: tenant, KnowledgeBaseID: kb, BatchID: b.ID, ExternalDocumentID: d.ID, KnowledgeID: MappedID(id, "knowledge"), RelativePath: d.RelativePath, ContentHash: hex.EncodeToString(digest[:]), Payload: string(raw), Status: "preview", Stage: "import", UpdatedAt: time.Now().UTC()}
			if e := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error; e != nil {
				return e
			}
			var actual Record
			if e := tx.Where("id = ? AND tenant_id = ? AND knowledge_base_id = ?", id, tenant, kb).First(&actual).Error; e != nil {
				return e
			}
			if actual.ContentHash != record.ContentHash {
				return fmt.Errorf("source identity conflict: document %s; use a new generation, existing knowledge %s is unchanged", d.ID, actual.KnowledgeID)
			}
			link := Membership{BatchID: b.ID, RecordID: actual.ID}
			if e := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; e != nil {
				return e
			}
			result = append(result, actual)
		}
		return nil
	})
	return b, result, err
}
func (s *Store) Batch(ctx context.Context, tenant uint64, kb, id string) (*Batch, []Record, error) {
	var b Batch
	var rows []Record
	if e := s.DB.WithContext(ctx).Where("id = ? AND tenant_id = ? AND knowledge_base_id = ?", id, tenant, kb).First(&b).Error; e != nil {
		return nil, nil, e
	}
	e := s.DB.WithContext(ctx).Table("external_chunk_records AS r").Select("r.*").Joins("JOIN external_chunk_memberships AS m ON m.record_id = r.id").Where("m.batch_id = ? AND r.tenant_id = ? AND r.knowledge_base_id = ?", id, tenant, kb).Order("r.relative_path").Scan(&rows).Error
	return &b, rows, e
}
func (s *Store) List(ctx context.Context, tenant uint64, kb string) ([]Batch, error) {
	var rows []Batch
	e := s.DB.WithContext(ctx).Where("tenant_id = ? AND knowledge_base_id = ?", tenant, kb).Order("created_at DESC").Limit(100).Find(&rows).Error
	return rows, e
}
func (s *Store) ForKnowledge(ctx context.Context, tenant uint64, id string) (*Record, error) {
	var r Record
	e := s.DB.WithContext(ctx).Where("tenant_id = ? AND knowledge_id = ?", tenant, id).First(&r).Error
	return &r, e
}
func (s *Store) Update(ctx context.Context, r *Record, status, stage, reason string) error {
	return s.DB.WithContext(ctx).Model(&Record{}).Where("id = ? AND tenant_id = ? AND knowledge_base_id = ?", r.ID, r.TenantID, r.KnowledgeBaseID).Updates(map[string]any{"status": status, "stage": stage, "error": reason, "updated_at": time.Now().UTC()}).Error
}

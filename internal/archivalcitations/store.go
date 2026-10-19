package archivalcitations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrImmutable = errors.New("citation evidence is already finalized for another answer revision")

type Record struct {
	ID                string `gorm:"primaryKey;size:64"`
	TenantID          uint64 `gorm:"index"`
	SessionID         string
	MessageID         string `gorm:"index"`
	KnowledgeTenantID uint64
	KnowledgeID       string
	ChunkID           string
	SummaryJSON       string
	Snapshot          string
	CreatedAt         time.Time
}

func (Record) TableName() string { return "message_citation_evidence" }

type ManifestRecord struct {
	TenantID     uint64 `gorm:"primaryKey;autoIncrement:false"`
	SessionID    string `gorm:"primaryKey;size:36"`
	MessageID    string `gorm:"primaryKey;size:36"`
	AnswerSHA256 string
	Payload      string
	CreatedAt    time.Time
}

func (ManifestRecord) TableName() string { return "message_citation_manifests" }

type Store struct{ DB *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{DB: db} }

// Save publishes the complete, validated manifest and immutable snapshots together.
func (s *Store) Save(ctx context.Context, tenant uint64, session, message, answer string, m Manifest, records []Record) error {
	if tenant == 0 || session == "" || message == "" || m.Version != SchemaVersion || len(records) != len(m.Citations) || len(records) > 256 {
		return fmt.Errorf("invalid citation scope/count")
	}
	if m.State != "verified" && m.State != "needs_review" && m.State != "legacy" {
		return fmt.Errorf("invalid citation state")
	}
	seen := map[string]bool{}
	for i := range records {
		r := &records[i]
		summary := m.Citations[i]
		if r.TenantID != tenant || r.SessionID != session || r.MessageID != message || r.KnowledgeTenantID == 0 || r.KnowledgeID != summary.KnowledgeID || r.ChunkID != summary.ChunkID || r.ID != summary.ID || r.ID != EvidenceID(tenant, session, message, r.ChunkID, r.Snapshot, summary.ArchiveVersionID, summary.ContentRevision) || Hash(r.Snapshot) != summary.ContentSHA256 || summary.Number != i+1 || seen[r.ID] || len(r.Snapshot) > 4<<20 {
			return fmt.Errorf("invalid citation evidence")
		}
		if e := ValidateLocation(summary.Location); e != nil {
			return e
		}
		seen[r.ID] = true
		raw, e := json.Marshal(summary)
		if e != nil {
			return e
		}
		r.SummaryJSON = string(raw)
		r.CreatedAt = time.Now().UTC()
	}
	raw, e := json.Marshal(m)
	if e != nil {
		return e
	}
	row := ManifestRecord{TenantID: tenant, SessionID: session, MessageID: message, AnswerSHA256: Hash(answer), Payload: string(raw), CreatedAt: time.Now().UTC()}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		created := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
		if created.Error != nil {
			return created.Error
		}
		if created.RowsAffected == 0 {
			var existing ManifestRecord
			if e := tx.Where("tenant_id=? AND session_id=? AND message_id=?", tenant, session, message).First(&existing).Error; e != nil {
				return e
			}
			if existing.AnswerSHA256 != row.AnswerSHA256 {
				return ErrImmutable
			}
			return nil
		}
		if len(records) > 0 {
			return tx.Create(&records).Error
		}
		return nil
	})
}
func (s *Store) Manifest(ctx context.Context, tenant uint64, session, message string, answer ...string) (Manifest, error) {
	var row ManifestRecord
	var m Manifest
	e := s.DB.WithContext(ctx).Where("tenant_id=? AND session_id=? AND message_id=?", tenant, session, message).First(&row).Error
	if e == nil {
		e = json.Unmarshal([]byte(row.Payload), &m)
	}
	if e == nil && len(answer) > 0 && Hash(answer[0]) != row.AnswerSHA256 {
		m.State = "answer_changed"
		m.Citations = []Summary{}
		m.Unbound = []UnboundBlock{}
	}
	return m, e
}
func (s *Store) Evidence(ctx context.Context, tenant uint64, session, message, id string) (Record, Summary, error) {
	var row Record
	var summary Summary
	e := s.DB.WithContext(ctx).Where("tenant_id=? AND session_id=? AND message_id=? AND id=?", tenant, session, message, id).First(&row).Error
	if e == nil {
		e = json.Unmarshal([]byte(row.SummaryJSON), &summary)
	}
	return row, summary, e
}

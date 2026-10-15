package archivemapping

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct{ DB *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{DB: db} }
func pathBase(p string) string    { return path.Base(p) }
func extension(p string) string   { return strings.TrimPrefix(strings.ToLower(path.Ext(p)), ".") }

func (s *Store) Plan(ctx context.Context, tenant uint64, kb, root, user string, files []ManifestFile) (*Batch, []Item, error) {
	root, err := NormalizePath(root, true)
	if err != nil {
		return nil, nil, err
	}
	if len(files) == 0 || len(files) > 10000 {
		return nil, nil, fmt.Errorf("每批必须包含 1～10000 个原件")
	}
	var knowledge []types.Knowledge
	if err = s.DB.WithContext(ctx).Where("tenant_id = ? AND knowledge_base_id = ? AND file_type IN ?", tenant, kb, []string{"md", "markdown"}).Find(&knowledge).Error; err != nil {
		return nil, nil, err
	}
	// Scoped by the selected KB/tenant. Same names in other folders are distinct.
	index := map[string][]Candidate{}
	for _, k := range knowledge {
		p := k.FileName
		if k.FolderPath != "" {
			p = k.FolderPath + "/" + p
		}
		p, e := NormalizePath(p, false)
		if e != nil {
			continue
		}
		rel, ok := UnderRoot(p, root)
		if !ok {
			continue
		}
		var count int64
		if e = s.DB.WithContext(ctx).Model(&Link{}).Where("tenant_id = ? AND knowledge_id = ?", tenant, k.ID).Count(&count).Error; e != nil {
			return nil, nil, e
		}
		key := MatchKey(rel)
		index[key] = append(index[key], Candidate{ID: k.ID, Title: k.Title, Path: p, Mapped: count > 0})
	}
	b := &Batch{ID: uuid.NewString(), TenantID: tenant, KnowledgeBaseID: kb, ParsedRoot: root, CreatedBy: user, CreatedAt: time.Now().UTC()}
	items := make([]Item, 0, len(files))
	seen := map[string]bool{}
	for _, f := range files {
		p, e := NormalizePath(f.RelativePath, false)
		if e != nil {
			return nil, nil, e
		}
		if seen[p] {
			return nil, nil, fmt.Errorf("清单相对路径重复: %s", p)
		}
		seen[p] = true
		if f.FileSize < 0 {
			return nil, nil, fmt.Errorf("文件大小无效: %s", p)
		}
		it := Item{ID: uuid.NewString(), BatchID: b.ID, RelativePath: p, FileSize: f.FileSize, Candidates: index[MatchKey(p)], Status: "unmatched", UpdatedAt: b.CreatedAt}
		if !AllowedOriginal(p) {
			it.Status = "failed"
			it.Error = "不支持的原件格式"
		} else if len(it.Candidates) > 1 {
			it.Status = "ambiguous"
			it.Error = "相同相对路径匹配多个 MD，请选择目标"
		} else if len(it.Candidates) == 1 {
			it.KnowledgeID = it.Candidates[0].ID
			it.Status = "ready"
			if it.Candidates[0].Mapped {
				it.Status = "conflict"
				it.Error = "已有原件，请明确选择替换或跳过"
			}
		} else {
			it.Error = "相对路径未匹配，请确认 MD 根目录或手动选择目标"
		}
		items = append(items, it)
	}
	err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(b).Error; e != nil {
			return e
		}
		return tx.CreateInBatches(&items, 200).Error
	})
	return b, items, err
}
func (s *Store) Batch(ctx context.Context, tenant uint64, kb, id string) (*Batch, []Item, error) {
	var b Batch
	var items []Item
	err := s.DB.WithContext(ctx).Where("id = ? AND tenant_id = ? AND knowledge_base_id = ?", id, tenant, kb).First(&b).Error
	if err == nil {
		err = s.DB.WithContext(ctx).Where("batch_id = ?", id).Order("relative_path").Find(&items).Error
	}
	return &b, items, err
}
func (s *Store) Batches(ctx context.Context, tenant uint64, kb string) ([]Batch, error) {
	var b []Batch
	err := s.DB.WithContext(ctx).Where("tenant_id = ? AND knowledge_base_id = ?", tenant, kb).Order("created_at DESC").Limit(30).Find(&b).Error
	return b, err
}
func (s *Store) Active(ctx context.Context, tenant uint64, id string) (*Version, error) {
	var l Link
	err := s.DB.WithContext(ctx).Where("knowledge_id = ? AND tenant_id = ?", id, tenant).First(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var v Version
	err = s.DB.WithContext(ctx).Where("id = ? AND knowledge_id = ? AND tenant_id = ?", l.VersionID, id, tenant).First(&v).Error
	return &v, err
}
func (s *Store) Versions(ctx context.Context, tenant uint64, id string) ([]Version, error) {
	var v []Version
	err := s.DB.WithContext(ctx).Where("tenant_id = ? AND knowledge_id = ?", tenant, id).Order("created_at DESC").Find(&v).Error
	return v, err
}
func (s *Store) Activate(ctx context.Context, tenant uint64, kb, id, version string) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if e := lockKnowledge(tx, tenant, kb, id); e != nil {
			return e
		}
		if version == "" {
			return tx.Where("tenant_id = ? AND knowledge_id = ?", tenant, id).Delete(&Link{}).Error
		}
		var v Version
		if e := tx.Where("id = ? AND tenant_id = ? AND knowledge_id = ?", version, tenant, id).First(&v).Error; e != nil {
			return e
		}
		return setLink(tx, tenant, id, version)
	})
}
func lockKnowledge(tx *gorm.DB, tenant uint64, kb, id string) error {
	var k types.Knowledge
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND tenant_id = ? AND knowledge_base_id = ?", id, tenant, kb).First(&k).Error
}
func setLink(tx *gorm.DB, tenant uint64, id, version string) error {
	l := Link{KnowledgeID: id, TenantID: tenant, VersionID: version, UpdatedAt: time.Now().UTC()}
	return tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "knowledge_id"}}, DoUpdates: clause.AssignmentColumns([]string{"tenant_id", "version_id", "updated_at"})}).Create(&l).Error
}

// Bind holds the knowledge lock through storage and metadata publication. A
// concurrent replace cannot silently overwrite another upload. Save is called
// only after all scope/conflict/idempotency checks pass.
func (s *Store) Bind(ctx context.Context, tenant uint64, kb, batch, item, target, user, hash string, size int64, replace bool, save func(string) (string, error)) (*Version, error) {
	var result *Version
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var b Batch
		if e := tx.Where("id = ? AND tenant_id = ? AND knowledge_base_id = ?", batch, tenant, kb).First(&b).Error; e != nil {
			return e
		}
		var it Item
		if e := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND batch_id = ?", item, batch).First(&it).Error; e != nil {
			return e
		}
		if it.Status == "bound" {
			var v Version
			e := tx.Where("id = ? AND tenant_id = ?", it.VersionID, tenant).First(&v).Error
			result = &v
			return e
		}
		if !AllowedOriginal(it.RelativePath) {
			return fmt.Errorf("不支持的原件格式")
		}
		if target == "" {
			target = it.KnowledgeID
		}
		if target == "" {
			return fmt.Errorf("请先选择目标 MD")
		}
		if e := lockKnowledge(tx, tenant, kb, target); e != nil {
			return e
		}
		var k types.Knowledge
		if e := tx.Where("id = ?", target).First(&k).Error; e != nil {
			return e
		}
		if k.FileType != "md" && k.FileType != "markdown" {
			return fmt.Errorf("目标必须是当前知识库的 MD 文档")
		}
		if size != it.FileSize {
			return fmt.Errorf("文件大小与清单不一致，请重新选择原文件")
		}
		current, e := (&Store{DB: tx}).Active(ctx, tenant, target)
		if e != nil {
			return e
		}
		if current != nil && current.SHA256 == hash {
			result = current
		} else {
			if current != nil && !replace {
				return fmt.Errorf("目标已有不同原件，请明确选择替换")
			}
			v := Version{ID: uuid.NewString(), TenantID: tenant, KnowledgeID: target, FileName: pathBase(it.RelativePath), RelativePath: it.RelativePath, FileType: extension(it.RelativePath), FileSize: size, SHA256: hash, CreatedBy: user, CreatedAt: time.Now().UTC()}
			v.StoragePath, e = save(v.ID)
			if e != nil {
				return e
			}
			if e = tx.Create(&v).Error; e != nil {
				return e
			}
			result = &v
		}
		if e = setLink(tx, tenant, target, result.ID); e != nil {
			return e
		}
		return tx.Model(&Item{}).Where("id = ?", item).Updates(map[string]any{"knowledge_id": target, "version_id": result.ID, "status": "bound", "error": "", "updated_at": time.Now().UTC()}).Error
	})
	return result, err
}
func (s *Store) ItemState(ctx context.Context, tenant uint64, kb, batch, item, status, reason string) error {
	_, _, err := s.Batch(ctx, tenant, kb, batch)
	if err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Model(&Item{}).Where("id = ? AND batch_id = ? AND status <> ?", item, batch, "bound").Updates(map[string]any{"status": status, "error": reason, "updated_at": time.Now().UTC()}).Error
}

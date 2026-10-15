package archivemapping

import (
	"context"
	"fmt"
	"github.com/Tencent/WeKnora/internal/types"
	"gorm.io/gorm"
)

func adjustStorage(tx *gorm.DB, tenant uint64, delta int64) error {
	// A single conditional update serializes quota reservation with other uploads.
	q := tx.Model(&types.Tenant{}).Where("id = ?", tenant)
	if delta > 0 {
		q = q.Where("storage_quota <= 0 OR storage_used + ? <= storage_quota", delta)
	}
	result := q.Update("storage_used", gorm.Expr("CASE WHEN storage_used + ? < 0 THEN 0 ELSE storage_used + ? END", delta, delta))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return fmt.Errorf("存储额度不足或租户不存在")
	}
	return nil
}
func (s *Store) RemoveVersion(ctx context.Context, tenant uint64, id, version string) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var v Version
		if e := tx.Where("id = ? AND tenant_id = ? AND knowledge_id = ?", version, tenant, id).First(&v).Error; e != nil {
			return e
		}
		if e := tx.Where("tenant_id = ? AND knowledge_id = ? AND version_id = ?", tenant, id, version).Delete(&Link{}).Error; e != nil {
			return e
		}
		if e := tx.Delete(&v).Error; e != nil {
			return e
		}
		return adjustStorage(tx, tenant, -v.FileSize)
	})
}
func (s *Store) AddCopies(ctx context.Context, versions []Version, active string) error {
	if len(versions) == 0 {
		return nil
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var bytes int64
		for _, v := range versions {
			bytes += v.FileSize
		}
		if e := adjustStorage(tx, versions[0].TenantID, bytes); e != nil {
			return e
		}
		if e := tx.Create(&versions).Error; e != nil {
			return e
		}
		if active != "" {
			return setLink(tx, versions[0].TenantID, versions[0].KnowledgeID, active)
		}
		return nil
	})
}
func (s *Store) Targets(ctx context.Context, tenant uint64, kb string) ([]Candidate, error) {
	var ks []types.Knowledge
	err := s.DB.WithContext(ctx).Where("tenant_id = ? AND knowledge_base_id = ? AND file_type IN ?", tenant, kb, []string{"md", "markdown"}).Order("folder_path, file_name").Find(&ks).Error
	result := make([]Candidate, 0, len(ks))
	for _, k := range ks {
		p := k.FileName
		if k.FolderPath != "" {
			p = k.FolderPath + "/" + p
		}
		result = append(result, Candidate{ID: k.ID, Title: k.Title, Path: p})
	}
	return result, err
}

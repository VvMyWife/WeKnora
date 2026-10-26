package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/processingcontrol"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const EnrichmentRetryCooldown = 30 * time.Minute
const EnrichmentRecoveryCooldown = 2 * time.Hour
const recoveryBackfillKB = "8575ca20-5e7e-48c4-9fc0-786d979fdf5d"

// One durable outbox per logical enrichment, not per failed delivery. A replay
// has MaxRetry(0); failures update this row instead of creating another budget.
type EnrichmentRecoveryRow struct {
	ID                 int64
	SourceDeadLetterID int64
	LogicalKey         string
	TenantID           uint64
	KnowledgeBaseID    string
	KnowledgeID        string
	Attempt            int
	TaskType           string
	Payload            json.RawMessage `gorm:"type:jsonb"`
	State              string
	Attempts           int
	DispatchNo         int
	AvailableAt        time.Time
	LeaseUntil         *time.Time
	QueuedAt           *time.Time
	SlotReserved       bool
	LastError          string
	UpdatedAt          time.Time
}

func (EnrichmentRecoveryRow) TableName() string { return "enrichment_recoveries" }

type EnrichmentRecovery struct {
	db        *gorm.DB
	tasks     interfaces.TaskEnqueuer
	inspector *asynq.Inspector
}

var enrichmentRecovery atomic.Pointer[EnrichmentRecovery]

func NewEnrichmentRecovery(db *gorm.DB, tasks interfaces.TaskEnqueuer, inspector *asynq.Inspector) *EnrichmentRecovery {
	return &EnrichmentRecovery{db: db, tasks: tasks, inspector: inspector}
}

// Run starts only after the schema migration and after pause admission is installed.
func (s *EnrichmentRecovery) Run(ctx context.Context) {
	enrichmentRecovery.Store(s)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		runCtx, cancel := context.WithTimeout(ctx, 50*time.Second)
		if err := s.sweepSafely(runCtx); err != nil {
			logger.Warnf(ctx, "[EnrichmentRecovery] sweep: %v", err)
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *EnrichmentRecovery) Sweep(ctx context.Context) error {
	if err := s.discover(ctx); err != nil {
		return err
	}
	var rows []EnrichmentRecoveryRow
	if err := s.db.WithContext(ctx).Where("state='queued' OR (state='ready' AND available_at<=?)", time.Now()).Order("available_at,id").Limit(200).Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if err := s.dispatch(ctx, row.ID); err != nil {
			logger.Warnf(ctx, "[EnrichmentRecovery] dispatch id=%d: %v", row.ID, err)
		}
	}
	return nil
}

func latestEnrichmentAttempt(tx *gorm.DB, kid string) (int, error) {
	var attempt int
	err := tx.Raw("SELECT COALESCE(MAX(attempt),0) FROM knowledge_processing_spans WHERE knowledge_id=?", kid).Scan(&attempt).Error
	return attempt, err
}

func recoveryAlreadySucceeded(tx *gorm.DB, row *EnrichmentRecoveryRow) (bool, error) {
	var statuses []string
	name := "postprocess.wiki"
	if row.TaskType == types.TypeSummaryGeneration {
		var status string
		err := tx.Raw("SELECT summary_status FROM knowledges WHERE id=?", row.KnowledgeID).Scan(&status).Error
		return status == types.SummaryStatusCompleted, err
	}
	if row.TaskType == types.TypeChunkExtract {
		var p types.ExtractChunkPayload
		if err := json.Unmarshal(row.Payload, &p); err != nil {
			return false, err
		}
		name = fmt.Sprintf("postprocess.graph.chunk[%d]", p.ChunkIndex)
	}
	err := tx.Raw("SELECT status FROM knowledge_processing_spans WHERE knowledge_id=? AND attempt=? AND name=? ORDER BY updated_at DESC,id DESC LIMIT 1", row.KnowledgeID, row.Attempt, name).Scan(&statuses).Error
	return len(statuses) > 0 && statuses[0] == types.SpanStatusDone, err
}

func (s *EnrichmentRecovery) discover(ctx context.Context) error {
	var letters []types.TaskDeadLetter
	// The migration records an immutable deployment boundary. Only the named KB
	// is backfilled; later failures in all KBs are eligible. Seen IDs are bounded
	// by the logical-key unique constraint, including duplicate dead-letter writes.
	err := s.db.WithContext(ctx).Raw(`SELECT d.* FROM task_dead_letters d
 JOIN knowledges k ON k.id=CASE WHEN d.scope='knowledge' THEN d.scope_id ELSE d.related_id END
 WHERE d.task_type IN ('chunk:extract','wiki:ingest') AND k.deleted_at IS NULL
 AND (k.knowledge_base_id=? OR d.failed_at >= (SELECT enabled_at FROM enrichment_retry_settings WHERE id=1))
 AND NOT EXISTS(SELECT 1 FROM enrichment_recovery_sources r WHERE r.dead_letter_id=d.id)
 ORDER BY d.id LIMIT 500`, recoveryBackfillKB).Scan(&letters).Error
	if err != nil {
		return err
	}
	for _, dl := range letters {
		kid := dl.RelatedID
		if dl.Scope == types.TaskScopeKnowledge {
			kid = dl.ScopeID
		}
		var k types.Knowledge
		if err := s.db.WithContext(ctx).Where("id=?", kid).First(&k).Error; err != nil {
			continue
		}
		attempt, err := latestEnrichmentAttempt(s.db.WithContext(ctx), kid)
		if err != nil {
			return err
		}
		row := EnrichmentRecoveryRow{SourceDeadLetterID: dl.ID, TenantID: dl.TenantID, KnowledgeBaseID: k.KnowledgeBaseID, KnowledgeID: kid, Attempt: attempt, TaskType: dl.TaskType, Payload: dl.Payload, State: "ready", AvailableAt: time.Now().Add(EnrichmentRecoveryCooldown), LastError: dl.LastError, UpdatedAt: time.Now()}
		if dl.TaskType == types.TypeChunkExtract {
			var p types.ExtractChunkPayload
			if err := json.Unmarshal(dl.Payload, &p); err != nil || p.ChunkID == "" {
				continue
			}
			row.Attempt = p.Attempt
			row.LogicalKey = fmt.Sprintf("graph:%d:%s:%d:%s", dl.TenantID, kid, p.Attempt, p.ChunkID)
		} else {
			var op WikiPendingOp
			if err := json.Unmarshal(dl.Payload, &op); err != nil || op.Op != WikiOpIngest || op.KnowledgeID == "" {
				continue
			}
			row.LogicalKey = fmt.Sprintf("wiki:%d:%s:%d", dl.TenantID, kid, attempt)
		}
		// An obsolete or successful dead letter remains recorded, but is not replayed.
		done, err := recoveryAlreadySucceeded(s.db.WithContext(ctx), &row)
		if err != nil {
			return err
		}
		if row.Attempt != attempt || done || k.ParseStatus == types.ParseStatusCancelled || k.ParseStatus == types.ParseStatusDeleting {
			row.State = "skipped"
		}
		if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
				return err
			}
			if row.ID == 0 {
				if err := tx.Where("logical_key=?", row.LogicalKey).First(&row).Error; err != nil {
					return err
				}
			}
			return tx.Exec("INSERT INTO enrichment_recovery_sources(dead_letter_id,recovery_id) VALUES(?,?) ON CONFLICT DO NOTHING", dl.ID, row.ID).Error
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *EnrichmentRecovery) dispatch(ctx context.Context, id int64) error {
	var row EnrichmentRecoveryRow
	if err := s.db.WithContext(ctx).First(&row, id).Error; err != nil {
		return err
	}
	if c := processingcontrol.Current(); c != nil && c.Supported() {
		state, err := c.State(ctx, row.TenantID, row.KnowledgeBaseID)
		if err != nil {
			return err
		}
		if state.RequestedPaused {
			return nil
		}
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
			return err
		}
		if row.State != "ready" && row.State != "queued" {
			return nil
		}
		var k types.Knowledge
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", row.KnowledgeID).First(&k).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return s.finishTx(tx, &row, "skipped", "knowledge deleted", false)
			}
			return err
		}
		attempt, err := latestEnrichmentAttempt(tx, k.ID)
		if err != nil {
			return err
		}
		if attempt != row.Attempt || k.ParseStatus == types.ParseStatusCancelled || k.ParseStatus == types.ParseStatusDeleting {
			return s.finishTx(tx, &row, "skipped", "obsolete or cancelled", attempt == row.Attempt)
		}
		if row.TaskType == types.TypeChunkExtract {
			var p types.ExtractChunkPayload
			if err := json.Unmarshal(row.Payload, &p); err != nil {
				return err
			}
			var n int64
			if err := tx.Table("chunks").Where("id=? AND knowledge_id=? AND deleted_at IS NULL", p.ChunkID, row.KnowledgeID).Count(&n).Error; err != nil {
				return err
			}
			if n == 0 {
				return s.finishTx(tx, &row, "skipped", "chunk deleted", true)
			}
		}
		if row.State == "queued" {
			return nil
		}
		if time.Now().Before(row.AvailableAt) {
			return nil
		}
		done, err := recoveryAlreadySucceeded(tx, &row)
		if err != nil {
			return err
		}
		if done {
			return s.finishTx(tx, &row, "skipped", "already succeeded", true)
		}
		if row.TaskType == types.TypeWikiIngest {
			var n int64
			if err := tx.Model(&types.TaskPendingOp{}).Where("task_type=? AND scope_id=? AND dedup_key=?", wikiTaskType, row.KnowledgeBaseID, k.ID).Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				return nil
			}
		}
		if !row.SlotReserved {
			if err := tx.Model(&types.Knowledge{}).Where("id=?", k.ID).Updates(map[string]interface{}{"pending_subtasks_count": gorm.Expr("pending_subtasks_count+1"), "parse_status": types.ParseStatusFinalizing, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
			row.SlotReserved = true
		}
		row.Attempts++
		row.DispatchNo++
		row.State = "queued"
		now := time.Now()
		row.QueuedAt = &now
		lease := now.Add(90 * time.Minute)
		row.LeaseUntil = &lease
		row.UpdatedAt = now
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		if row.TaskType == types.TypeWikiIngest {
			var op WikiPendingOp
			if err := json.Unmarshal(row.Payload, &op); err != nil {
				return err
			}
			op.RecoveryID = row.ID
			op.RecoveryDispatch = int64(row.DispatchNo)
			payload, err := json.Marshal(op)
			if err != nil {
				return err
			}
			return tx.Create(&types.TaskPendingOp{TenantID: row.TenantID, TaskType: wikiTaskType, Scope: wikiTaskScope, ScopeID: row.KnowledgeBaseID, Op: op.Op, DedupKey: row.KnowledgeID, Payload: payload, EnqueuedAt: now}).Error
		}
		return nil
	})
	if err != nil {
		return err
	}
	if row.State != "queued" {
		return nil
	}
	if row.TaskType == types.TypeWikiIngest {
		taskID := wikiRecoveryTaskID(row)
		if s.inspector != nil {
			info, e := s.inspector.GetTaskInfo(types.QueueWiki, taskID)
			if e == nil && (info.State == asynq.TaskStateCompleted || info.State == asynq.TaskStateArchived) {
				var n int64
				if e := s.db.WithContext(ctx).Model(&types.TaskPendingOp{}).Where("task_type=? AND scope_id=? AND payload->>'recovery_id'=?", wikiTaskType, row.KnowledgeBaseID, fmt.Sprint(row.ID)).Count(&n).Error; e != nil {
					return e
				}
				if n == 0 {
					done, e := recoveryAlreadySucceeded(s.db.WithContext(ctx), &row)
					if e != nil {
						return e
					}
					if done {
						return s.Complete(ctx, row.ID, row.DispatchNo, nil)
					}
					return s.Complete(ctx, row.ID, row.DispatchNo, errors.New("wiki recovery ended without successful completion"))
				}
				if info.State == asynq.TaskStateCompleted {
					done, e := recoveryAlreadySucceeded(s.db.WithContext(ctx), &row)
					if e != nil {
						return e
					}
					if done {
						return s.Complete(ctx, row.ID, row.DispatchNo, nil)
					}
					if row.LeaseUntil != nil && time.Now().After(*row.LeaseUntil) {
						previous := row.QueuedAt
						now := time.Now()
						lease := now.Add(90 * time.Minute)
						result := s.db.WithContext(ctx).Model(&EnrichmentRecoveryRow{}).Where("id=? AND state='queued' AND dispatch_no=? AND queued_at=?", row.ID, row.DispatchNo, previous).Updates(map[string]interface{}{"queued_at": now, "lease_until": lease, "updated_at": now})
						if result.Error != nil {
							return result.Error
						}
						if result.RowsAffected == 0 {
							return nil
						}
						row.QueuedAt = &now
						row.LeaseUntil = &lease
					} else {
						return nil
					}
				}
				// An abnormal batch exit left the durable op intact; settle this single
				// recovery attempt instead of creating another short-interval batch loop.
				if info.State == asynq.TaskStateArchived {
					return s.CompleteWikiAbandoned(ctx, row, info.LastErr)
				}
			}
		}
		payload, _ := json.Marshal(WikiIngestPayload{TenantID: row.TenantID, KnowledgeBaseID: row.KnowledgeBaseID})
		_, err = s.tasks.Enqueue(asynq.NewTask(types.TypeWikiIngest, payload, asynq.Queue(types.QueueWiki), asynq.TaskID(wikiRecoveryTaskID(row)), asynq.MaxRetry(0), asynq.Timeout(time.Hour), asynq.Retention(24*time.Hour)))
	} else if row.TaskType == types.TypeSummaryGeneration {
		var p types.SummaryGenerationPayload
		if err = json.Unmarshal(row.Payload, &p); err != nil {
			return err
		}
		p.RecoveryID, p.RecoveryDispatch = row.ID, int64(row.DispatchNo)
		payload, _ := json.Marshal(p)
		_, err = s.tasks.Enqueue(asynq.NewTask(types.TypeSummaryGeneration, payload,
			asynq.Queue(types.QueueSummary), asynq.TaskID(fmt.Sprintf("recovery-summary-%d-%d", row.ID, row.DispatchNo)),
			asynq.MaxRetry(0), asynq.Timeout(30*time.Minute), asynq.Retention(24*time.Hour)))
	} else {
		var p types.ExtractChunkPayload
		if err = json.Unmarshal(row.Payload, &p); err != nil {
			return err
		}
		p.RecoveryID = row.ID
		p.RecoveryDispatch = int64(row.DispatchNo)
		payload, _ := json.Marshal(p)
		taskID := fmt.Sprintf("recovery-graph-%d-%d", row.ID, row.DispatchNo)
		if s.inspector != nil {
			info, e := s.inspector.GetTaskInfo(types.QueueGraph, taskID)
			if e == nil && info.State == asynq.TaskStateArchived {
				return s.Complete(ctx, row.ID, row.DispatchNo, errors.New(info.LastErr))
			}
			if e == nil && info.State == asynq.TaskStateCompleted {
				return s.Complete(ctx, row.ID, row.DispatchNo, nil)
			}
		}
		_, err = s.tasks.Enqueue(asynq.NewTask(types.TypeChunkExtract, payload, asynq.Queue(types.QueueGraph), asynq.TaskID(taskID), asynq.MaxRetry(0), asynq.Timeout(30*time.Minute), asynq.Retention(24*time.Hour)))
	}
	if errors.Is(err, asynq.ErrTaskIDConflict) || errors.Is(err, asynq.ErrDuplicateTask) {
		return nil
	}
	return err
}

// finishTx updates both the recovery owner and its counter in one transaction.
func (s *EnrichmentRecovery) finishTx(tx *gorm.DB, row *EnrichmentRecoveryRow, state, message string, release bool) error {
	if row.SlotReserved && release {
		if err := tx.Exec(`UPDATE knowledges SET pending_subtasks_count=GREATEST(pending_subtasks_count-1,0),
   parse_status=CASE WHEN parse_status='finalizing' AND pending_subtasks_count<=1 THEN 'completed' ELSE parse_status END,
   updated_at=NOW() WHERE id=? AND deleted_at IS NULL`, row.KnowledgeID).Error; err != nil {
			return err
		}
	}
	row.SlotReserved = false
	row.State = state
	row.LastError = message
	row.UpdatedAt = time.Now()
	return tx.Save(row).Error
}

func (s *EnrichmentRecovery) Complete(ctx context.Context, id int64, dispatch int, taskErr error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row EnrichmentRecoveryRow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
			return err
		}
		if row.State != "queued" || row.DispatchNo != dispatch {
			return nil
		}
		attempt, err := latestEnrichmentAttempt(tx, row.KnowledgeID)
		if err != nil {
			return err
		}
		if attempt != row.Attempt {
			return s.finishTx(tx, &row, "skipped", "superseded", false)
		}
		if row.TaskType == types.TypeWikiIngest {
			if err := tx.Where("task_type=? AND scope_id=? AND payload->>'recovery_id'=? AND payload->>'recovery_dispatch'=?", wikiTaskType, row.KnowledgeBaseID, fmt.Sprint(row.ID), fmt.Sprint(row.DispatchNo)).Delete(&types.TaskPendingOp{}).Error; err != nil {
				return err
			}
		}
		if taskErr == nil {
			return s.finishTx(tx, &row, "recovered", "", true)
		}
		if errors.Is(taskErr, context.Canceled) || errors.Is(taskErr, processingcontrol.ErrPaused) {
			row.Attempts--
		}
		if row.Attempts >= 2 {
			return s.finishTx(tx, &row, "exhausted", taskErr.Error(), true)
		}
		row.State = "ready"
		row.LastError = taskErr.Error()
		row.AvailableAt = time.Now().Add(EnrichmentRecoveryCooldown)
		row.UpdatedAt = time.Now()
		return tx.Save(&row).Error
	})
}

func completeEnrichmentRecovery(ctx context.Context, id int64, dispatch int64, taskErr error) error {
	s := enrichmentRecovery.Load()
	if s == nil {
		return errors.New("enrichment recovery service unavailable")
	}
	c, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	return s.Complete(c, id, int(dispatch), taskErr)
}

func (s *EnrichmentRecovery) CompleteWikiAbandoned(ctx context.Context, row EnrichmentRecoveryRow, message string) error {
	if err := s.Complete(ctx, row.ID, row.DispatchNo, errors.New(message)); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Where("task_type=? AND scope_id=? AND payload->>'recovery_id'=? AND payload->>'recovery_dispatch'=?", wikiTaskType, row.KnowledgeBaseID, fmt.Sprint(row.ID), fmt.Sprint(row.DispatchNo)).Delete(&types.TaskPendingOp{}).Error
}

// EnrichmentRecoveryRuntime is additive to the existing runtime queue response.
func EnrichmentRecoveryRuntime(ctx context.Context) map[string]interface{} {
	out := map[string]interface{}{"ordinary_cooldown_seconds": 1800, "recovery_cooldown_seconds": 7200, "max_recovery_attempts": 2, "graph_concurrency": 8}
	if s := enrichmentRecovery.Load(); s != nil {
		var counts []struct {
			State  string
			Count  int64
			NextAt *time.Time
		}
		if err := s.db.WithContext(ctx).Model(&EnrichmentRecoveryRow{}).Select("state,count(*) AS count,min(available_at) AS next_at").Group("state").Scan(&counts).Error; err != nil {
			out["error"] = err.Error()
		} else {
			out["states"] = counts
		}
	}
	return out
}

func (s *EnrichmentRecovery) sweepSafely(ctx context.Context) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("recovery sweep panic (will retry): %v", p)
		}
	}()
	return s.Sweep(ctx)
}

func wikiRecoveryTaskID(row EnrichmentRecoveryRow) string {
	var stamp int64
	if row.QueuedAt != nil {
		stamp = row.QueuedAt.UnixNano()
	}
	return fmt.Sprintf("recovery-wiki-%d-%d-%d", row.ID, row.DispatchNo, stamp)
}

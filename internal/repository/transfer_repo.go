package repository

import (
	"context"
	"errors"
	"fss-mining/internal/model"
	"time"

	"gorm.io/gorm"
)

type TransferRepo struct {
	db *gorm.DB
}

func NewTransferRepo(db *gorm.DB) *TransferRepo {
	return &TransferRepo{db: db}
}

func (r *TransferRepo) Create(ctx context.Context, tx *gorm.DB, t *model.Transfer) error {
	return tx.WithContext(ctx).Create(t).Error
}

func (r *TransferRepo) GetByID(ctx context.Context, id uint) (*model.Transfer, error) {
	t := &model.Transfer{}
	err := r.db.WithContext(ctx).First(t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return t, err
}

func (r *TransferRepo) GetByIdempotentKey(ctx context.Context, key string) (*model.Transfer, error) {
	t := &model.Transfer{}
	err := r.db.WithContext(ctx).Where("idempotent_key = ?", key).First(t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return t, err
}

func (r *TransferRepo) UpdateStatus(ctx context.Context, tx *gorm.DB, id uint, status model.TransferStatus, extra map[string]any) error {
	updates := map[string]any{"status": status}
	for k, v := range extra {
		updates[k] = v
	}
	return tx.WithContext(ctx).Model(&model.Transfer{}).Where("id = ?", id).Updates(updates).Error
}

// ListByUser 分页查询用户划转记录
func (r *TransferRepo) ListByUser(ctx context.Context, userID uint, page, size int) ([]*model.Transfer, int64, error) {
	var list []*model.Transfer
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Transfer{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// ScanPendingTransfers 扫描待处理划转（满足24小时延迟条件的）
func (r *TransferRepo) ScanPendingTransfers(ctx context.Context, beforeTime time.Time) ([]*model.Transfer, error) {
	var list []*model.Transfer
	err := r.db.WithContext(ctx).
		Where("status = ? AND apply_at <= ?", model.TransferStatusLocking, beforeTime).
		Find(&list).Error
	return list, err
}

// ScanEffectiveTransfers 扫描已过24小时、等待次日00:00生效的转入记录
func (r *TransferRepo) ScanEffectiveTransfers(ctx context.Context, now time.Time) ([]*model.Transfer, error) {
	var list []*model.Transfer
	err := r.db.WithContext(ctx).
		Where("status = ? AND direction = ? AND effective_at <= ?",
			model.TransferStatusEffective, model.TransferDirectionIn, now).
		Find(&list).Error
	return list, err
}

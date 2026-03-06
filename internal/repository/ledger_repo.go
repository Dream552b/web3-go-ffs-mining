package repository

import (
	"context"
	"fss-mining/internal/model"

	"gorm.io/gorm"
)

type LedgerRepo struct {
	db *gorm.DB
}

func NewLedgerRepo(db *gorm.DB) *LedgerRepo {
	return &LedgerRepo{db: db}
}

func (r *LedgerRepo) Create(ctx context.Context, tx *gorm.DB, l *model.Ledger) error {
	return tx.WithContext(ctx).Create(l).Error
}

// ListByUser 分页查询用户流水
func (r *LedgerRepo) ListByUser(ctx context.Context, userID uint, ledgerType string, page, size int) ([]*model.Ledger, int64, error) {
	var list []*model.Ledger
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Ledger{}).Where("user_id = ?", userID)
	if ledgerType != "" {
		query = query.Where("ledger_type = ?", ledgerType)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

func (r *LedgerRepo) ExistsByIdempotentKey(ctx context.Context, key string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Ledger{}).
		Where("idempotent_key = ?", key).Count(&count).Error
	return count > 0, err
}

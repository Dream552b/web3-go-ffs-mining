package repository

import (
	"context"
	"errors"
	"fss-mining/internal/model"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MiningRepo struct {
	db *gorm.DB
}

func NewMiningRepo(db *gorm.DB) *MiningRepo {
	return &MiningRepo{db: db}
}

// GetActiveConfig 获取当前激活的挖矿配置
func (r *MiningRepo) GetActiveConfig(ctx context.Context) (*model.MiningPeriodConfig, error) {
	cfg := &model.MiningPeriodConfig{}
	err := r.db.WithContext(ctx).Where("is_active = ?", true).First(cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return cfg, err
}

// GetConfigByMonth 按月份获取配置
func (r *MiningRepo) GetConfigByMonth(ctx context.Context, month string) (*model.MiningPeriodConfig, error) {
	cfg := &model.MiningPeriodConfig{}
	err := r.db.WithContext(ctx).Where("period_month = ?", month).First(cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return cfg, err
}

func (r *MiningRepo) CreateConfig(ctx context.Context, cfg *model.MiningPeriodConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *MiningRepo) ActivateConfig(ctx context.Context, month string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.MiningPeriodConfig{}).Where("is_active = ?", true).
			Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Model(&model.MiningPeriodConfig{}).Where("period_month = ?", month).
			Update("is_active", true).Error
	})
}

// CreateSnapshot 创建日快照
func (r *MiningRepo) CreateSnapshot(ctx context.Context, snap *model.StakeSnapshot) error {
	return r.db.WithContext(ctx).Create(snap).Error
}

// BatchCreateSnapshot 批量创建日快照
func (r *MiningRepo) BatchCreateSnapshot(ctx context.Context, snaps []*model.StakeSnapshot) error {
	return r.db.WithContext(ctx).CreateInBatches(snaps, 500).Error
}

// GetSnapshotByUserDate 获取用户某天快照
func (r *MiningRepo) GetSnapshotByUserDate(ctx context.Context, userID uint, date time.Time) (*model.StakeSnapshot, error) {
	snap := &model.StakeSnapshot{}
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND snapshot_date = ?", userID, date.Format("2006-01-02")).
		First(snap).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return snap, err
}

// GetTotalNetworkPower 获取某天全网有效算力
func (r *MiningRepo) GetTotalNetworkPower(ctx context.Context, date time.Time) (decimal.Decimal, int, error) {
	type Result struct {
		TotalPower decimal.Decimal
		ValidCount int
	}
	var res Result
	err := r.db.WithContext(ctx).Model(&model.StakeSnapshot{}).
		Select("SUM(effective_power) as total_power, COUNT(*) as valid_count").
		Where("snapshot_date = ? AND is_valid = ?", date.Format("2006-01-02"), true).
		Scan(&res).Error
	return res.TotalPower, res.ValidCount, err
}

// GetBatchByDate 获取结算批次
func (r *MiningRepo) GetBatchByDate(ctx context.Context, date time.Time) (*model.SettleBatch, error) {
	batch := &model.SettleBatch{}
	err := r.db.WithContext(ctx).
		Where("batch_date = ?", date.Format("2006-01-02")).
		First(batch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return batch, err
}

func (r *MiningRepo) CreateBatch(ctx context.Context, batch *model.SettleBatch) error {
	return r.db.WithContext(ctx).Create(batch).Error
}

func (r *MiningRepo) UpdateBatch(ctx context.Context, id uint, updates map[string]any) error {
	return r.db.WithContext(ctx).Model(&model.SettleBatch{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MiningRepo) CreateSettlement(ctx context.Context, tx *gorm.DB, s *model.DailySettlement) error {
	return tx.WithContext(ctx).Create(s).Error
}

func (r *MiningRepo) BatchCreateSettlement(ctx context.Context, tx *gorm.DB, list []*model.DailySettlement) error {
	return tx.WithContext(ctx).CreateInBatches(list, 500).Error
}

// GetUserSettlements 分页查询用户结算记录
func (r *MiningRepo) GetUserSettlements(ctx context.Context, userID uint, page, size int) ([]*model.DailySettlement, int64, error) {
	var list []*model.DailySettlement
	var total int64
	query := r.db.WithContext(ctx).Model(&model.DailySettlement{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("settle_date DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error
	return list, total, err
}

// GetUserTotalReward 获取用户累计奖励
func (r *MiningRepo) GetUserTotalReward(ctx context.Context, userID uint) (decimal.Decimal, error) {
	type Result struct {
		Total decimal.Decimal
	}
	var res Result
	err := r.db.WithContext(ctx).Model(&model.DailySettlement{}).
		Select("SUM(total_reward) as total").
		Where("user_id = ? AND status = ?", userID, "paid").
		Scan(&res).Error
	return res.Total, err
}

// GetNetworkStats 全网统计
func (r *MiningRepo) GetNetworkStats(ctx context.Context, date time.Time) (map[string]any, error) {
	type StakeStats struct {
		TotalStake    decimal.Decimal
		TotalEffective decimal.Decimal
		ValidUsers    int
	}
	var stats StakeStats
	err := r.db.WithContext(ctx).Model(&model.StakeSnapshot{}).
		Select("SUM(stake_amount) as total_stake, SUM(effective_power) as total_effective, COUNT(*) as valid_users").
		Where("snapshot_date = ? AND is_valid = ?", date.Format("2006-01-02"), true).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"total_stake":     stats.TotalStake,
		"total_effective": stats.TotalEffective,
		"valid_users":     stats.ValidUsers,
	}, nil
}

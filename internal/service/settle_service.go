package service

import (
	"context"
	"fmt"
	"fss-mining/internal/model"
	"fss-mining/internal/repository"
	"fss-mining/pkg/logger"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SettleService struct {
	db          *gorm.DB
	accountRepo *repository.AccountRepo
	miningRepo  *repository.MiningRepo
	ledgerRepo  *repository.LedgerRepo
}

func NewSettleService(
	db *gorm.DB,
	ar *repository.AccountRepo,
	mr *repository.MiningRepo,
	lr *repository.LedgerRepo,
) *SettleService {
	return &SettleService{db: db, accountRepo: ar, miningRepo: mr, ledgerRepo: lr}
}

// TakeSnapshot 对当日所有挖矿区账户生成算力快照（每日 00:05 执行，对昨日做快照）
func (s *SettleService) TakeSnapshot(ctx context.Context, snapshotDate time.Time) error {
	cfg, err := s.miningRepo.GetActiveConfig(ctx)
	if err != nil || cfg == nil {
		return fmt.Errorf("未找到激活的挖矿配置")
	}

	accs, err := s.accountRepo.GetMiningAccounts(ctx)
	if err != nil {
		return err
	}

	snaps := make([]*model.StakeSnapshot, 0, len(accs))
	for _, acc := range accs {
		power, valid := CalcEffectivePower(acc.Balance, cfg.MinStake, cfg.MaxStake)
		snaps = append(snaps, &model.StakeSnapshot{
			UserID:         acc.UserID,
			SnapshotDate:   snapshotDate,
			StakeAmount:    acc.Balance,
			EffectivePower: power,
			IsValid:        valid,
		})
	}

	logger.Info("生成算力快照", zap.Time("date", snapshotDate), zap.Int("count", len(snaps)))
	return s.miningRepo.BatchCreateSnapshot(ctx, snaps)
}

// RunDailySettle 执行昨日结算（每日 00:30 执行）
func (s *SettleService) RunDailySettle(ctx context.Context, settleDate time.Time) error {
	// 幂等：如果该日批次已存在且 completed，则跳过
	existBatch, err := s.miningRepo.GetBatchByDate(ctx, settleDate)
	if err != nil {
		return err
	}
	if existBatch != nil && existBatch.Status == "completed" {
		logger.Info("结算批次已完成，跳过", zap.Time("date", settleDate))
		return nil
	}

	cfg, err := s.miningRepo.GetActiveConfig(ctx)
	if err != nil || cfg == nil {
		return fmt.Errorf("未找到激活的挖矿配置")
	}

	// 1. 获取全网算力
	totalPower, validCount, err := s.miningRepo.GetTotalNetworkPower(ctx, settleDate)
	if err != nil {
		return err
	}
	if totalPower.IsZero() {
		logger.Warn("全网有效算力为 0，跳过结算", zap.Time("date", settleDate))
		return nil
	}

	dailyOutput := cfg.DailyOutput
	staticPool := dailyOutput.Mul(cfg.StaticRatio).Round(8)
	dynamicPool := dailyOutput.Mul(cfg.DynamicRatio).Round(8)

	// 2. 创建结算批次
	batch := &model.SettleBatch{
		BatchDate:         settleDate,
		TotalOutput:       dailyOutput,
		StaticPool:        staticPool,
		DynamicPool:       dynamicPool,
		TotalNetworkPower: totalPower,
		ValidUserCount:    validCount,
		Status:            "running",
	}
	if existBatch == nil {
		if err := s.miningRepo.CreateBatch(ctx, batch); err != nil {
			return err
		}
	} else {
		batch = existBatch
		_ = s.miningRepo.UpdateBatch(ctx, batch.ID, map[string]any{"status": "running"})
	}

	// 3. 获取当日有效快照
	var snapshots []*model.StakeSnapshot
	err = s.db.WithContext(ctx).
		Where("snapshot_date = ? AND is_valid = ?", settleDate.Format("2006-01-02"), true).
		Find(&snapshots).Error
	if err != nil {
		return err
	}

	// 4. 计算每个用户的静态奖励
	settlements := make([]*model.DailySettlement, 0, len(snapshots))
	var totalPaid decimal.Decimal

	for _, snap := range snapshots {
		staticReward := CalcStaticReward(snap.EffectivePower, totalPower, staticPool)
		// 动态奖励待后续实现
		dynamicReward := decimal.Zero

		settlements = append(settlements, &model.DailySettlement{
			SettleDate:     settleDate,
			UserID:         snap.UserID,
			StaticReward:   staticReward,
			DynamicReward:  dynamicReward,
			TotalReward:    staticReward.Add(dynamicReward),
			EffectivePower: snap.EffectivePower,
			NetworkPower:   totalPower,
			Status:         "settled",
			BatchID:        batch.ID,
		})
		totalPaid = totalPaid.Add(staticReward)
	}

	// 5. 批量写入结算记录并发放奖励
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.miningRepo.BatchCreateSettlement(ctx, tx, settlements); err != nil {
			return err
		}

		// 6. 逐用户发放到挖矿区账户
		for _, settle := range settlements {
			if settle.TotalReward.IsZero() {
				continue
			}
			if err := s.accountRepo.Credit(ctx, tx, settle.UserID, model.AccountTypeMining, settle.TotalReward); err != nil {
				return fmt.Errorf("发放奖励失败 userID=%d: %w", settle.UserID, err)
			}
			// 记录流水
			miningAcc, _ := s.accountRepo.GetByUserAndType(ctx, settle.UserID, model.AccountTypeMining)
			ledger := &model.Ledger{
				UserID:        settle.UserID,
				AccountType:   model.AccountTypeMining,
				LedgerType:    model.LedgerTypeMiningReward,
				Amount:        settle.TotalReward,
				BalanceAfter:  miningAcc.Balance.Add(settle.TotalReward),
				RelatedID:     &settle.ID,
				Remark:        fmt.Sprintf("%s 静态挖矿奖励", settleDate.Format("2006-01-02")),
				IdempotentKey: fmt.Sprintf("mining-reward-%d-%d", settle.UserID, settleDate.Unix()),
			}
			if err := s.ledgerRepo.Create(ctx, tx, ledger); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		_ = s.miningRepo.UpdateBatch(ctx, batch.ID, map[string]any{"status": "failed", "remark": err.Error()})
		return err
	}

	// 7. 更新批次状态
	_ = s.miningRepo.UpdateBatch(ctx, batch.ID, map[string]any{
		"status":      "completed",
		"actual_paid": totalPaid,
	})

	logger.Info("日结算完成",
		zap.Time("date", settleDate),
		zap.String("total_paid", totalPaid.String()),
		zap.Int("users", len(settlements)),
	)
	return nil
}

// InitPeriodConfig 初始化或生成某月的挖矿配置
func (s *SettleService) InitPeriodConfig(ctx context.Context, targetMonth time.Time, initialCirculation, growthRate decimal.Decimal, monthNumber int) error {
	periodMonth := GetPeriodMonth(targetMonth)

	exist, err := s.miningRepo.GetConfigByMonth(ctx, periodMonth)
	if err != nil {
		return err
	}
	if exist != nil {
		return fmt.Errorf("配置已存在: %s", periodMonth)
	}

	circulation, monthlyOutput := CalcMonthlyOutput(monthNumber, initialCirculation, growthRate)
	days := GetDaysInMonth(targetMonth)
	dailyOutput := CalcDailyOutput(monthlyOutput, days)

	staticRatio := decimal.NewFromFloat(0.5)
	dynamicRatio := decimal.NewFromFloat(0.5)

	cfg := &model.MiningPeriodConfig{
		PeriodMonth:        periodMonth,
		InitialCirculation: circulation,
		MonthlyOutput:      monthlyOutput,
		DailyOutput:        dailyOutput,
		StaticRatio:        staticRatio,
		DynamicRatio:       dynamicRatio,
		MinStake:           decimal.NewFromFloat(100),
		MaxStake:           decimal.NewFromFloat(2000),
		TotalDays:          days,
		EffectiveDate:      time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, targetMonth.Location()),
		IsActive:           false,
	}

	return s.miningRepo.CreateConfig(ctx, cfg)
}

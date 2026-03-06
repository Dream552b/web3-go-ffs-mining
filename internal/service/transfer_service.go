package service

import (
	"context"
	"fmt"
	"fss-mining/internal/model"
	"fss-mining/internal/repository"
	"fss-mining/pkg/config"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransferService struct {
	db          *gorm.DB
	accountRepo *repository.AccountRepo
	transferRepo *repository.TransferRepo
	ledgerRepo  *repository.LedgerRepo
}

func NewTransferService(
	db *gorm.DB,
	ar *repository.AccountRepo,
	tr *repository.TransferRepo,
	lr *repository.LedgerRepo,
) *TransferService {
	return &TransferService{db: db, accountRepo: ar, transferRepo: tr, ledgerRepo: lr}
}

// TransferIn 从交易区转入挖矿区
func (s *TransferService) TransferIn(ctx context.Context, userID uint, amount decimal.Decimal, idempotentKey string) (*model.Transfer, error) {
	// 幂等检查
	exist, err := s.transferRepo.GetByIdempotentKey(ctx, idempotentKey)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return exist, nil
	}

	cfg := config.Get().Mining
	minStake := decimal.NewFromFloat(cfg.MinStake)
	if amount.LessThan(minStake) {
		return nil, fmt.Errorf("划转金额不能小于最低质押门槛 %s FSS", minStake)
	}

	now := time.Now()
	effectiveAt := CalcTransferEffectiveTime(now, cfg.TransferDelayHours)

	transfer := &model.Transfer{
		UserID:        userID,
		Direction:     model.TransferDirectionIn,
		Amount:        amount,
		Status:        model.TransferStatusLocking,
		ApplyAt:       now,
		EffectiveAt:   &effectiveAt,
		IdempotentKey: idempotentKey,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 确保挖矿区账户存在
		_, err := s.accountRepo.GetOrCreate(ctx, userID, model.AccountTypeMining)
		if err != nil {
			return err
		}

		// 2. 冻结交易区余额
		if err := s.accountRepo.Freeze(ctx, tx, userID, model.AccountTypeTrade, amount); err != nil {
			return err
		}

		// 3. 创建划转记录
		if err := s.transferRepo.Create(ctx, tx, transfer); err != nil {
			return err
		}

		// 4. 记录流水（冻结操作）
		tradeAcc, err := s.accountRepo.GetByUserAndType(ctx, userID, model.AccountTypeTrade)
		if err != nil {
			return err
		}
		ledger := &model.Ledger{
			UserID:        userID,
			AccountType:   model.AccountTypeTrade,
			LedgerType:    model.LedgerTypeFreeze,
			Amount:        amount.Neg(),
			BalanceAfter:  tradeAcc.AvailableBalance(),
			RelatedID:     &transfer.ID,
			Remark:        fmt.Sprintf("划转转入挖矿区冻结，生效时间: %s", effectiveAt.Format("2006-01-02 15:04:05")),
			IdempotentKey: "freeze-" + idempotentKey,
		}
		return s.ledgerRepo.Create(ctx, tx, ledger)
	})

	if err != nil {
		return nil, err
	}
	return transfer, nil
}

// TransferOut 从挖矿区转出到交易区
func (s *TransferService) TransferOut(ctx context.Context, userID uint, amount decimal.Decimal, idempotentKey string) (*model.Transfer, error) {
	exist, err := s.transferRepo.GetByIdempotentKey(ctx, idempotentKey)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return exist, nil
	}

	now := time.Now()
	cfg := config.Get().Mining
	// 转出也需要24小时延迟
	completedAt := now.Add(time.Duration(cfg.TransferDelayHours) * time.Hour)

	transfer := &model.Transfer{
		UserID:        userID,
		Direction:     model.TransferDirectionOut,
		Amount:        amount,
		Status:        model.TransferStatusLocking,
		ApplyAt:       now,
		EffectiveAt:   &completedAt,
		IdempotentKey: idempotentKey,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 冻结挖矿区余额
		if err := s.accountRepo.Freeze(ctx, tx, userID, model.AccountTypeMining, amount); err != nil {
			return err
		}

		if err := s.transferRepo.Create(ctx, tx, transfer); err != nil {
			return err
		}

		miningAcc, err := s.accountRepo.GetByUserAndType(ctx, userID, model.AccountTypeMining)
		if err != nil {
			return err
		}
		ledger := &model.Ledger{
			UserID:        userID,
			AccountType:   model.AccountTypeMining,
			LedgerType:    model.LedgerTypeFreeze,
			Amount:        amount.Neg(),
			BalanceAfter:  miningAcc.AvailableBalance(),
			RelatedID:     &transfer.ID,
			Remark:        "划转转出挖矿区冻结",
			IdempotentKey: "freeze-out-" + idempotentKey,
		}
		return s.ledgerRepo.Create(ctx, tx, ledger)
	})

	if err != nil {
		return nil, err
	}
	return transfer, nil
}

// CompleteTransferIn 划转转入完成（到达生效时间，实际从交易区扣款并入挖矿区）
func (s *TransferService) CompleteTransferIn(ctx context.Context, transfer *model.Transfer) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. 解冻交易区并扣款
		if err := s.accountRepo.UnfreezeAndDebit(ctx, tx, transfer.UserID, model.AccountTypeTrade, transfer.Amount); err != nil {
			return err
		}
		// 2. 挖矿区入账
		if err := s.accountRepo.Credit(ctx, tx, transfer.UserID, model.AccountTypeMining, transfer.Amount); err != nil {
			return err
		}
		// 3. 更新划转状态
		completedAt := time.Now()
		if err := s.transferRepo.UpdateStatus(ctx, tx, transfer.ID, model.TransferStatusCompleted, map[string]any{
			"completed_at": completedAt,
		}); err != nil {
			return err
		}
		// 4. 记录流水
		miningAcc, err := s.accountRepo.GetByUserAndType(ctx, transfer.UserID, model.AccountTypeMining)
		if err != nil {
			return err
		}
		ledger := &model.Ledger{
			UserID:        transfer.UserID,
			AccountType:   model.AccountTypeMining,
			LedgerType:    model.LedgerTypeTransferIn,
			Amount:        transfer.Amount,
			BalanceAfter:  miningAcc.Balance,
			RelatedID:     &transfer.ID,
			Remark:        "划转转入挖矿区到账",
			IdempotentKey: fmt.Sprintf("transfer-in-complete-%d", transfer.ID),
		}
		return s.ledgerRepo.Create(ctx, tx, ledger)
	})
}

// ScanAndCompleteTransfers 扫描并批量完成满足条件的划转单
func (s *TransferService) ScanAndCompleteTransfers(ctx context.Context) error {
	now := time.Now()

	// 扫描转入：已过24h且到达 effective_at 的
	inList, err := s.transferRepo.ScanEffectiveTransfers(ctx, now)
	if err != nil {
		return err
	}
	for _, t := range inList {
		if err := s.CompleteTransferIn(ctx, t); err != nil {
			// 单条失败不影响其他，继续处理
			_ = err
		}
	}

	// 扫描转出：已过24h的（effective_at <= now）
	outList, err := s.transferRepo.ScanPendingTransfers(ctx, now)
	if err != nil {
		return err
	}
	for _, t := range outList {
		if t.Direction == model.TransferDirectionOut {
			if err := s.CompleteTransferOut(ctx, t); err != nil {
				_ = err
			}
		}
	}
	return nil
}

// CompleteTransferOut 划转转出完成（24h 后，挖矿区扣款并入交易区）
func (s *TransferService) CompleteTransferOut(ctx context.Context, transfer *model.Transfer) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.accountRepo.UnfreezeAndDebit(ctx, tx, transfer.UserID, model.AccountTypeMining, transfer.Amount); err != nil {
			return err
		}
		// 确保交易区账户存在
		if _, err := s.accountRepo.GetOrCreate(ctx, transfer.UserID, model.AccountTypeTrade); err != nil {
			return err
		}
		if err := s.accountRepo.Credit(ctx, tx, transfer.UserID, model.AccountTypeTrade, transfer.Amount); err != nil {
			return err
		}
		completedAt := time.Now()
		if err := s.transferRepo.UpdateStatus(ctx, tx, transfer.ID, model.TransferStatusCompleted, map[string]any{
			"completed_at": completedAt,
		}); err != nil {
			return err
		}
		tradeAcc, err := s.accountRepo.GetByUserAndType(ctx, transfer.UserID, model.AccountTypeTrade)
		if err != nil {
			return err
		}
		ledger := &model.Ledger{
			UserID:        transfer.UserID,
			AccountType:   model.AccountTypeTrade,
			LedgerType:    model.LedgerTypeTransferOut,
			Amount:        transfer.Amount,
			BalanceAfter:  tradeAcc.Balance,
			RelatedID:     &transfer.ID,
			Remark:        "划转转出到交易区到账",
			IdempotentKey: fmt.Sprintf("transfer-out-complete-%d", transfer.ID),
		}
		return s.ledgerRepo.Create(ctx, tx, ledger)
	})
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"fss-mining/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (r *AccountRepo) GetOrCreate(ctx context.Context, userID uint, accountType model.AccountType) (*model.Account, error) {
	acc := &model.Account{}
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND account_type = ?", userID, accountType).
		FirstOrCreate(acc, model.Account{
			UserID:      userID,
			AccountType: accountType,
		}).Error
	return acc, err
}

func (r *AccountRepo) GetByUserAndType(ctx context.Context, userID uint, accountType model.AccountType) (*model.Account, error) {
	acc := &model.Account{}
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND account_type = ?", userID, accountType).
		First(acc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return acc, err
}

// GetForUpdate 加行锁读取（事务内使用）
func (r *AccountRepo) GetForUpdate(ctx context.Context, tx *gorm.DB, userID uint, accountType model.AccountType) (*model.Account, error) {
	acc := &model.Account{}
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND account_type = ?", userID, accountType).
		First(acc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("账户不存在: userID=%d type=%s", userID, accountType)
	}
	return acc, err
}

// Debit 扣款（带行锁，事务内使用）
func (r *AccountRepo) Debit(ctx context.Context, tx *gorm.DB, userID uint, accountType model.AccountType, amount decimal.Decimal) error {
	acc, err := r.GetForUpdate(ctx, tx, userID, accountType)
	if err != nil {
		return err
	}
	if acc.AvailableBalance().LessThan(amount) {
		return fmt.Errorf("余额不足: available=%s need=%s", acc.AvailableBalance(), amount)
	}
	return tx.WithContext(ctx).Model(acc).
		Where("version = ?", acc.Version).
		Updates(map[string]any{
			"balance": acc.Balance.Sub(amount),
			"version": acc.Version + 1,
		}).Error
}

// Credit 入账（带行锁，事务内使用）
func (r *AccountRepo) Credit(ctx context.Context, tx *gorm.DB, userID uint, accountType model.AccountType, amount decimal.Decimal) error {
	acc, err := r.GetForUpdate(ctx, tx, userID, accountType)
	if err != nil {
		return err
	}
	return tx.WithContext(ctx).Model(acc).
		Where("version = ?", acc.Version).
		Updates(map[string]any{
			"balance": acc.Balance.Add(amount),
			"version": acc.Version + 1,
		}).Error
}

// Freeze 冻结余额
func (r *AccountRepo) Freeze(ctx context.Context, tx *gorm.DB, userID uint, accountType model.AccountType, amount decimal.Decimal) error {
	acc, err := r.GetForUpdate(ctx, tx, userID, accountType)
	if err != nil {
		return err
	}
	if acc.AvailableBalance().LessThan(amount) {
		return fmt.Errorf("可用余额不足以冻结: available=%s need=%s", acc.AvailableBalance(), amount)
	}
	return tx.WithContext(ctx).Model(acc).
		Where("version = ?", acc.Version).
		Updates(map[string]any{
			"frozen_amt": acc.FrozenAmt.Add(amount),
			"version":    acc.Version + 1,
		}).Error
}

// Unfreeze 解冻并扣减余额（划转完成时调用）
func (r *AccountRepo) UnfreezeAndDebit(ctx context.Context, tx *gorm.DB, userID uint, accountType model.AccountType, amount decimal.Decimal) error {
	acc, err := r.GetForUpdate(ctx, tx, userID, accountType)
	if err != nil {
		return err
	}
	if acc.FrozenAmt.LessThan(amount) {
		return fmt.Errorf("冻结金额不足: frozen=%s need=%s", acc.FrozenAmt, amount)
	}
	return tx.WithContext(ctx).Model(acc).
		Where("version = ?", acc.Version).
		Updates(map[string]any{
			"balance":    acc.Balance.Sub(amount),
			"frozen_amt": acc.FrozenAmt.Sub(amount),
			"version":    acc.Version + 1,
		}).Error
}

// GetMiningAccounts 获取所有挖矿区账户（用于快照和结算）
func (r *AccountRepo) GetMiningAccounts(ctx context.Context) ([]*model.Account, error) {
	var accs []*model.Account
	err := r.db.WithContext(ctx).
		Where("account_type = ? AND balance > 0", model.AccountTypeMining).
		Find(&accs).Error
	return accs, err
}

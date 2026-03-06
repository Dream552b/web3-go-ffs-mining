package model

import "github.com/shopspring/decimal"

// AccountType 账户类型
type AccountType string

const (
	AccountTypeTrade  AccountType = "trade"  // 交易区账户
	AccountTypeMining AccountType = "mining" // 挖矿区账户
)

// Account 账户表（每个用户有交易区和挖矿区两个账户）
type Account struct {
	BaseModel
	UserID      uint            `gorm:"uniqueIndex:idx_user_type;not null;index" json:"user_id"`
	AccountType AccountType     `gorm:"uniqueIndex:idx_user_type;size:16;not null" json:"account_type"`
	Balance     decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"balance"`
	FrozenAmt   decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"frozen_amt"` // 冻结中（划转待处理）
	Version     int64           `gorm:"default:0" json:"-"`                             // 乐观锁
}

func (Account) TableName() string { return "accounts" }

// AvailableBalance 可用余额 = 余额 - 冻结
func (a *Account) AvailableBalance() decimal.Decimal {
	return a.Balance.Sub(a.FrozenAmt)
}

package model

import "github.com/shopspring/decimal"

// LedgerType 流水类型
type LedgerType string

const (
	LedgerTypeTransferIn    LedgerType = "transfer_in"    // 从交易区转入挖矿区
	LedgerTypeTransferOut   LedgerType = "transfer_out"   // 从挖矿区转出到交易区
	LedgerTypeMiningReward  LedgerType = "mining_reward"  // 挖矿静态奖励
	LedgerTypeDynamicReward LedgerType = "dynamic_reward" // 挖矿动态奖励
	LedgerTypeFreeze        LedgerType = "freeze"         // 冻结
	LedgerTypeUnfreeze      LedgerType = "unfreeze"       // 解冻
	LedgerTypeWithdraw      LedgerType = "withdraw"       // 提币
	LedgerTypeDeposit       LedgerType = "deposit"        // 充币
)

// Ledger 账本流水表（所有余额变化可审计）
type Ledger struct {
	BaseModel
	UserID        uint            `gorm:"index;not null" json:"user_id"`
	AccountType   AccountType     `gorm:"size:16;not null" json:"account_type"`
	LedgerType    LedgerType      `gorm:"size:32;not null;index" json:"ledger_type"`
	Amount        decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"amount"`        // 变动金额（正=增加 负=减少）
	BalanceAfter  decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"balance_after"` // 变动后余额
	RelatedID     *uint           `gorm:"index" json:"related_id"`                          // 关联业务 ID（划转单/结算批次）
	Remark        string          `gorm:"size:256" json:"remark"`
	IdempotentKey string          `gorm:"uniqueIndex;size:64" json:"idempotent_key"` // 幂等键
}

func (Ledger) TableName() string { return "ledgers" }

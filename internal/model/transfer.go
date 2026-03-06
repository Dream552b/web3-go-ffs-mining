package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransferDirection 划转方向
type TransferDirection string

const (
	TransferDirectionIn  TransferDirection = "in"  // 转入挖矿区
	TransferDirectionOut TransferDirection = "out" // 转出到交易区
)

// TransferStatus 划转状态
type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"   // 申请中（冻结资金）
	TransferStatusLocking   TransferStatus = "locking"   // 锁定中（等待24h）
	TransferStatusEffective TransferStatus = "effective" // 已生效（到账 00:00 生效）
	TransferStatusCompleted TransferStatus = "completed" // 已完成
	TransferStatusFailed    TransferStatus = "failed"    // 失败
)

// Transfer 划转申请表
type Transfer struct {
	BaseModel
	UserID        uint              `gorm:"index;not null" json:"user_id"`
	Direction     TransferDirection `gorm:"size:8;not null;index" json:"direction"`
	Amount        decimal.Decimal   `gorm:"type:decimal(30,8);not null" json:"amount"`
	Status        TransferStatus    `gorm:"size:16;not null;index;default:pending" json:"status"`
	ApplyAt       time.Time         `json:"apply_at"`                                  // 申请时间
	EffectiveAt   *time.Time        `json:"effective_at"`                              // 预计生效时间（转入挖矿区专用：24h后次日00:00）
	CompletedAt   *time.Time        `json:"completed_at"`                              // 实际完成时间
	IdempotentKey string            `gorm:"uniqueIndex;size:64" json:"idempotent_key"` // 幂等键
	Remark        string            `gorm:"size:256" json:"remark"`
}

func (Transfer) TableName() string { return "transfers" }

package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// StakeSnapshot 持仓与有效算力日快照
type StakeSnapshot struct {
	BaseModel
	UserID         uint            `gorm:"uniqueIndex:idx_user_date;index;not null" json:"user_id"`
	SnapshotDate   time.Time       `gorm:"uniqueIndex:idx_user_date;type:date;not null;index" json:"snapshot_date"`
	StakeAmount    decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"stake_amount"`    // 当日质押量（挖矿区实际余额）
	EffectivePower decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"effective_power"` // 有效算力 = min(max(stake, min_stake), max_stake)
	IsValid        bool            `gorm:"default:false;index" json:"is_valid"`                // 是否满足最低门槛
}

func (StakeSnapshot) TableName() string { return "stake_snapshots" }

// DailySettlement 日产出结算表
type DailySettlement struct {
	BaseModel
	SettleDate     time.Time       `gorm:"uniqueIndex:idx_settle_user;type:date;not null;index" json:"settle_date"`
	UserID         uint            `gorm:"uniqueIndex:idx_settle_user;index;not null" json:"user_id"`
	StaticReward   decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"static_reward"`   // 静态奖励
	DynamicReward  decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"dynamic_reward"`  // 动态奖励
	TotalReward    decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"total_reward"`    // 合计奖励
	EffectivePower decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"effective_power"` // 当日有效算力
	NetworkPower   decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"network_power"`   // 全网有效算力
	Status         string          `gorm:"size:16;default:pending;index" json:"status"`         // pending | settled | paid
	BatchID        uint            `gorm:"index" json:"batch_id"`
}

func (DailySettlement) TableName() string { return "daily_settlements" }

// SettleBatch 结算批次表
type SettleBatch struct {
	BaseModel
	BatchDate         time.Time       `gorm:"uniqueIndex;type:date;not null" json:"batch_date"`
	TotalOutput       decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"total_output"`        // 当日总产出
	StaticPool        decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"static_pool"`         // 静态池
	DynamicPool       decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"dynamic_pool"`        // 动态池
	TotalNetworkPower decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"total_network_power"` // 全网有效算力
	ValidUserCount    int             `gorm:"not null" json:"valid_user_count"`                       // 有效用户数
	ActualPaid        decimal.Decimal `gorm:"type:decimal(30,8);default:0" json:"actual_paid"`        // 实际发放总额
	Status            string          `gorm:"size:16;default:pending;index" json:"status"`            // pending | running | completed | failed
	Remark            string          `gorm:"size:256" json:"remark"`
}

func (SettleBatch) TableName() string { return "settle_batches" }

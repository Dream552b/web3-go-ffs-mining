package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// MiningPeriodConfig 产出配置表（每月一条记录）
type MiningPeriodConfig struct {
	BaseModel
	PeriodMonth        string          `gorm:"uniqueIndex;size:8;not null" json:"period_month"` // 格式：2024-01
	InitialCirculation decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"initial_circulation"` // 当月初始流通量
	MonthlyOutput      decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"monthly_output"`      // 当月总产出
	DailyOutput        decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"daily_output"`        // 每日产出（总产出/当月天数）
	StaticRatio        decimal.Decimal `gorm:"type:decimal(5,4);not null" json:"static_ratio"`         // 静态比例
	DynamicRatio       decimal.Decimal `gorm:"type:decimal(5,4);not null" json:"dynamic_ratio"`        // 动态比例
	MinStake           decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"min_stake"`           // 最低有效质押
	MaxStake           decimal.Decimal `gorm:"type:decimal(30,8);not null" json:"max_stake"`           // 有效算力上限
	TotalDays          int             `gorm:"not null" json:"total_days"`                             // 当月天数
	EffectiveDate      time.Time       `json:"effective_date"`                                         // 该配置生效日期
	IsActive           bool            `gorm:"default:false;index" json:"is_active"`                   // 当前是否激活
}

func (MiningPeriodConfig) TableName() string { return "mining_period_configs" }

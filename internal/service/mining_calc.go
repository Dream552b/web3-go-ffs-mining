package service

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// CalcEffectivePower 计算有效算力
// effective = min(max(stake, minStake), maxStake)
// 如果 stake < minStake 则无效，返回 0 和 false
func CalcEffectivePower(stake, minStake, maxStake decimal.Decimal) (power decimal.Decimal, valid bool) {
	if stake.LessThan(minStake) {
		return decimal.Zero, false
	}
	if stake.GreaterThan(maxStake) {
		return maxStake, true
	}
	return stake, true
}

// CalcMonthlyOutput 按复利模型计算第 N 个月的产出参数
// month 从 1 开始，initialCirculation 为第 1 个月初始流通量
func CalcMonthlyOutput(month int, initialCirculation, growthRate decimal.Decimal) (circulation, output decimal.Decimal) {
	// 当月初始流通 = initialCirculation * (1 + growthRate)^(month-1)
	multiplier := decimal.NewFromInt(1).Add(growthRate).Pow(decimal.NewFromInt(int64(month - 1)))
	circulation = initialCirculation.Mul(multiplier).Round(8)
	output = circulation.Mul(growthRate).Round(8)
	return
}

// CalcDailyOutput 计算每日产出
func CalcDailyOutput(monthlyOutput decimal.Decimal, days int) decimal.Decimal {
	if days <= 0 {
		days = 30
	}
	return monthlyOutput.Div(decimal.NewFromInt(int64(days))).Round(8)
}

// CalcStaticReward 计算单个用户的静态奖励
// reward = (effectivePower / totalNetworkPower) * staticPool
func CalcStaticReward(effectivePower, totalNetworkPower, staticPool decimal.Decimal) decimal.Decimal {
	if totalNetworkPower.IsZero() {
		return decimal.Zero
	}
	return effectivePower.Div(totalNetworkPower).Mul(staticPool).Round(8)
}

// CalcTransferEffectiveTime 计算划转转入挖矿区的生效时间
// 规则：申请时间 + 24小时 后的次日 00:00
func CalcTransferEffectiveTime(applyAt time.Time, delayHours int) time.Time {
	// 加上延迟小时数
	after24h := applyAt.Add(time.Duration(delayHours) * time.Hour)
	// 取次日 00:00（本地时区）
	y, m, d := after24h.Date()
	nextDay := time.Date(y, m, d+1, 0, 0, 0, 0, after24h.Location())
	return nextDay
}

// GetDaysInMonth 获取某月天数
func GetDaysInMonth(t time.Time) int {
	firstOfNext := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	return int(firstOfNext.Sub(time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())).Hours() / 24)
}

// GetPeriodMonth 获取月份字符串，格式 2024-01
func GetPeriodMonth(t time.Time) string {
	return fmt.Sprintf("%d-%02d", t.Year(), int(t.Month()))
}

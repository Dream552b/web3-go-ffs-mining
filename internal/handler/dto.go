package handler

// ===================== 统一响应 =====================

// ResponseOK 成功响应
type ResponseOK struct {
	Code    int    `json:"code" example:"0"`
	Message string `json:"message" example:"成功"`
	Data    any    `json:"data,omitempty"`
}

// ResponseErr 错误响应
type ResponseErr struct {
	Code    int    `json:"code" example:"10001"`
	Message string `json:"message" example:"请求参数错误"`
}

// ===================== 用户 =====================

// UserInfo 用户信息
type UserInfo struct {
	ID         uint   `json:"id" example:"1"`
	Username   string `json:"username" example:"alice"`
	Email      string `json:"email" example:"alice@example.com"`
	Phone      string `json:"phone" example:"13800138000"`
	Role       string `json:"role" example:"user"`
	InviteCode string `json:"invite_code" example:"a1b2c3d4"`
}

// RegisterRespData 注册成功响应数据
type RegisterRespData struct {
	UserID     uint   `json:"user_id" example:"1"`
	Username   string `json:"username" example:"alice"`
	InviteCode string `json:"invite_code" example:"a1b2c3d4"`
}

// LoginRespData 登录成功响应数据
type LoginRespData struct {
	Token string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserInfo `json:"user"`
}

// ===================== 资产 =====================

// AccountDetail 单个账户详情
type AccountDetail struct {
	Balance   string `json:"balance" example:"1000.00000000"`
	Frozen    string `json:"frozen" example:"0.00000000"`
	Available string `json:"available" example:"1000.00000000"`
}

// MiningAccountDetail 挖矿区账户详情
type MiningAccountDetail struct {
	Balance        string `json:"balance" example:"500.00000000"`
	Frozen         string `json:"frozen" example:"0.00000000"`
	Available      string `json:"available" example:"500.00000000"`
	EffectivePower string `json:"effective_power" example:"500.00000000"`
	IsValidMiner   bool   `json:"is_valid_miner" example:"true"`
}

// MiningRules 挖矿规则
type MiningRules struct {
	MinStake float64 `json:"min_stake" example:"100"`
	MaxStake float64 `json:"max_stake" example:"2000"`
}

// AssetsRespData 资产查询响应数据
type AssetsRespData struct {
	Trade       AccountDetail       `json:"trade"`
	Mining      MiningAccountDetail `json:"mining"`
	MiningRules MiningRules         `json:"mining_rules"`
}

// ===================== 划转 =====================

// TransferRespData 划转响应数据
type TransferRespData struct {
	TransferID  uint    `json:"transfer_id" example:"10"`
	Status      string  `json:"status" example:"locking"`
	Amount      string  `json:"amount" example:"200.00000000"`
	EffectiveAt *string `json:"effective_at" example:"2024-01-02T00:00:00+08:00"`
}

// ===================== 挖矿奖励 =====================

// SettlementItem 单条结算记录
type SettlementItem struct {
	ID             uint   `json:"id" example:"1"`
	SettleDate     string `json:"settle_date" example:"2024-01-01"`
	StaticReward   string `json:"static_reward" example:"1.23456789"`
	DynamicReward  string `json:"dynamic_reward" example:"0.00000000"`
	TotalReward    string `json:"total_reward" example:"1.23456789"`
	EffectivePower string `json:"effective_power" example:"500.00000000"`
	NetworkPower   string `json:"network_power" example:"100000.00000000"`
	Status         string `json:"status" example:"paid"`
}

// RewardsRespData 奖励查询响应数据
type RewardsRespData struct {
	Total       int64            `json:"total" example:"30"`
	Page        int              `json:"page" example:"1"`
	Size        int              `json:"size" example:"20"`
	List        []SettlementItem `json:"list"`
	TotalReward string           `json:"total_reward" example:"36.98765432"`
}

// ===================== 全网看板 =====================

// NetworkStatsData 全网统计数据
type NetworkStatsData struct {
	TotalStake     string `json:"total_stake" example:"5000000.00000000"`
	TotalEffective string `json:"total_effective" example:"3000000.00000000"`
	ValidUsers     int    `json:"valid_users" example:"1500"`
}

// BatchInfo 结算批次信息
type BatchInfo struct {
	TotalOutput string `json:"total_output" example:"33000.00000000"`
	StaticPool  string `json:"static_pool" example:"16500.00000000"`
	DynamicPool string `json:"dynamic_pool" example:"16500.00000000"`
	ActualPaid  string `json:"actual_paid" example:"16499.98765432"`
	ValidUsers  int    `json:"valid_users" example:"1500"`
	Status      string `json:"status" example:"completed"`
}

// PeriodConfigInfo 当前周期配置
type PeriodConfigInfo struct {
	PeriodMonth   string `json:"period_month" example:"2024-01"`
	DailyOutput   string `json:"daily_output" example:"33333.33333333"`
	MonthlyOutput string `json:"monthly_output" example:"1000000.00000000"`
	StaticRatio   string `json:"static_ratio" example:"0.5000"`
	DynamicRatio  string `json:"dynamic_ratio" example:"0.5000"`
	MinStake      string `json:"min_stake" example:"100.00000000"`
	MaxStake      string `json:"max_stake" example:"2000.00000000"`
}

// NetworkBoardRespData 全网看板响应数据
type NetworkBoardRespData struct {
	YesterdayDate string           `json:"yesterday_date" example:"2024-01-01"`
	NetworkStats  NetworkStatsData `json:"network_stats"`
	YesterdayBatch BatchInfo       `json:"yesterday_batch"`
	CurrentConfig  PeriodConfigInfo `json:"current_config"`
}

// ===================== 管理员 =====================

// AdminSettleRespData 管理员结算响应
type AdminSettleRespData struct {
	Code    int    `json:"code" example:"0"`
	Message string `json:"message" example:"结算执行成功"`
}

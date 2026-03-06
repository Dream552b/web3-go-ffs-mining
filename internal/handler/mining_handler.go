package handler

import (
	"fss-mining/internal/middleware"
	"fss-mining/internal/repository"
	"fss-mining/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type MiningHandler struct {
	miningRepo *repository.MiningRepo
	ledgerRepo *repository.LedgerRepo
}

func NewMiningHandler(mr *repository.MiningRepo, lr *repository.LedgerRepo) *MiningHandler {
	return &MiningHandler{miningRepo: mr, ledgerRepo: lr}
}

// GetUserRewards godoc
// @Summary      查询个人挖矿奖励记录
// @Description  分页查询当前用户的每日挖矿奖励明细，包含静态奖励、动态奖励、当日有效算力及全网算力
// @Tags         挖矿
// @Security     BearerAuth
// @Produce      json
// @Param        page  query     int  false  "页码（默认1）"  default(1)
// @Param        size  query     int  false  "每页数量（默认20，最大100）"  default(20)
// @Success      200   {object}  ResponseOK{data=RewardsRespData}  "奖励记录列表"
// @Failure      200   {object}  ResponseErr                       "未授权"
// @Router       /mining/rewards [get]
func (h *MiningHandler) GetUserRewards(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	ctx := c.Request.Context()
	list, total, err := h.miningRepo.GetUserSettlements(ctx, userID, page, size)
	if err != nil {
		response.ServerError(c)
		return
	}

	totalReward, err := h.miningRepo.GetUserTotalReward(ctx, userID)
	if err != nil {
		response.ServerError(c)
		return
	}

	response.Success(c, gin.H{
		"total":        total,
		"page":         page,
		"size":         size,
		"list":         list,
		"total_reward": totalReward,
	})
}

// GetNetworkBoard godoc
// @Summary      全网挖矿看板（公开）
// @Description  返回全网质押规模、有效算力、参与用户数、昨日产出及当前挖矿周期配置，无需登录可访问
// @Tags         挖矿
// @Produce      json
// @Success      200  {object}  ResponseOK{data=NetworkBoardRespData}  "全网看板数据"
// @Router       /mining/network [get]
func (h *MiningHandler) GetNetworkBoard(c *gin.Context) {
	ctx := c.Request.Context()
	yesterday := time.Now().AddDate(0, 0, -1)

	stats, err := h.miningRepo.GetNetworkStats(ctx, yesterday)
	if err != nil {
		response.ServerError(c)
		return
	}

	batch, err := h.miningRepo.GetBatchByDate(ctx, yesterday)
	if err != nil {
		response.ServerError(c)
		return
	}

	cfg, err := h.miningRepo.GetActiveConfig(ctx)
	if err != nil {
		response.ServerError(c)
		return
	}

	result := gin.H{
		"network_stats":  stats,
		"yesterday_date": yesterday.Format("2006-01-02"),
	}

	if batch != nil {
		result["yesterday_batch"] = gin.H{
			"total_output": batch.TotalOutput,
			"static_pool":  batch.StaticPool,
			"dynamic_pool": batch.DynamicPool,
			"actual_paid":  batch.ActualPaid,
			"valid_users":  batch.ValidUserCount,
			"status":       batch.Status,
		}
	}

	if cfg != nil {
		result["current_config"] = gin.H{
			"period_month":   cfg.PeriodMonth,
			"daily_output":   cfg.DailyOutput,
			"monthly_output": cfg.MonthlyOutput,
			"static_ratio":   cfg.StaticRatio,
			"dynamic_ratio":  cfg.DynamicRatio,
			"min_stake":      cfg.MinStake,
			"max_stake":      cfg.MaxStake,
		}
	}

	response.Success(c, result)
}

// GetCurrentPeriod godoc
// @Summary      查询当前挖矿周期配置（公开）
// @Description  返回当前激活的月度挖矿配置，包含每日产出、月产出、静态/动态比例、质押门槛等
// @Tags         挖矿
// @Produce      json
// @Success      200  {object}  ResponseOK{data=PeriodConfigInfo}  "当前周期配置"
// @Failure      200  {object}  ResponseErr                        "配置不存在"
// @Router       /mining/period [get]
func (h *MiningHandler) GetCurrentPeriod(c *gin.Context) {
	cfg, err := h.miningRepo.GetActiveConfig(c.Request.Context())
	if err != nil {
		response.ServerError(c)
		return
	}
	if cfg == nil {
		response.Fail(c, response.CodeNotFound)
		return
	}
	response.Success(c, cfg)
}

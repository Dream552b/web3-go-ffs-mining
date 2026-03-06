package handler

import (
	"fss-mining/internal/middleware"
	"fss-mining/internal/model"
	"fss-mining/internal/repository"
	"fss-mining/internal/service"
	"fss-mining/pkg/config"
	"fss-mining/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type AccountHandler struct {
	accountRepo *repository.AccountRepo
	miningRepo  *repository.MiningRepo
}

func NewAccountHandler(ar *repository.AccountRepo, mr *repository.MiningRepo) *AccountHandler {
	return &AccountHandler{accountRepo: ar, miningRepo: mr}
}

// GetAssets godoc
// @Summary      查询用户资产
// @Description  返回当前用户的交易区余额、挖矿区余额及有效算力。有效算力规则：min(max(质押量, 100), 2000)
// @Tags         资产
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ResponseOK{data=AssetsRespData}  "资产信息"
// @Failure      200  {object}  ResponseErr                      "未授权"
// @Router       /account/assets [get]
func (h *AccountHandler) GetAssets(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx := c.Request.Context()

	tradeAcc, err := h.accountRepo.GetOrCreate(ctx, userID, model.AccountTypeTrade)
	if err != nil {
		response.ServerError(c)
		return
	}

	miningAcc, err := h.accountRepo.GetOrCreate(ctx, userID, model.AccountTypeMining)
	if err != nil {
		response.ServerError(c)
		return
	}

	cfg := config.Get().Mining
	minStake := decimal.NewFromFloat(cfg.MinStake)
	maxStake := decimal.NewFromFloat(cfg.MaxStake)
	effectivePower, isValid := service.CalcEffectivePower(miningAcc.Balance, minStake, maxStake)

	response.Success(c, gin.H{
		"trade": gin.H{
			"balance":   tradeAcc.Balance,
			"frozen":    tradeAcc.FrozenAmt,
			"available": tradeAcc.AvailableBalance(),
		},
		"mining": gin.H{
			"balance":         miningAcc.Balance,
			"frozen":          miningAcc.FrozenAmt,
			"available":       miningAcc.AvailableBalance(),
			"effective_power": effectivePower,
			"is_valid_miner":  isValid,
		},
		"mining_rules": gin.H{
			"min_stake": cfg.MinStake,
			"max_stake": cfg.MaxStake,
		},
	})
}

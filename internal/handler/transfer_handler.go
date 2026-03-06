package handler

import (
	"fmt"
	"fss-mining/internal/middleware"
	"fss-mining/internal/service"
	"fss-mining/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type TransferHandlerV2 struct {
	svc *service.TransferService
}

func NewTransferHandler(svc *service.TransferService) *TransferHandlerV2 {
	return &TransferHandlerV2{svc: svc}
}

// TransferIn godoc
// @Summary      从交易区转入挖矿区
// @Description  将交易区资产划转到挖矿区。划转后资金冻结，24小时后的次日00:00正式生效参与挖矿。支持幂等键防止重复提交。
// @Tags         划转
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      TransferInReq                      true  "划转信息"
// @Success      200   {object}  ResponseOK{data=TransferRespData}  "申请成功，返回划转单信息"
// @Failure      200   {object}  ResponseErr                        "余额不足 / 低于最低门槛"
// @Router       /transfer/in [post]
func (h *TransferHandlerV2) TransferIn(c *gin.Context) {
	var req TransferInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.IsNegative() || amount.IsZero() {
		response.ParamError(c, "划转金额无效")
		return
	}

	idempotentKey := fmt.Sprintf("ti-%d-%s-%d", userID, req.Amount, time.Now().UnixMilli())
	if req.IdempotentKey != "" {
		idempotentKey = req.IdempotentKey
	}

	transfer, err := h.svc.TransferIn(c.Request.Context(), userID, amount, idempotentKey)
	if err != nil {
		response.FailMsg(c, response.CodeInsufficientBal, err.Error())
		return
	}

	response.SuccessMsg(c, "划转申请已提交", gin.H{
		"transfer_id":  transfer.ID,
		"status":       transfer.Status,
		"amount":       transfer.Amount,
		"effective_at": transfer.EffectiveAt,
	})
}

// TransferOut godoc
// @Summary      从挖矿区转出到交易区
// @Description  将挖矿区资产划转回交易区。资金冻结，24小时后到账。支持幂等键防止重复提交。
// @Tags         划转
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      TransferOutReq                     true  "划转信息"
// @Success      200   {object}  ResponseOK{data=TransferRespData}  "申请成功，返回划转单信息"
// @Failure      200   {object}  ResponseErr                        "余额不足"
// @Router       /transfer/out [post]
func (h *TransferHandlerV2) TransferOut(c *gin.Context) {
	var req TransferOutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.IsNegative() || amount.IsZero() {
		response.ParamError(c, "划转金额无效")
		return
	}

	idempotentKey := fmt.Sprintf("to-%d-%s-%d", userID, req.Amount, time.Now().UnixMilli())
	if req.IdempotentKey != "" {
		idempotentKey = req.IdempotentKey
	}

	transfer, err := h.svc.TransferOut(c.Request.Context(), userID, amount, idempotentKey)
	if err != nil {
		response.FailMsg(c, response.CodeInsufficientBal, err.Error())
		return
	}

	response.SuccessMsg(c, "划转申请已提交", gin.H{
		"transfer_id":  transfer.ID,
		"status":       transfer.Status,
		"amount":       transfer.Amount,
		"effective_at": transfer.EffectiveAt,
	})
}

// --- 请求结构体 ---

// TransferInReq 转入挖矿区请求
type TransferInReq struct {
	Amount        string `json:"amount" binding:"required" example:"200.00"`
	IdempotentKey string `json:"idempotent_key" example:"my-unique-key-001"`
}

// TransferOutReq 转出到交易区请求
type TransferOutReq struct {
	Amount        string `json:"amount" binding:"required" example:"100.00"`
	IdempotentKey string `json:"idempotent_key" example:"my-unique-key-002"`
}

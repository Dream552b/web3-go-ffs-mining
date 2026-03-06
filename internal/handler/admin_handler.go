package handler

import (
	"fss-mining/internal/repository"
	"fss-mining/internal/service"
	"fss-mining/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type AdminHandler struct {
	settleService *service.SettleService
	miningRepo    *repository.MiningRepo
}

func NewAdminHandler(ss *service.SettleService, mr *repository.MiningRepo) *AdminHandler {
	return &AdminHandler{settleService: ss, miningRepo: mr}
}

// InitPeriodConfig godoc
// @Summary      [管理] 初始化某月挖矿配置
// @Description  根据月份、月编号（用于复利计算）、初始流通量和增长率生成该月的挖矿参数（每日产出、月产出等），生成后需调用 activate 接口激活
// @Tags         管理员
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      InitPeriodReq  true  "配置信息"
// @Success      200   {object}  ResponseOK     "初始化成功"
// @Failure      200   {object}  ResponseErr    "参数错误 / 配置已存在"
// @Router       /admin/mining/config/init [post]
func (h *AdminHandler) InitPeriodConfig(c *gin.Context) {
	var req InitPeriodReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	t, err := time.Parse("2006-01", req.Month)
	if err != nil {
		response.ParamError(c, "月份格式错误，应为 2006-01")
		return
	}

	initialCirculation, err := decimal.NewFromString(req.InitialCirculation)
	if err != nil {
		response.ParamError(c, "初始流通量格式错误")
		return
	}

	growthRate, err := decimal.NewFromString(req.GrowthRate)
	if err != nil {
		response.ParamError(c, "增长率格式错误")
		return
	}

	if err := h.settleService.InitPeriodConfig(c.Request.Context(), t, initialCirculation, growthRate, req.MonthNumber); err != nil {
		response.FailMsg(c, response.CodeServerError, err.Error())
		return
	}

	response.SuccessMsg(c, "配置初始化成功", nil)
}

// ActivatePeriodConfig godoc
// @Summary      [管理] 激活某月挖矿配置
// @Description  将指定月份的挖矿配置设置为当前激活配置（同时停用其他月份配置），激活后系统使用该配置进行每日结算
// @Tags         管理员
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      ActivatePeriodReq  true  "激活信息"
// @Success      200   {object}  ResponseOK         "激活成功"
// @Failure      200   {object}  ResponseErr        "月份不存在"
// @Router       /admin/mining/config/activate [post]
func (h *AdminHandler) ActivatePeriodConfig(c *gin.Context) {
	var req ActivatePeriodReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	if err := h.miningRepo.ActivateConfig(c.Request.Context(), req.Month); err != nil {
		response.FailMsg(c, response.CodeServerError, err.Error())
		return
	}
	response.SuccessMsg(c, "配置已激活", nil)
}

// ManualSettle godoc
// @Summary      [管理] 手动触发补算
// @Description  对指定日期重新执行静态奖励结算，已完成的批次会被跳过（幂等保护），用于修复漏算或异常场景
// @Tags         管理员
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      ManualSettleReq  true  "结算日期"
// @Success      200   {object}  ResponseOK       "结算成功"
// @Failure      200   {object}  ResponseErr      "日期格式错误 / 结算失败"
// @Router       /admin/mining/settle [post]
func (h *AdminHandler) ManualSettle(c *gin.Context) {
	var req ManualSettleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	settleDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.ParamError(c, "日期格式错误，应为 2006-01-02")
		return
	}

	if err := h.settleService.RunDailySettle(c.Request.Context(), settleDate); err != nil {
		response.FailMsg(c, response.CodeServerError, err.Error())
		return
	}

	response.SuccessMsg(c, "结算执行成功", nil)
}

// ListPeriodConfigs godoc
// @Summary      [管理] 查询所有挖矿周期配置
// @Description  返回当前激活的挖矿周期配置详情
// @Tags         管理员
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ResponseOK{data=PeriodConfigInfo}  "配置列表"
// @Router       /admin/mining/configs [get]
func (h *AdminHandler) ListPeriodConfigs(c *gin.Context) {
	var configs []interface{}
	cfg, err := h.miningRepo.GetActiveConfig(c.Request.Context())
	if err != nil {
		response.ServerError(c)
		return
	}
	if cfg != nil {
		configs = append(configs, cfg)
	}
	response.Success(c, configs)
}

// --- 请求结构体 ---

// InitPeriodReq 初始化月度配置请求
type InitPeriodReq struct {
	Month              string `json:"month" binding:"required" example:"2024-01"`
	MonthNumber        int    `json:"month_number" binding:"required,min=1" example:"1"`
	InitialCirculation string `json:"initial_circulation" binding:"required" example:"10000000"`
	GrowthRate         string `json:"growth_rate" binding:"required" example:"0.1"`
}

// ActivatePeriodReq 激活月度配置请求
type ActivatePeriodReq struct {
	Month string `json:"month" binding:"required" example:"2024-01"`
}

// ManualSettleReq 手动结算请求
type ManualSettleReq struct {
	Date string `json:"date" binding:"required" example:"2024-01-01"`
}

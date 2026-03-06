package handler

import (
	"fss-mining/internal/middleware"
	"fss-mining/internal/service"
	"fss-mining/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register godoc
// @Summary      用户注册
// @Description  注册新用户，系统自动创建交易区和挖矿区账户，并生成唯一邀请码
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterReq              true  "注册信息"
// @Success      200   {object}  ResponseOK{data=RegisterRespData}  "注册成功"
// @Failure      200   {object}  ResponseErr                        "参数错误 / 用户名已存在"
// @Router       /user/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	user, err := h.svc.Register(c.Request.Context(), service.RegisterReq{
		Username:   req.Username,
		Password:   req.Password,
		Email:      req.Email,
		Phone:      req.Phone,
		InviteCode: req.InviteCode,
	})
	if err != nil {
		response.FailMsg(c, response.CodeParamError, err.Error())
		return
	}

	response.SuccessMsg(c, "注册成功", gin.H{
		"user_id":     user.ID,
		"username":    user.Username,
		"invite_code": user.InviteCode,
	})
}

// Login godoc
// @Summary      用户登录
// @Description  使用用户名和密码登录，返回 JWT Token（有效期72小时）
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body  body      LoginReq                        true  "登录信息"
// @Success      200   {object}  ResponseOK{data=LoginRespData}  "登录成功，返回 token"
// @Failure      200   {object}  ResponseErr                     "用户名或密码错误"
// @Router       /user/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		response.FailMsg(c, response.CodeUnauthorized, err.Error())
		return
	}

	response.Success(c, resp)
}

// Profile godoc
// @Summary      获取当前用户信息
// @Description  返回当前登录用户的基本信息
// @Tags         用户
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ResponseOK{data=UserInfo}  "用户信息"
// @Failure      200  {object}  ResponseErr                "未授权"
// @Router       /user/profile [get]
func (h *UserHandler) Profile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.svc.GetProfile(c.Request.Context(), userID)
	if err != nil || user == nil {
		response.Fail(c, response.CodeNotFound)
		return
	}
	response.Success(c, user)
}

// --- 请求结构体 ---

// RegisterReq 注册请求
type RegisterReq struct {
	Username   string `json:"username" binding:"required,min=3,max=32" example:"alice"`
	Password   string `json:"password" binding:"required,min=6,max=64" example:"password123"`
	Email      string `json:"email" binding:"omitempty,email" example:"alice@example.com"`
	Phone      string `json:"phone" binding:"omitempty" example:"13800138000"`
	InviteCode string `json:"invite_code" example:"a1b2c3d4"`
}

// LoginReq 登录请求
type LoginReq struct {
	Username string `json:"username" binding:"required" example:"alice"`
	Password string `json:"password" binding:"required" example:"password123"`
}

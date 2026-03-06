package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 错误码定义
const (
	CodeSuccess          = 0
	CodeParamError       = 10001
	CodeUnauthorized     = 10002
	CodeForbidden        = 10003
	CodeNotFound         = 10004
	CodeServerError      = 10005
	CodeInsufficientBal  = 20001
	CodeTransferPending  = 20002
	CodeStakeTooLow      = 20003
	CodeStakeTooHigh     = 20004
	CodeDuplicateRequest = 20005
)

var codeMsg = map[int]string{
	CodeSuccess:          "成功",
	CodeParamError:       "请求参数错误",
	CodeUnauthorized:     "未授权，请先登录",
	CodeForbidden:        "无权限",
	CodeNotFound:         "资源不存在",
	CodeServerError:      "服务器内部错误",
	CodeInsufficientBal:  "余额不足",
	CodeTransferPending:  "划转处理中",
	CodeStakeTooLow:      "持仓低于最低挖矿门槛",
	CodeStakeTooHigh:     "有效算力已达上限",
	CodeDuplicateRequest: "重复请求",
}

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: codeMsg[CodeSuccess],
		Data:    data,
	})
}

func SuccessMsg(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: msg,
		Data:    data,
	})
}

func Fail(c *gin.Context, code int) {
	msg, ok := codeMsg[code]
	if !ok {
		msg = "未知错误"
	}
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
	})
}

func FailMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
	})
}

func ParamError(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeParamError,
		Message: msg,
	})
}

func ServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    CodeServerError,
		Message: codeMsg[CodeServerError],
	})
}

func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    CodeUnauthorized,
		Message: codeMsg[CodeUnauthorized],
	})
}

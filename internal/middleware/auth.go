package middleware

import (
	"fss-mining/pkg/jwt"
	"fss-mining/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const CtxUserIDKey = "userID"
const CtxUserRoleKey = "userRole"

// Auth JWT 鉴权中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		claims, err := jwt.Parse(token)
		if err != nil {
			response.FailMsg(c, response.CodeUnauthorized, err.Error())
			c.Abort()
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxUserRoleKey, claims.Role)
		c.Next()
	}
}

// AdminOnly 管理员权限中间件，必须在 Auth() 之后使用
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(CtxUserRoleKey)
		if role != "admin" {
			response.Fail(c, response.CodeForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	bearer := c.GetHeader("Authorization")
	if strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}
	// 也支持 query 参数
	return c.Query("token")
}

// GetUserID 从 context 获取当前用户 ID
func GetUserID(c *gin.Context) uint {
	id, _ := c.Get(CtxUserIDKey)
	uid, _ := id.(uint)
	return uid
}

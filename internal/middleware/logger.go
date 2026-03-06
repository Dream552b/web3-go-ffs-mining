package middleware

import (
	"fss-mining/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if len(c.Errors) > 0 {
			logger.Error(c.Errors.ByType(gin.ErrorTypePrivate).String(), fields...)
		} else if status >= 500 {
			logger.Error("服务器错误", fields...)
		} else if status >= 400 {
			logger.Warn("客户端错误", fields...)
		} else {
			logger.Info("请求完成", fields...)
		}
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(nil, func(c *gin.Context, err any) {
		logger.Error("panic recovered", zap.Any("error", err))
		c.AbortWithStatusJSON(500, gin.H{
			"code":    10005,
			"message": "服务器内部错误",
		})
	})
}

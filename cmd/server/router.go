package main

import (
	"fss-mining/internal/handler"
	"fss-mining/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupRouter(
	userH    *handler.UserHandler,
	accountH *handler.AccountHandler,
	transferH *handler.TransferHandlerV2,
	miningH  *handler.MiningHandler,
	adminH   *handler.AdminHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogger())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		// 公开接口（无需鉴权）
		user := api.Group("/user")
		{
			user.POST("/register", userH.Register)
			user.POST("/login", userH.Login)
		}

		// 公开挖矿数据
		mining := api.Group("/mining")
		{
			mining.GET("/network", miningH.GetNetworkBoard)
			mining.GET("/period", miningH.GetCurrentPeriod)
		}

		// 需要登录的接口
		auth := api.Group("", middleware.Auth())
		{
			auth.GET("/user/profile", userH.Profile)

			// 资产
			auth.GET("/account/assets", accountH.GetAssets)

			// 划转
			transfer := auth.Group("/transfer")
			{
				transfer.POST("/in", transferH.TransferIn)
				transfer.POST("/out", transferH.TransferOut)
			}

			// 挖矿奖励
			auth.GET("/mining/rewards", miningH.GetUserRewards)

			// 管理员接口
			admin := auth.Group("/admin", middleware.AdminOnly())
			{
				admin.POST("/mining/config/init", adminH.InitPeriodConfig)
				admin.POST("/mining/config/activate", adminH.ActivatePeriodConfig)
				admin.POST("/mining/settle", adminH.ManualSettle)
				admin.GET("/mining/configs", adminH.ListPeriodConfigs)
			}
		}
	}

	return r
}

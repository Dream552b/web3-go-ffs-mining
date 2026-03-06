// @title           FSS 挖矿系统 API
// @version         1.0
// @description     FSS 代币挖矿系统后端接口文档。支持用户注册登录、资产查询、划转（交易区⇄挖矿区）、挖矿奖励查询、全网看板及管理员操作。
// @termsOfService  http://swagger.io/terms/

// @contact.name   FSS 技术团队
// @contact.email  dev@fss-mining.io

// @license.name  MIT

// @host      localhost:8888
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 格式：Bearer {token}

package main

import (
	"context"
	"flag"
	"fmt"
	_ "fss-mining/docs"
	"fss-mining/internal/handler"
	"fss-mining/internal/model"
	"fss-mining/internal/repository"
	"fss-mining/internal/scheduler"
	"fss-mining/internal/service"
	"fss-mining/pkg/config"
	"fss-mining/pkg/database"
	"fss-mining/pkg/logger"
	pkgredis "fss-mining/pkg/redis"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfgFile := flag.String("config", "./configs/config.yaml", "配置文件路径")
	flag.Parse()

	// 1. 加载配置
	cfg, err := config.Load(*cfgFile)
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 2. 初始化日志
	logger.Init(cfg.Log.Level, cfg.Log.Output, cfg.Log.FilePath)
	logger.Info("FSS 挖矿系统启动中...")

	// 3. 初始化数据库
	if err := database.Init(cfg.Database); err != nil {
		logger.Fatal("数据库连接失败", zap.Error(err))
	}

	// 4. 自动迁移表结构
	db := database.GetDB()
	if err := db.AutoMigrate(
		&model.User{},
		&model.Account{},
		&model.Ledger{},
		&model.Transfer{},
		&model.MiningPeriodConfig{},
		&model.StakeSnapshot{},
		&model.DailySettlement{},
		&model.SettleBatch{},
	); err != nil {
		logger.Fatal("数据库迁移失败", zap.Error(err))
	}
	logger.Info("数据库迁移完成")

	// 5. 初始化 Redis
	if err := pkgredis.Init(cfg.Redis); err != nil {
		logger.Fatal("Redis 连接失败", zap.Error(err))
	}

	// 6. 初始化 Repositories
	userRepo := repository.NewUserRepo(db)
	accountRepo := repository.NewAccountRepo(db)
	transferRepo := repository.NewTransferRepo(db)
	ledgerRepo := repository.NewLedgerRepo(db)
	miningRepo := repository.NewMiningRepo(db)

	// 7. 初始化 Services
	userSvc := service.NewUserService(db, userRepo, accountRepo)
	transferSvc := service.NewTransferService(db, accountRepo, transferRepo, ledgerRepo)
	settleSvc := service.NewSettleService(db, accountRepo, miningRepo, ledgerRepo)

	// 8. 初始化 Handlers
	userHandler := handler.NewUserHandler(userSvc)
	accountHandler := handler.NewAccountHandler(accountRepo, miningRepo)
	transferHandler := handler.NewTransferHandler(transferSvc)
	miningHandler := handler.NewMiningHandler(miningRepo, ledgerRepo)
	adminHandler := handler.NewAdminHandler(settleSvc, miningRepo)

	// 9. 启动定时任务
	sched := scheduler.New(settleSvc, transferSvc)
	sched.Start()
	defer sched.Stop()

	// 10. 配置 Gin
	gin.SetMode(cfg.Server.Mode)
	router := setupRouter(userHandler, accountHandler, transferHandler, miningHandler, adminHandler)

	// 11. 启动 HTTP 服务
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("HTTP 服务启动", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP 服务启动失败", zap.Error(err))
		}
	}()

	// 12. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("服务正在关闭...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务关闭异常", zap.Error(err))
	}
	logger.Info("服务已关闭")
}

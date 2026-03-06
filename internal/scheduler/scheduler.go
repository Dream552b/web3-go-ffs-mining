package scheduler

import (
	"context"
	"fss-mining/internal/service"
	"fss-mining/pkg/logger"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Scheduler struct {
	c              *cron.Cron
	settleService  *service.SettleService
	transferService *service.TransferService
}

func New(ss *service.SettleService, ts *service.TransferService) *Scheduler {
	c := cron.New(cron.WithLocation(time.Local), cron.WithSeconds())
	return &Scheduler{
		c:               c,
		settleService:   ss,
		transferService: ts,
	}
}

func (s *Scheduler) Start() {
	// 每日 00:05:00 生成算力快照（对当日00:00时刻的挖矿区余额做快照）
	s.c.AddFunc("0 5 0 * * *", func() {
		date := time.Now().Truncate(24 * time.Hour) // 今天 00:00
		logger.Info("执行算力快照任务", zap.Time("date", date))
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if err := s.settleService.TakeSnapshot(ctx, date); err != nil {
			logger.Error("算力快照失败", zap.Error(err))
		}
	})

	// 每日 00:30:00 执行前一日结算
	s.c.AddFunc("0 30 0 * * *", func() {
		settleDate := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
		logger.Info("执行日结算任务", zap.Time("date", settleDate))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		if err := s.settleService.RunDailySettle(ctx, settleDate); err != nil {
			logger.Error("日结算失败", zap.Error(err))
		}
	})

	// 每小时扫描划转单，处理已满24h的记录
	s.c.AddFunc("0 0 * * * *", func() {
		logger.Info("扫描划转单")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		s.processTransfers(ctx)
	})

	s.c.Start()
	logger.Info("定时任务调度器已启动")
}

func (s *Scheduler) Stop() {
	s.c.Stop()
	logger.Info("定时任务调度器已停止")
}

// processTransfers 扫描并处理满足条件的划转单
func (s *Scheduler) processTransfers(ctx context.Context) {
	// 转入划转：已过 24h，且已到 effective_at 时间 => 完成入账
	// 转出划转：已过 24h => 完成出账
	// 具体逻辑由 TransferService 中的 CompleteTransferIn/Out 执行
	// 此处委托给 TransferService 的扫描方法

	if err := s.transferService.ScanAndCompleteTransfers(ctx); err != nil {
		logger.Error("扫描划转单失败", zap.Error(err))
	}
}

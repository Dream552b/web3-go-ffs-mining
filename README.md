# FSS 挖矿系统

基于 **Go + Gin** 框架开发的 FSS 代币挖矿后端系统。

## 技术栈

| 组件     | 选型                    |
| -------- | ----------------------- |
| Web 框架 | gin v1.12               |
| ORM      | GORM v2 + MySQL         |
| 缓存     | go-redis v9             |
| 鉴权     | JWT (golang-jwt/jwt/v5) |
| 配置     | viper                   |
| 定时任务 | robfig/cron/v3          |
| 日志     | uber/zap                |
| 精度计算 | shopspring/decimal      |

## 项目结构

```
fss-mining/
├── cmd/server/           # 程序入口（main.go + router.go）
├── internal/
│   ├── handler/          # HTTP 处理层
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据模型
│   ├── middleware/        # Gin 中间件（JWT/日志/恢复）
│   └── scheduler/        # 定时任务
├── pkg/
│   ├── config/           # 配置加载（viper）
│   ├── database/         # MySQL 连接
│   ├── redis/            # Redis 连接
│   ├── jwt/              # JWT 生成/解析
│   ├── logger/           # 日志封装
│   └── response/         # 统一响应结构
├── configs/config.yaml   # 配置文件
└── migrations/           # SQL 参考文件
```

## 快速启动

### 1. 准备数据库

```sql
CREATE DATABASE fss_mining DEFAULT CHARACTER SET utf8mb4;
```

### 2. 修改配置

编辑 `configs/config.yaml`，填写数据库和 Redis 连接信息。

### 3. 启动

```bash
go run ./cmd/server/
# 或指定配置文件
go run ./cmd/server/ -config ./configs/config.yaml
```

### 4. 构建

```bash
go build -o fss-mining-server ./cmd/server/
./fss-mining-server
```

## API 接口

### 公开接口

| 方法 | 路径                   | 说明             |
| ---- | ---------------------- | ---------------- |
| POST | /api/v1/user/register  | 用户注册         |
| POST | /api/v1/user/login     | 用户登录         |
| GET  | /api/v1/mining/network | 全网挖矿看板     |
| GET  | /api/v1/mining/period  | 当前挖矿周期配置 |
| GET  | /health                | 健康检查         |

### 需要登录（Bearer Token）

| 方法 | 路径                   | 说明                      |
| ---- | ---------------------- | ------------------------- |
| GET  | /api/v1/user/profile   | 用户信息                  |
| GET  | /api/v1/account/assets | 查询资产（交易区+挖矿区） |
| POST | /api/v1/transfer/in    | 从交易区转入挖矿区        |
| POST | /api/v1/transfer/out   | 从挖矿区转出到交易区      |
| GET  | /api/v1/mining/rewards | 查询挖矿奖励记录          |

### 管理员接口（需要 admin 角色）

| 方法 | 路径                                 | 说明               |
| ---- | ------------------------------------ | ------------------ |
| POST | /api/v1/admin/mining/config/init     | 初始化某月挖矿配置 |
| POST | /api/v1/admin/mining/config/activate | 激活某月挖矿配置   |
| POST | /api/v1/admin/mining/settle          | 手动触发补算       |
| GET  | /api/v1/admin/mining/configs         | 查询所有配置       |

## 核心业务规则

### 产出增长（复利模型）

```
第 N 月初始流通 = 初始流通 × (1 + 10%)^(N-1)
第 N 月产出     = 第 N 月初始流通 × 10%
每日产出        = 月产出 ÷ 当月天数
静态产出        = 每日产出 × 50%
动态产出        = 每日产出 × 50%（算法待确认）
```

### 有效算力

```
effective_power = clamp(stake, 100, 2000)
- stake < 100 FSS：不参与挖矿
- stake > 2000 FSS：按 2000 FSS 计算
```

### 划转延迟

- **转入挖矿区**：申请后等待 24 小时，然后在**下一个 00:00**生效开始挖矿
- **转出交易区**：申请后 24 小时到账

### 定时任务

| 时间       | 任务                             |
| ---------- | -------------------------------- |
| 每日 00:05 | 生成前日算力快照                 |
| 每日 00:30 | 执行前日静态奖励结算             |
| 每小时整点 | 扫描划转单，处理已满足条件的划转 |

## 数据表

| 表名                  | 说明                     |
| --------------------- | ------------------------ |
| users                 | 用户表                   |
| accounts              | 账户表（交易区/挖矿区）  |
| ledgers               | 账本流水表（全量可审计） |
| transfers             | 划转申请表               |
| mining_period_configs | 月度挖矿配置表           |
| stake_snapshots       | 日算力快照表             |
| daily_settlements     | 日结算明细表             |
| settle_batches        | 结算批次表               |

## 每次修改代码后重新生成文档

swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

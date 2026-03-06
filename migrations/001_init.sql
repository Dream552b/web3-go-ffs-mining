-- FSS 挖矿系统初始化 SQL
-- 使用 GORM AutoMigrate 自动建表，此文件作为参考备份

CREATE DATABASE IF NOT EXISTS fss_mining DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE fss_mining;

-- 以下仅为参考，实际由 GORM AutoMigrate 创建
-- AutoMigrate 会自动创建和更新以下表：
-- users, accounts, ledgers, transfers, mining_period_configs, stake_snapshots, daily_settlements, settle_batches

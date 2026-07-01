package repository

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"family-score/internal/logger"

	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

// Repository 表示数据层，负责 SQLite 连接、迁移和基础初始化。
type Repository struct {
	DB      *sql.DB
	DataDir string
	DBPath  string
}

// New 创建数据层实例并初始化本地 SQLite。
func New(dataDir string) (*Repository, error) {
	logger.L().Info("初始化数据目录", zap.String("data_dir", dataDir))
	for _, dir := range []string{"", "backups", "files", "exports", "logs"} {
		path := filepath.Join(dataDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			logger.L().Error("创建数据目录失败", zap.Error(err), zap.String("path", path))
			return nil, err
		}
	}

	dbPath := filepath.Join(dataDir, "family-score.db")
	logger.L().Info("打开 SQLite 数据库", zap.String("db_path", dbPath))
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		logger.L().Error("打开 SQLite 数据库失败", zap.Error(err), zap.String("db_path", dbPath))
		return nil, err
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; PRAGMA foreign_keys=ON;`); err != nil {
		logger.L().Error("设置 SQLite PRAGMA 失败", zap.Error(err))
		return nil, err
	}

	r := &Repository{DB: db, DataDir: dataDir, DBPath: dbPath}
	if err := r.Migrate(context.Background()); err != nil {
		logger.L().Error("数据库迁移失败", zap.Error(err))
		return nil, err
	}
	logger.L().Info("数据层初始化完成", zap.String("db_path", dbPath))
	return r, nil
}

// Close 关闭数据库连接。
func (r *Repository) Close() error { return r.DB.Close() }

// Migrate 执行本地 SQLite 表结构迁移。
func (r *Repository) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (version VARCHAR(64) PRIMARY KEY, description VARCHAR(256) NOT NULL DEFAULT '', applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS families (id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR(128) NOT NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, family_id INTEGER NOT NULL, child_id INTEGER NULL, display_name VARCHAR(64) NOT NULL, role VARCHAR(32) NOT NULL, login_name VARCHAR(64) NOT NULL UNIQUE, password_hash VARCHAR(256) NOT NULL, enabled INTEGER NOT NULL DEFAULT 1, failed_login_date VARCHAR(10) NOT NULL DEFAULT '', failed_login_count INTEGER NOT NULL DEFAULT 0, locked_until TIMESTAMP NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS children (id INTEGER PRIMARY KEY AUTOINCREMENT, family_id INTEGER NOT NULL, parent_user_id INTEGER NULL, name VARCHAR(64) NOT NULL, age INTEGER NOT NULL, gender VARCHAR(16) NOT NULL, profile_note VARCHAR(512) NOT NULL DEFAULT '', created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS score_accounts (id INTEGER PRIMARY KEY AUTOINCREMENT, child_id INTEGER NOT NULL, base_score INTEGER NOT NULL DEFAULT 100, bonus_score INTEGER NOT NULL DEFAULT 0, star_count INTEGER NOT NULL DEFAULT 0, team_score INTEGER NOT NULL DEFAULT 0, status_level VARCHAR(32) NOT NULL DEFAULT 'GREEN', current_month VARCHAR(7) NOT NULL, last_exchange_date DATE NULL, appeal_count_this_week INTEGER NOT NULL DEFAULT 0, version INTEGER NOT NULL DEFAULT 0, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(child_id, current_month));`,
		`CREATE TABLE IF NOT EXISTS score_records (id INTEGER PRIMARY KEY AUTOINCREMENT, child_id INTEGER NOT NULL, account_id INTEGER NOT NULL, record_type VARCHAR(32) NOT NULL, target_account VARCHAR(32) NOT NULL, item_name VARCHAR(128) NOT NULL, score_change INTEGER NOT NULL, before_value INTEGER NOT NULL, after_value INTEGER NOT NULL, operator_role VARCHAR(32) NOT NULL, operator_id INTEGER NULL, reason VARCHAR(512) NOT NULL, evidence VARCHAR(1024) NOT NULL DEFAULT '', confirm_status VARCHAR(32) NOT NULL DEFAULT 'CONFIRMED', occurred_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE INDEX IF NOT EXISTS idx_score_records_child_time ON score_records(child_id, occurred_at);`,
		`CREATE TABLE IF NOT EXISTS task_templates (id INTEGER PRIMARY KEY AUTOINCREMENT, family_id INTEGER NOT NULL, task_name VARCHAR(128) NOT NULL, task_type VARCHAR(32) NOT NULL, category VARCHAR(32) NOT NULL, score_value INTEGER NOT NULL, target_account VARCHAR(32) NOT NULL, need_parent_confirm INTEGER NOT NULL DEFAULT 1, daily_limit INTEGER NOT NULL DEFAULT 1, weekly_limit INTEGER NOT NULL DEFAULT 0, enabled INTEGER NOT NULL DEFAULT 1, description VARCHAR(512) NOT NULL DEFAULT '', created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS task_instances (id INTEGER PRIMARY KEY AUTOINCREMENT, child_id INTEGER NOT NULL, template_id INTEGER NULL, task_name VARCHAR(128) NOT NULL, task_type VARCHAR(32) NOT NULL, category VARCHAR(32) NOT NULL, score_value INTEGER NOT NULL, target_account VARCHAR(32) NOT NULL, status VARCHAR(32) NOT NULL DEFAULT 'TODO', submit_note VARCHAR(512) NOT NULL DEFAULT '', audit_note VARCHAR(512) NOT NULL DEFAULT '', task_date DATE NOT NULL, submitted_at TIMESTAMP NULL, audited_at TIMESTAMP NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE INDEX IF NOT EXISTS idx_task_instances_child_date ON task_instances(child_id, task_date);`,
		`CREATE TABLE IF NOT EXISTS rewards (id INTEGER PRIMARY KEY AUTOINCREMENT, family_id INTEGER NOT NULL, reward_name VARCHAR(128) NOT NULL, reward_type VARCHAR(32) NOT NULL, cost_score INTEGER NOT NULL DEFAULT 0, cost_star INTEGER NOT NULL DEFAULT 0, weekly_limit INTEGER NOT NULL DEFAULT 0, monthly_limit INTEGER NOT NULL DEFAULT 0, health_risk VARCHAR(32) NOT NULL DEFAULT 'NONE', need_parent_confirm INTEGER NOT NULL DEFAULT 1, enabled INTEGER NOT NULL DEFAULT 1, description VARCHAR(512) NOT NULL DEFAULT '', created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS exchange_orders (id INTEGER PRIMARY KEY AUTOINCREMENT, child_id INTEGER NOT NULL, reward_id INTEGER NOT NULL, cost_score INTEGER NOT NULL DEFAULT 0, cost_star INTEGER NOT NULL DEFAULT 0, status VARCHAR(32) NOT NULL DEFAULT 'PENDING', apply_note VARCHAR(512) NOT NULL DEFAULT '', audit_note VARCHAR(512) NOT NULL DEFAULT '', applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, audited_at TIMESTAMP NULL, completed_at TIMESTAMP NULL, created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
		`CREATE TABLE IF NOT EXISTS appeals (id INTEGER PRIMARY KEY AUTOINCREMENT, child_id INTEGER NOT NULL, record_id INTEGER NOT NULL, appeal_reason VARCHAR(512) NOT NULL, expected_solution VARCHAR(512) NOT NULL DEFAULT '', status VARCHAR(32) NOT NULL DEFAULT 'PENDING', handle_result VARCHAR(32) NULL, handle_note VARCHAR(512) NOT NULL DEFAULT '', created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, handled_at TIMESTAMP NULL);`,
		`CREATE TABLE IF NOT EXISTS weekly_reviews (id INTEGER PRIMARY KEY AUTOINCREMENT, child_id INTEGER NOT NULL, week_start DATE NOT NULL, week_end DATE NOT NULL, add_score INTEGER NOT NULL DEFAULT 0, deduct_score INTEGER NOT NULL DEFAULT 0, repair_score INTEGER NOT NULL DEFAULT 0, task_complete_count INTEGER NOT NULL DEFAULT 0, exchange_count INTEGER NOT NULL DEFAULT 0, summary VARCHAR(1024) NOT NULL DEFAULT '', parent_note VARCHAR(1024) NOT NULL DEFAULT '', created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(child_id, week_start, week_end));`,
		`CREATE TABLE IF NOT EXISTS monthly_settlements (id INTEGER PRIMARY KEY AUTOINCREMENT, child_id INTEGER NOT NULL, month VARCHAR(7) NOT NULL, base_score_before_reset INTEGER NOT NULL, bonus_score_before_clear INTEGER NOT NULL, converted_stars INTEGER NOT NULL DEFAULT 0, cleared_bonus_score INTEGER NOT NULL DEFAULT 0, team_score INTEGER NOT NULL DEFAULT 0, team_level VARCHAR(32) NOT NULL, settlement_status VARCHAR(32) NOT NULL DEFAULT 'DONE', created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(child_id, month));`,
		`CREATE TABLE IF NOT EXISTS family_configs (id INTEGER PRIMARY KEY AUTOINCREMENT, family_id INTEGER NOT NULL, config_key VARCHAR(128) NOT NULL, config_value VARCHAR(256) NOT NULL, description VARCHAR(512) NOT NULL DEFAULT '', created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(family_id, config_key));`,
		`CREATE TABLE IF NOT EXISTS backup_records (id INTEGER PRIMARY KEY AUTOINCREMENT, operation_type VARCHAR(32) NOT NULL, file_path VARCHAR(1024) NOT NULL, file_size INTEGER NOT NULL DEFAULT 0, status VARCHAR(32) NOT NULL, operator_id INTEGER NULL, remark VARCHAR(512) NOT NULL DEFAULT '', created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
	}
	logger.L().Info("开始执行数据库迁移", zap.Int("statement_count", len(stmts)))
	for _, stmt := range stmts {
		if _, err := r.DB.ExecContext(ctx, stmt); err != nil {
			logger.L().Error("执行数据库迁移 SQL 失败", zap.Error(err), zap.String("sql", stmt))
			return err
		}
	}
	if err := addColumnIfMissing(ctx, r.DB, "users", "child_id", "INTEGER NULL"); err != nil {
		return err
	}
	if err := addColumnIfMissing(ctx, r.DB, "children", "parent_user_id", "INTEGER NULL"); err != nil {
		return err
	}
	if err := addColumnIfMissing(ctx, r.DB, "users", "failed_login_date", "VARCHAR(10) NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := addColumnIfMissing(ctx, r.DB, "users", "failed_login_count", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := addColumnIfMissing(ctx, r.DB, "users", "locked_until", "TIMESTAMP NULL"); err != nil {
		return err
	}
	if err := normalizeDefaultTaskScores(ctx, r.DB); err != nil {
		return err
	}
	_, _ = r.DB.ExecContext(ctx, `INSERT OR IGNORE INTO schema_migrations(version, description) VALUES('001_init', 'initial schema')`)
	logger.L().Info("数据库迁移完成")
	return nil
}

func normalizeDefaultTaskScores(ctx context.Context, db *sql.DB) error {
	updates := map[string]int{
		"按时起床穿衣":    1,
		"洗脸刷牙":      1,
		"按时开始并完成作业": 3,
		"阅读 20 分钟":  2,
		"整理书包":      1,
		"收拾书桌 / 玩具": 1,
		"运动 20 分钟":  2,
		"按时睡觉":      2,
		"饭前摆碗筷":     1,
		"主动喝水达标":    1,
	}
	for name, score := range updates {
		if _, err := db.ExecContext(ctx, `UPDATE task_templates SET score_value = ? WHERE task_name = ? AND task_type = 'DAILY'`, score, name); err != nil {
			return err
		}
	}
	_, _ = db.ExecContext(ctx, `INSERT INTO task_templates(family_id, task_name, task_type, category, score_value, target_account, description)
		SELECT id, '饭前摆碗筷', 'DAILY', 'HOUSEWORK', 1, 'AUTO', '主动参与饭前准备' FROM families
		WHERE NOT EXISTS (SELECT 1 FROM task_templates WHERE task_name = '饭前摆碗筷' AND task_type = 'DAILY')`)
	_, _ = db.ExecContext(ctx, `INSERT INTO task_templates(family_id, task_name, task_type, category, score_value, target_account, description)
		SELECT id, '主动喝水达标', 'DAILY', 'HEALTH', 1, 'AUTO', '按约定完成喝水目标' FROM families
		WHERE NOT EXISTS (SELECT 1 FROM task_templates WHERE task_name = '主动喝水达标' AND task_type = 'DAILY')`)
	return nil
}

func addColumnIfMissing(ctx context.Context, db *sql.DB, table string, column string, definition string) error {
	rows, err := db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	_, err = db.ExecContext(ctx, `ALTER TABLE `+table+` ADD COLUMN `+column+` `+definition)
	return err
}

// InsertDefaults 初始化家庭默认配置、任务模板和兑换项。
func InsertDefaults(ctx context.Context, tx *sql.Tx, familyID int64) error {
	configs := map[string]string{"base_score_default": "100", "daily_add_limit": "10", "weekly_exchange_limit": "20", "high_value_block_score": "90", "all_exchange_block_score": "80", "bonus_to_star_rate": "10", "star_limit": "20", "weekly_appeal_limit": "2", "server_host": "127.0.0.1", "server_port": "8080"}
	for k, v := range configs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO family_configs(family_id, config_key, config_value) VALUES(?, ?, ?)`, familyID, k, v); err != nil {
			return err
		}
	}
	tasks := []struct {
		name, typ, cat string
		score          int
		target, desc   string
	}{{"按时起床穿衣", "DAILY", "SELF_CARE", 1, "AUTO", "不反复催促，自己完成穿衣整理"}, {"洗脸刷牙", "DAILY", "SELF_CARE", 1, "AUTO", "早晚洗漱按标准完成"}, {"按时开始并完成作业", "DAILY", "STUDY", 3, "AUTO", "不拖到很晚，态度认真"}, {"阅读 20 分钟", "DAILY", "STUDY", 2, "AUTO", "亲子阅读或自主阅读均可"}, {"整理书包", "DAILY", "SELF_CARE", 1, "AUTO", "第二天物品准备齐全"}, {"收拾书桌 / 玩具", "DAILY", "HOUSEWORK", 1, "AUTO", "玩后归位，书桌整洁"}, {"运动 20 分钟", "DAILY", "HEALTH", 2, "AUTO", "跳绳、跑步、球类等"}, {"按时睡觉", "DAILY", "HEALTH", 2, "AUTO", "睡前不拖延"}, {"饭前摆碗筷", "DAILY", "HOUSEWORK", 1, "AUTO", "主动参与饭前准备"}, {"主动喝水达标", "DAILY", "HEALTH", 1, "AUTO", "按约定完成喝水目标"}, {"整理自己的书桌 / 玩具区", "REPAIR", "SELF_CARE", 2, "BASE", "用于乱扔物品后的责任修复"}, {"给被影响的人道歉并说明补救方式", "REPAIR", "EMOTION", 3, "BASE", "用于冲突后的责任修复"}, {"家庭阅读 / 运动", "TEAM", "HEALTH", 3, "TEAM", "全家共同完成一次阅读或运动"}}
	for _, t := range tasks {
		if _, err := tx.ExecContext(ctx, `INSERT INTO task_templates(family_id, task_name, task_type, category, score_value, target_account, description) VALUES(?, ?, ?, ?, ?, ?, ?)`, familyID, t.name, t.typ, t.cat, t.score, t.target, t.desc); err != nil {
			return err
		}
	}
	rewards := []struct {
		name, typ   string
		score, star int
		risk, desc  string
	}{{"贴纸、卡片、小文具", "STATIONERY", 5, 0, "NONE", "低积分推荐奖励"}, {"小零食一份", "SNACK", 8, 0, "LOW", "控制频率，小份兑换"}, {"选择一次睡前故事", "ACTIVITY", 5, 0, "NONE", "非物质奖励"}, {"亲子桌游 15 分钟", "ACTIVITY", 6, 0, "NONE", "增加陪伴"}, {"自主安排 30 分钟自由时间", "PRIVILEGE", 10, 0, "NONE", "不影响作业和睡眠"}, {"漫画书 / 科普书一本", "BOOK", 20, 0, "NONE", "鼓励阅读"}, {"科技馆、博物馆、动物园活动", "ACTIVITY", 0, 2, "NONE", "月度或季度奖励"}}
	for _, rw := range rewards {
		if _, err := tx.ExecContext(ctx, `INSERT INTO rewards(family_id, reward_name, reward_type, cost_score, cost_star, weekly_limit, monthly_limit, health_risk, description) VALUES(?, ?, ?, ?, ?, 1, 4, ?, ?)`, familyID, rw.name, rw.typ, rw.score, rw.star, rw.risk, rw.desc); err != nil {
			return err
		}
	}
	return nil
}

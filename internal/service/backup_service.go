package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"family-score/internal/logger"
	"family-score/internal/protocol"
	"family-score/pkg/utils"

	"go.uber.org/zap"
)

// Backup 创建本地 SQLite 备份。
func (s *Service) Backup(ctx context.Context, sess protocol.Session) (string, int64, error) {
	logger.L().Info("开始创建本地备份", zap.Int64("operator_id", sess.UserID))
	if !IsAdmin(sess) {
		return "", 0, errors.New("只有管理员可以执行该操作")
	}
	name := fmt.Sprintf("family-score-%s.db", time.Now().Format("20060102-150405"))
	backupPath := filepath.Join(s.repo.DataDir, "backups", name)
	if _, err := s.repo.DB.ExecContext(ctx, `VACUUM INTO `+utils.SQLiteQuote(backupPath)); err != nil {
		s.recordBackup(ctx, "BACKUP", backupPath, 0, "FAILED", sess.UserID, err.Error())
		return "", 0, err
	}
	info, _ := os.Stat(backupPath)
	size := int64(0)
	if info != nil {
		size = info.Size()
	}
	s.recordBackup(ctx, "BACKUP", backupPath, size, "DONE", sess.UserID, "手动备份")
	logger.L().Info("本地备份创建完成", zap.String("backup_path", backupPath), zap.Int64("file_size", size), zap.Int64("operator_id", sess.UserID))
	return backupPath, size, nil
}

// Backups 查询本地备份记录。
func (s *Service) Backups(ctx context.Context, sess protocol.Session) ([]protocol.BackupRecord, error) {
	if !IsAdmin(sess) {
		return nil, errors.New("只有管理员可以查看备份记录")
	}
	rows, err := s.repo.DB.QueryContext(ctx, `SELECT id, operation_type, file_path, file_size, status, remark, created_at FROM backup_records ORDER BY id DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.BackupRecord{}
	for rows.Next() {
		var it protocol.BackupRecord
		if err := rows.Scan(&it.ID, &it.OperationType, &it.FilePath, &it.FileSize, &it.Status, &it.Remark, &it.CreatedAt); err == nil {
			items = append(items, it)
		}
	}
	return items, nil
}

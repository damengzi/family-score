package service

import "context"

func (s *Service) recordBackup(ctx context.Context, typ, path string, size int64, status string, operatorID int64, remark string) {
	_, _ = s.repo.DB.ExecContext(ctx, `INSERT INTO backup_records(operation_type, file_path, file_size, status, operator_id, remark) VALUES(?, ?, ?, ?, ?, ?)`, typ, path, size, status, operatorID, remark)
}

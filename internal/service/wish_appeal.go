package service

import (
	"context"
	"errors"
	"strings"

	"family-score/internal/consts"
	"family-score/internal/logger"
	"family-score/internal/protocol"

	"go.uber.org/zap"
)

// Wishes 查询当前用户可见的愿望列表。
func (s *Service) Wishes(ctx context.Context, sess protocol.Session) ([]protocol.Wish, error) {
	query := `SELECT w.id, w.child_id, w.wish_name, w.wish_type, w.expected_score, w.expected_star, w.reason, w.status, w.audit_note, w.created_at, COALESCE(w.audited_at, '') FROM wishes w JOIN children c ON c.id = w.child_id WHERE c.family_id = ? ORDER BY w.id DESC LIMIT 200`
	args := []any{sess.FamilyID}
	if sess.Role == consts.RoleChild {
		query = `SELECT w.id, w.child_id, w.wish_name, w.wish_type, w.expected_score, w.expected_star, w.reason, w.status, w.audit_note, w.created_at, COALESCE(w.audited_at, '') FROM wishes w JOIN children c ON c.id = w.child_id WHERE c.family_id = ? AND w.child_id = ? ORDER BY w.id DESC LIMIT 200`
		args = append(args, sess.ChildID)
	}
	rows, err := s.repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.Wish{}
	for rows.Next() {
		var it protocol.Wish
		if err := rows.Scan(&it.ID, &it.ChildID, &it.WishName, &it.WishType, &it.ExpectedScore, &it.ExpectedStar, &it.Reason, &it.Status, &it.AuditNote, &it.CreatedAt, &it.AuditedAt); err == nil && s.CanAccessChild(ctx, sess, it.ChildID) {
			items = append(items, it)
		}
	}
	return items, nil
}

// CreateWish 创建孩子愿望申请。
func (s *Service) CreateWish(ctx context.Context, sess protocol.Session, req protocol.CreateWishParam) (int64, error) {
	if req.ChildID <= 0 || !s.CanAccessChild(ctx, sess, req.ChildID) {
		return 0, errors.New("无权访问该孩子档案")
	}
	req.WishName = strings.TrimSpace(req.WishName)
	req.WishType = strings.ToUpper(strings.TrimSpace(req.WishType))
	req.Reason = strings.TrimSpace(req.Reason)
	if req.WishName == "" {
		return 0, errors.New("愿望名称不能为空")
	}
	if req.WishType == "" {
		req.WishType = "REWARD"
	}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO wishes(child_id, wish_name, wish_type, expected_score, expected_star, reason) VALUES(?, ?, ?, ?, ?, ?)`, req.ChildID, req.WishName, req.WishType, req.ExpectedScore, req.ExpectedStar, req.Reason)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	logger.L().Info("新增愿望", zap.Int64("wish_id", id), zap.Int64("child_id", req.ChildID), zap.Int64("operator_id", sess.UserID))
	return id, nil
}

// AuditWish 审批孩子愿望。
func (s *Service) AuditWish(ctx context.Context, sess protocol.Session, wishID int64, req protocol.AuditWishParam) error {
	if !CanOperate(sess) {
		return errors.New("只有管理员或家长可以审批愿望")
	}
	var childID int64
	var status string
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT child_id, status FROM wishes WHERE id = ?`, wishID).Scan(&childID, &status); err != nil {
		return errors.New("愿望不存在")
	}
	if !s.CanAccessChild(ctx, sess, childID) {
		return errors.New("无权审批该愿望")
	}
	if status != "PENDING" {
		return errors.New("愿望状态不可审批")
	}
	result := strings.ToUpper(strings.TrimSpace(req.Result))
	if result != "APPROVED" {
		result = "REJECTED"
	}
	_, err := s.repo.DB.ExecContext(ctx, `UPDATE wishes SET status = ?, audit_note = ?, audited_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, result, req.AuditNote, wishID)
	return err
}

// Appeals 查询当前用户可见的申诉列表。
func (s *Service) Appeals(ctx context.Context, sess protocol.Session) ([]protocol.Appeal, error) {
	query := `SELECT a.id, a.child_id, COALESCE(a.target_type, 'SCORE'), COALESCE(a.target_id, 0), a.record_id, a.appeal_reason, a.expected_solution, a.status, COALESCE(a.handle_result, ''), a.handle_note, a.created_at, COALESCE(a.handled_at, '') FROM appeals a JOIN children c ON c.id = a.child_id WHERE c.family_id = ? ORDER BY a.id DESC LIMIT 200`
	args := []any{sess.FamilyID}
	if sess.Role == consts.RoleChild {
		query = `SELECT a.id, a.child_id, COALESCE(a.target_type, 'SCORE'), COALESCE(a.target_id, 0), a.record_id, a.appeal_reason, a.expected_solution, a.status, COALESCE(a.handle_result, ''), a.handle_note, a.created_at, COALESCE(a.handled_at, '') FROM appeals a JOIN children c ON c.id = a.child_id WHERE c.family_id = ? AND a.child_id = ? ORDER BY a.id DESC LIMIT 200`
		args = append(args, sess.ChildID)
	}
	rows, err := s.repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.Appeal{}
	for rows.Next() {
		var it protocol.Appeal
		if err := rows.Scan(&it.ID, &it.ChildID, &it.TargetType, &it.TargetID, &it.RecordID, &it.AppealReason, &it.ExpectedSolution, &it.Status, &it.HandleResult, &it.HandleNote, &it.CreatedAt, &it.HandledAt); err == nil && s.CanAccessChild(ctx, sess, it.ChildID) {
			items = append(items, it)
		}
	}
	return items, nil
}

// CreateAppeal 创建申诉。
func (s *Service) CreateAppeal(ctx context.Context, sess protocol.Session, req protocol.CreateAppealParam) (int64, error) {
	if req.ChildID <= 0 || !s.CanAccessChild(ctx, sess, req.ChildID) {
		return 0, errors.New("无权访问该孩子档案")
	}
	req.TargetType = strings.ToUpper(strings.TrimSpace(req.TargetType))
	if req.TargetType == "" {
		req.TargetType = "TASK"
	}
	req.AppealReason = strings.TrimSpace(req.AppealReason)
	req.ExpectedSolution = strings.TrimSpace(req.ExpectedSolution)
	if req.AppealReason == "" {
		return 0, errors.New("申诉原因不能为空")
	}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO appeals(child_id, target_type, target_id, record_id, appeal_reason, expected_solution) VALUES(?, ?, ?, ?, ?, ?)`, req.ChildID, req.TargetType, req.TargetID, req.RecordID, req.AppealReason, req.ExpectedSolution)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	logger.L().Info("新增申诉", zap.Int64("appeal_id", id), zap.Int64("child_id", req.ChildID), zap.String("target_type", req.TargetType), zap.Int64("operator_id", sess.UserID))
	return id, nil
}

// HandleAppeal 处理申诉。
func (s *Service) HandleAppeal(ctx context.Context, sess protocol.Session, appealID int64, req protocol.HandleAppealParam) error {
	if !CanOperate(sess) {
		return errors.New("只有管理员或家长可以处理申诉")
	}
	var item protocol.Appeal
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT id, child_id, COALESCE(target_type, 'SCORE'), COALESCE(target_id, 0), record_id, status FROM appeals WHERE id = ?`, appealID).Scan(&item.ID, &item.ChildID, &item.TargetType, &item.TargetID, &item.RecordID, &item.Status); err != nil {
		return errors.New("申诉不存在")
	}
	if !s.CanAccessChild(ctx, sess, item.ChildID) {
		return errors.New("无权处理该申诉")
	}
	if item.Status != "PENDING" {
		return errors.New("申诉状态不可处理")
	}
	result := strings.ToUpper(strings.TrimSpace(req.Result))
	if result != "APPROVED" {
		result = "REJECTED"
	}
	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if result == "APPROVED" && item.TargetType == "TASK" && item.TargetID > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE task_instances SET status = 'TODO', audit_note = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND child_id = ?`, "申诉通过："+req.HandleNote, item.TargetID, item.ChildID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE appeals SET status = ?, handle_result = ?, handle_note = ?, handled_at = CURRENT_TIMESTAMP WHERE id = ?`, result, result, req.HandleNote, appealID); err != nil {
		return err
	}
	return tx.Commit()
}

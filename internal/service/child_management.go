package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"family-score/internal/consts"
	"family-score/internal/logger"
	"family-score/internal/protocol"
	"family-score/pkg/utils"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// CreateChild 创建孩子档案并可同步创建孩子登录账号。
func (s *Service) CreateChild(ctx context.Context, sess protocol.Session, req CreateChildParam) (int64, error) {
	if !CanOperate(sess) {
		return 0, errors.New("只有管理员或家长可以添加孩子")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Gender = strings.ToUpper(strings.TrimSpace(req.Gender))
	req.ChildLoginName = strings.TrimSpace(req.ChildLoginName)
	if req.Name == "" || req.Age <= 0 {
		return 0, errors.New("孩子姓名和年龄不能为空")
	}
	if req.Gender == "" {
		req.Gender = "BOY"
	}
	if req.Gender != "BOY" && req.Gender != "GIRL" {
		return 0, errors.New("性别只能是 BOY 或 GIRL")
	}
	parentUserID, parentGroup, err := s.resolveChildBinding(ctx, sess, req.ParentUserID, req.ParentGroup)
	if err != nil {
		return 0, err
	}
	withChildAccount := req.ChildLoginName != "" || req.ChildPassword != ""
	if withChildAccount && (req.ChildLoginName == "" || len(req.ChildPassword) < 4) {
		return 0, errors.New("孩子登录名不能为空，密码至少 4 位")
	}
	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `INSERT INTO children(family_id, parent_user_id, parent_group, name, age, gender, profile_note) VALUES(?, ?, ?, ?, ?, ?, ?)`, sess.FamilyID, nullableParentID(parentUserID), parentGroup, req.Name, req.Age, req.Gender, "家长名下孩子档案")
	if err != nil {
		return 0, err
	}
	childID, _ := res.LastInsertId()
	if _, err := tx.ExecContext(ctx, `INSERT INTO score_accounts(child_id, current_month) VALUES(?, ?)`, childID, utils.CurrentMonth()); err != nil {
		return 0, err
	}
	if withChildAccount {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.ChildPassword), bcrypt.DefaultCost)
		if err != nil {
			return 0, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO users(family_id, child_id, display_name, role, login_name, password_hash, enabled) VALUES(?, ?, ?, ?, ?, ?, 1)`, sess.FamilyID, childID, req.Name, consts.RoleChild, req.ChildLoginName, string(hash))
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				return 0, errors.New("孩子登录名已存在")
			}
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	logger.L().Info("新增孩子档案", zap.Int64("child_id", childID), zap.Int64("parent_user_id", parentUserID), zap.Bool("with_child_account", withChildAccount), zap.Int64("operator_id", sess.UserID))
	return childID, nil
}

// UpdateChild 修改孩子档案和归属家长。
func (s *Service) UpdateChild(ctx context.Context, sess protocol.Session, childID int64, req protocol.UpdateChildParam) (protocol.Child, error) {
	if !CanOperate(sess) {
		return protocol.Child{}, errors.New("只有管理员或家长可以修改孩子")
	}
	if childID <= 0 || !s.CanAccessChild(ctx, sess, childID) {
		return protocol.Child{}, errors.New("无权访问该孩子档案")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Gender = strings.ToUpper(strings.TrimSpace(req.Gender))
	if req.Name == "" || req.Age <= 0 {
		return protocol.Child{}, errors.New("孩子姓名和年龄不能为空")
	}
	if req.Gender == "" {
		req.Gender = "BOY"
	}
	if req.Gender != "BOY" && req.Gender != "GIRL" {
		return protocol.Child{}, errors.New("性别只能是 BOY 或 GIRL")
	}
	parentUserID, parentGroup, err := s.resolveChildBinding(ctx, sess, req.ParentUserID, req.ParentGroup)
	if err != nil {
		return protocol.Child{}, err
	}
	if err := s.ensureGuardianGroup(ctx, sess.FamilyID, parentGroup); err != nil {
		return protocol.Child{}, err
	}
	res, err := s.repo.DB.ExecContext(ctx, `UPDATE children SET name = ?, age = ?, gender = ?, parent_user_id = ?, parent_group = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND family_id = ?`, req.Name, req.Age, req.Gender, nullableParentID(parentUserID), parentGroup, childID, sess.FamilyID)
	if err != nil {
		return protocol.Child{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return protocol.Child{}, errors.New("孩子档案不存在")
	}
	logger.L().Info("修改孩子档案", zap.Int64("child_id", childID), zap.Int64("parent_user_id", parentUserID), zap.String("parent_group", parentGroup), zap.Int64("operator_id", sess.UserID))
	return protocol.Child{ID: childID, FamilyID: sess.FamilyID, ParentUserID: parentUserID, ParentGroup: parentGroup, Name: req.Name, Age: req.Age, Gender: req.Gender}, nil
}

// DeleteChild 删除孩子档案及其关联数据。
func (s *Service) resolveChildBinding(ctx context.Context, sess protocol.Session, requestedParentID int64, requestedGroup string) (int64, string, error) {
	parentUserID := requestedParentID
	parentGroup := strings.TrimSpace(requestedGroup)
	if sess.Role == consts.RoleParent {
		parentUserID = sess.UserID
		ownGroup := s.parentGroupOf(ctx, sess.FamilyID, sess.UserID)
		if parentGroup == "" {
			parentGroup = ownGroup
		}
		if ownGroup != "" && parentGroup != ownGroup {
			return 0, "", errors.New("家长只能绑定到自己的监护组")
		}
	}
	if parentUserID > 0 {
		var group string
		if err := s.repo.DB.QueryRowContext(ctx, `SELECT COALESCE(parent_group, '') FROM users WHERE id = ? AND family_id = ? AND role = 'PARENT' AND enabled = 1`, parentUserID, sess.FamilyID).Scan(&group); err != nil {
			return 0, "", errors.New("归属家长不存在或未启用")
		}
		if parentGroup == "" {
			parentGroup = group
		}
	}
	parentGroup = strings.TrimSpace(parentGroup)
	if len([]rune(parentGroup)) > 32 {
		parentGroup = string([]rune(parentGroup)[:32])
	}
	return parentUserID, parentGroup, nil
}

func nullableParentID(parentUserID int64) sql.NullInt64 {
	return sql.NullInt64{Int64: parentUserID, Valid: parentUserID > 0}
}

func (s *Service) DeleteChild(ctx context.Context, sess protocol.Session, childID int64) error {
	if !CanOperate(sess) {
		return errors.New("只有管理员或家长可以删除孩子")
	}
	if childID <= 0 || !s.CanAccessChild(ctx, sess, childID) {
		return errors.New("无权访问该孩子档案")
	}
	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT id FROM users WHERE child_id = ? AND role = ?`, childID, consts.RoleChild)
	if err != nil {
		return err
	}
	linkedUsers := map[int64]struct{}{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err == nil {
			linkedUsers[id] = struct{}{}
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	deleteSQL := []string{
		`DELETE FROM task_instances WHERE child_id = ?`,
		`DELETE FROM exchange_orders WHERE child_id = ?`,
		`DELETE FROM appeals WHERE child_id = ?`,
		`DELETE FROM weekly_reviews WHERE child_id = ?`,
		`DELETE FROM monthly_settlements WHERE child_id = ?`,
		`DELETE FROM score_records WHERE child_id = ?`,
		`DELETE FROM score_accounts WHERE child_id = ?`,
	}
	for _, stmt := range deleteSQL {
		if _, err := tx.ExecContext(ctx, stmt, childID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET enabled = 0, child_id = NULL, updated_at = CURRENT_TIMESTAMP WHERE child_id = ? AND role = ?`, childID, consts.RoleChild); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM children WHERE id = ? AND family_id = ?`, childID, sess.FamilyID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("孩子档案不存在")
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.clearSessionsByUsersOrChild(linkedUsers, childID)
	logger.L().Info("删除孩子档案", zap.Int64("child_id", childID), zap.Int64("operator_id", sess.UserID))
	return nil
}

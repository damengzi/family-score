package service

import (
	"context"
	"errors"
	"strings"

	"family-score/internal/logger"
	"family-score/internal/protocol"

	"go.uber.org/zap"
)

// GuardianGroups 查询监护组列表。
func (s *Service) GuardianGroups(ctx context.Context, sess protocol.Session) ([]protocol.GuardianGroup, error) {
	if !CanOperate(sess) {
		return nil, errors.New("只有管理员或家长可以查看监护组")
	}
	if err := s.syncGuardianGroups(ctx, sess.FamilyID); err != nil {
		return nil, err
	}
	query := `SELECT g.id, g.family_id, g.name, g.description,
		(SELECT COUNT(*) FROM users u WHERE u.family_id = g.family_id AND u.role = 'PARENT' AND u.enabled = 1 AND u.parent_group = g.name),
		(SELECT COUNT(*) FROM children c WHERE c.family_id = g.family_id AND c.parent_group = g.name),
		g.created_at
		FROM guardian_groups g WHERE g.family_id = ? ORDER BY g.id DESC`
	args := []any{sess.FamilyID}
	if sess.Role != "ADMIN" {
		group := s.parentGroupOf(ctx, sess.FamilyID, sess.UserID)
		query = `SELECT g.id, g.family_id, g.name, g.description,
			(SELECT COUNT(*) FROM users u WHERE u.family_id = g.family_id AND u.role = 'PARENT' AND u.enabled = 1 AND u.parent_group = g.name),
			(SELECT COUNT(*) FROM children c WHERE c.family_id = g.family_id AND c.parent_group = g.name),
			g.created_at
			FROM guardian_groups g WHERE g.family_id = ? AND g.name = ? ORDER BY g.id DESC`
		args = append(args, group)
	}
	rows, err := s.repo.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.GuardianGroup{}
	for rows.Next() {
		var it protocol.GuardianGroup
		if err := rows.Scan(&it.ID, &it.FamilyID, &it.Name, &it.Description, &it.ParentCount, &it.ChildCount, &it.CreatedAt); err == nil {
			items = append(items, it)
		}
	}
	return items, nil
}

// CreateGuardianGroup 创建监护组，仅管理员可用。
func (s *Service) CreateGuardianGroup(ctx context.Context, sess protocol.Session, req protocol.CreateGuardianGroupParam) (int64, error) {
	if !IsAdmin(sess) {
		return 0, errors.New("只有管理员可以创建监护组")
	}
	name := normalizeParentGroup(req.Name)
	description := strings.TrimSpace(req.Description)
	if name == "" {
		return 0, errors.New("监护组名称不能为空")
	}
	res, err := s.repo.DB.ExecContext(ctx, `INSERT INTO guardian_groups(family_id, name, description) VALUES(?, ?, ?)`, sess.FamilyID, name, description)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, errors.New("监护组名称已存在")
		}
		return 0, err
	}
	id, _ := res.LastInsertId()
	logger.L().Info("新增监护组", zap.Int64("group_id", id), zap.String("name", name), zap.Int64("operator_id", sess.UserID))
	return id, nil
}

// UpdateGuardianGroup 修改监护组名称和说明，并同步已绑定的家长与孩子。
func (s *Service) UpdateGuardianGroup(ctx context.Context, sess protocol.Session, id int64, req protocol.UpdateGuardianGroupParam) (protocol.GuardianGroup, error) {
	if !IsAdmin(sess) {
		return protocol.GuardianGroup{}, errors.New("只有管理员可以修改监护组")
	}
	name := normalizeParentGroup(req.Name)
	description := strings.TrimSpace(req.Description)
	if id <= 0 || name == "" {
		return protocol.GuardianGroup{}, errors.New("监护组不存在或名称为空")
	}
	var oldName string
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT name FROM guardian_groups WHERE id = ? AND family_id = ?`, id, sess.FamilyID).Scan(&oldName); err != nil {
		return protocol.GuardianGroup{}, errors.New("监护组不存在")
	}
	tx, err := s.repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return protocol.GuardianGroup{}, err
	}
	defer tx.Rollback()
	if name != oldName {
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM guardian_groups WHERE family_id = ? AND name = ? AND id <> ?`, sess.FamilyID, name, id).Scan(&exists); err != nil {
			return protocol.GuardianGroup{}, err
		}
		if exists > 0 {
			return protocol.GuardianGroup{}, errors.New("监护组名称已存在")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE users SET parent_group = ?, updated_at = CURRENT_TIMESTAMP WHERE family_id = ? AND parent_group = ?`, name, sess.FamilyID, oldName); err != nil {
			return protocol.GuardianGroup{}, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE children SET parent_group = ?, updated_at = CURRENT_TIMESTAMP WHERE family_id = ? AND parent_group = ?`, name, sess.FamilyID, oldName); err != nil {
			return protocol.GuardianGroup{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE guardian_groups SET name = ?, description = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND family_id = ?`, name, description, id, sess.FamilyID); err != nil {
		return protocol.GuardianGroup{}, err
	}
	if err := tx.Commit(); err != nil {
		return protocol.GuardianGroup{}, err
	}
	logger.L().Info("修改监护组", zap.Int64("group_id", id), zap.String("old_name", oldName), zap.String("new_name", name), zap.Int64("operator_id", sess.UserID))
	items, _ := s.GuardianGroups(ctx, sess)
	for _, it := range items {
		if it.ID == id {
			return it, nil
		}
	}
	return protocol.GuardianGroup{ID: id, FamilyID: sess.FamilyID, Name: name, Description: description}, nil
}

// DeleteGuardianGroup 删除未被使用的监护组。
func (s *Service) DeleteGuardianGroup(ctx context.Context, sess protocol.Session, id int64) error {
	if !IsAdmin(sess) {
		return errors.New("只有管理员可以删除监护组")
	}
	var name string
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT name FROM guardian_groups WHERE id = ? AND family_id = ?`, id, sess.FamilyID).Scan(&name); err != nil {
		return errors.New("监护组不存在")
	}
	var used int
	if err := s.repo.DB.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM users WHERE family_id = ? AND parent_group = ?) + (SELECT COUNT(*) FROM children WHERE family_id = ? AND parent_group = ?)`, sess.FamilyID, name, sess.FamilyID, name).Scan(&used); err != nil {
		return err
	}
	if used > 0 {
		return errors.New("监护组仍有家长或孩子绑定，不能删除")
	}
	_, err := s.repo.DB.ExecContext(ctx, `DELETE FROM guardian_groups WHERE id = ? AND family_id = ?`, id, sess.FamilyID)
	return err
}

func (s *Service) ensureGuardianGroup(ctx context.Context, familyID int64, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	_, err := s.repo.DB.ExecContext(ctx, `INSERT OR IGNORE INTO guardian_groups(family_id, name, description) VALUES(?, ?, '')`, familyID, name)
	return err
}

func (s *Service) syncGuardianGroups(ctx context.Context, familyID int64) error {
	stmts := []string{
		`INSERT OR IGNORE INTO guardian_groups(family_id, name, description) SELECT DISTINCT family_id, parent_group, '' FROM users WHERE family_id = ? AND parent_group <> ''`,
		`INSERT OR IGNORE INTO guardian_groups(family_id, name, description) SELECT DISTINCT family_id, parent_group, '' FROM children WHERE family_id = ? AND parent_group <> ''`,
	}
	for _, stmt := range stmts {
		if _, err := s.repo.DB.ExecContext(ctx, stmt, familyID); err != nil {
			return err
		}
	}
	return nil
}
